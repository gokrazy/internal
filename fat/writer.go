package fat

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	sectorSize        = uint16(512)
	sectorsPerCluster = uint8(4)

	clusterSize = int(sectorSize) * int(sectorsPerCluster)

	// unusableSectors is the number of clusters which are always unusable in a
	// FAT: the first two entries have special meaning (copy of the media
	// descriptor and file system state).
	unusableClusters = uint16(2)

	// endOfChain marks the end of a cluster chain in the FAT.
	endOfChain = uint16(0xFFFF)

	// hardDisk is the media descriptor for a hard disk (as opposed to floppy).
	hardDisk = uint8(0xF8)

	// clean describes a cleanly unmounted FAT file system.
	clean = uint16(0xFFFF)
)

type paddingWriter struct {
	w     io.Writer
	count int
	padTo int
}

func (pw *paddingWriter) Write(p []byte) (n int, err error) {
	pw.count += int(len(p))
	return pw.w.Write(p)
}

func (pw *paddingWriter) Flush() error {
	if pw.count%pw.padTo == 0 {
		return nil
	}
	remainder := pw.padTo - (pw.count % pw.padTo)
	pw.count += remainder
	_, err := pw.w.Write(make([]byte, remainder))
	return err
}

type entry interface {
	Name() [8]byte
	Ext() [3]byte
	Attr() uint8
	Size() uint32
	FirstCluster() uint16
	Date() uint16
	Time() uint16
}

type common struct {
	name         string
	ext          string
	modTime      time.Time
	size         uint32
	firstCluster uint16
}

var empty = [8]byte{' ', ' ', ' ', ' ', ' ', ' ', ' ', ' '}

func (c *common) Name() [8]byte {
	var result [8]byte
	copy(result[:], empty[:])
	copy(result[:], []byte(c.name))
	return result
}

func (c *common) Ext() [3]byte {
	var result [3]byte
	copy(result[:], empty[:3])
	copy(result[:], []byte(c.ext))
	return result
}

func (c *common) Size() uint32 {
	return c.size
}

func (c *common) FirstCluster() uint16 {
	return c.firstCluster
}

func (c *common) Time() uint16 {
	return uint16(c.modTime.Hour())<<11 |
		uint16(c.modTime.Minute())<<5 |
		uint16(c.modTime.Second()/2)
}

func (c *common) Date() uint16 {
	return uint16(c.modTime.Year()-1980)<<9 |
		uint16(c.modTime.Month())<<5 |
		uint16(c.modTime.Day())
}

type file struct {
	common
}

func (f *file) Attr() uint8 {
	return 0x1 // read-only
}

type directory struct {
	common
	entries []entry
	byName  map[string]entry
	parent  *directory
}

func (d *directory) Attr() uint8 {
	return 0x10 // directory
}

type Writer struct {
	w io.Writer

	// dataTmp is a temporary file to which all file data will be
	// written. Calling Flush will write the appropriate headers (for
	// which the file data must be known) to the writer, then append
	// dataTmp’s contents.
	dataTmp *os.File

	// fat is a File Allocation Table holding one entry for each
	// sector in the data area, pointing to the FAT entry index of the
	// next sector or (with special value 0xFFFF) marking the end of
	// the file.
	fat []uint16

	root *directory

	pending *fatUpdatingWriter
}

// NewWriter returns a Writer which will write a FAT16B file system
// image to w once Flush is called.
//
// Because the position of the data area in the resulting image
// depends on the size of the file allocation table and number of root
// directory entries, a temporary file is used to store data until
// Flush is called.
func NewWriter(w io.Writer) (*Writer, error) {
	f, err := ioutil.TempFile("", "writefat")
	if err != nil {
		return nil, err
	}

	return &Writer{
		w:       w,
		dataTmp: f,
		root: &directory{
			byName: make(map[string]entry),
		},
	}, nil
}

func (fw *Writer) currentCluster() uint16 {
	return unusableClusters + uint16(len(fw.fat))
}

func (fw *Writer) dir(path string) (*directory, error) {
	cur := fw.root
	for _, component := range strings.Split(path, "/") {
		if component == "" {
			continue
		}
		if _, ok := cur.byName[component]; !ok {
			dir := &directory{
				common: common{
					name: component,
				},
				parent: cur,
				byName: make(map[string]entry),
			}
			cur.entries = append(cur.entries, dir)
			cur.byName[component] = dir
		}
		var ok bool
		cur, ok = cur.byName[component].(*directory)
		if !ok {
			return nil, fmt.Errorf("path %q invalid: component %q identifies a file", path, component)
		}
	}
	return cur, nil
}

// Mkdir creates an empty directory with the given full path,
// e.g. Mkdir("usr/share/lib").
func (fw *Writer) Mkdir(path string, modTime time.Time) error {
	if fw.pending != nil {
		if err := fw.pending.Close(); err != nil {
			return err
		}
		fw.pending = nil
	}
	d, err := fw.dir(path)
	d.common.modTime = modTime.UTC()
	return err
}

type fatUpdatingWriter struct {
	fw    *Writer
	pw    *paddingWriter
	count uint32
	file  *file
}

func (fuw *fatUpdatingWriter) Write(p []byte) (n int, err error) {
	fuw.count += uint32(len(p))
	return fuw.pw.Write(p)
}

func (fuw *fatUpdatingWriter) Close() error {
	if err := fuw.pw.Flush(); err != nil {
		return err
	}
	fw := fuw.fw // for convenience
	for i := 0; i < fuw.pw.count/clusterSize; i++ {
		// Append a pointer to the next FAT entry
		fw.fat = append(fw.fat, fw.currentCluster()+1)
	}
	fw.fat[len(fw.fat)-1] = endOfChain
	if fuw.file != nil {
		fuw.file.size = uint32(fuw.count)
	}
	return nil
}

// File creates a file with the specified path and modTime. The
// returned io.Writer stays valid until the next call to File, Flush
// or Mkdir.
func (fw *Writer) File(path string, modTime time.Time) (io.Writer, error) {
	if fw.pending != nil {
		if err := fw.pending.Close(); err != nil {
			return nil, err
		}
	}
	dir, err := fw.dir(filepath.Dir(path))
	if err != nil {
		return nil, err
	}
	filename := filepath.Base(path)
	parts := strings.Split(filename+".", ".")
	f := &file{
		common: common{
			name:         parts[0],
			ext:          parts[1],
			modTime:      modTime.UTC(),
			firstCluster: fw.currentCluster()}}
	dir.entries = append(dir.entries, f)
	dir.byName[filename] = f
	fw.pending = &fatUpdatingWriter{
		fw: fw,
		pw: &paddingWriter{
			w:     fw.dataTmp,
			padTo: clusterSize,
		},
		file: f,
	}
	return fw.pending, nil
}

func (fw *Writer) writeFAT() error {
	w := &paddingWriter{
		w:     fw.w,
		padTo: int(sectorSize)}

	for _, entry := range append([]uint16{
		(uint16(0xFF) << 8) | uint16(hardDisk), // media descriptor
		clean, // file system state
	}, fw.fat...) {
		if err := binary.Write(w, binary.LittleEndian, entry); err != nil {
			return err
		}
	}

	return w.Flush()
}

func (fw *Writer) writeBootSector(w io.Writer, fatSectors, reservedSectors int) error {
	dataSectors := len(fw.fat) * int(sectorsPerCluster)
	rootDirSectors := 1 // TODO
	totalSectors := reservedSectors + rootDirSectors + fatSectors + dataSectors
	var (
		jumpCode            = [3]byte{0xEB, 0x3C, 0x90}
		OEM                 = [8]byte{'g', 'o', 'k', 'r', 'a', 'z', 'y', '!'}
		volumeLabel         = [11]byte{'g', 'o', 'k', 'r', 'a', 'z', 'y', ' ', ' ', ' ', ' '}
		fileSystemType      = [8]byte{'F', 'A', 'T', '1', '6', ' ', ' ', ' '}
		bootCode            = [448]byte{}
		bootSectorSignature = [2]byte{0x55, 0xAA}
	)
	for _, v := range []interface{}{
		jumpCode,                // jump code: intel 80x86 jump instruction
		OEM,                     // OEM
		sectorSize,              // in bytes
		sectorsPerCluster,       // i.e. each FAT entry covers sectorsPerCluster*sectorSize bytes
		uint16(reservedSectors), // reserved sectors
		uint8(1),                // one copy of the FAT
		uint16(16),              // root directory entries, rounded up to entire blocks (max 16 per block) TODO
		uint16(0),               // 0 = use uint32 number of sectors following later
		hardDisk,                // media descriptor
		uint16(fatSectors),      // number of sectors per FAT
		uint16(32),              // (only for bootcode) number of sectors per track
		uint16(4),               // (only for bootcode) number of heads
		uint32(1),               // no hidden sectors
		uint32(totalSectors),    // total number of sectors
		uint8(0x80),             // (only for bootcode) drive number
		uint8(0),                // (only for bootcode) current head
		uint8(0x29),             // magic value: boot signature
		uint32(0xf3f37b84),      // TODO: volume ID
		volumeLabel,
		fileSystemType,
		bootCode,
		bootSectorSignature,
	} {
		if err := binary.Write(w, binary.LittleEndian, v); err != nil {
			return err
		}
	}
	return nil
}

func (fw *Writer) writeDirEntries(w io.Writer, d *directory) error {
	allEntries := d.entries
	// For non-root directories, add dot and dotdot
	if d.parent != nil {
		allEntries = append([]entry{
			&directory{
				common: common{
					name:         ".",
					firstCluster: fw.currentCluster(),
				},
				parent: d,
			},
			&directory{
				common: common{
					name:         "..",
					firstCluster: d.parent.firstCluster,
				},
				parent: d.parent,
			},
		}, allEntries...)
	}
	for _, entry := range allEntries {
		for _, v := range []interface{}{
			entry.Name(),
			entry.Ext(),
			entry.Attr(),
			[10]byte{}, // reserved
			entry.Time(),
			entry.Date(),
			entry.FirstCluster(),
			entry.Size(), // file size in bytes
		} {
			if err := binary.Write(w, binary.LittleEndian, v); err != nil {
				return err
			}
		}
	}

	return nil
}

func (fw *Writer) writeDir(d *directory) error {
	fuw := &fatUpdatingWriter{
		fw: fw,
		pw: &paddingWriter{
			w:     fw.dataTmp,
			padTo: clusterSize,
		},
	}

	d.firstCluster = fw.currentCluster()
	if err := fw.writeDirEntries(fuw, d); err != nil {
		return err
	}

	if err := fuw.Close(); err != nil {
		return err
	}

	for _, e := range d.entries {
		if e.Attr() != 0x10 {
			continue
		}
		if err := fw.writeDir(e.(*directory)); err != nil {
			return err
		}
	}

	return nil
}

func fullSectors(bytes int) int {
	sectors := bytes / int(sectorSize)
	if bytes%int(sectorSize) > 0 {
		sectors++
	}
	return sectors
}

func fullClusters(bytes int) int {
	clusters := bytes / clusterSize
	if bytes%clusterSize > 0 {
		clusters++
	}
	return clusters
}

// Flush writes the image. The Writer must not be used after calling
// Flush.
func (fw *Writer) Flush() error {
	if fw.pending != nil {
		if err := fw.pending.Close(); err != nil {
			return err
		}
	}

	// Write all non-root directory entries recursively
	for _, e := range fw.root.entries {
		if e.Attr() != 0x10 {
			continue
		}
		if err := fw.writeDir(e.(*directory)); err != nil {
			return err
		}
	}

	// Blow up FAT to at least 4085 entries so that 16-bit FAT values
	// must be used, which is more convenient and the only size of FAT
	// values we support.
	if len(fw.fat) < 4085 {
		pad := make([]uint16, 4085-len(fw.fat))
		fw.fat = append(fw.fat, pad...)
	}

	fatSectors := fullSectors(len(fw.fat) * 2)

	// We only need to reserve the boot sector, but the number of reserved
	// sectors must be aligned to clusters (at least on the Raspberry Pi 3).
	reservedSectors := fullClusters(1*int(sectorSize)) * int(sectorsPerCluster)

	pw := &paddingWriter{w: fw.w, padTo: clusterSize}
	if err := fw.writeBootSector(pw, fatSectors, reservedSectors); err != nil {
		return err
	}
	if err := pw.Flush(); err != nil {
		return err
	}

	if err := fw.writeFAT(); err != nil {
		return err
	}

	// root directory
	pw = &paddingWriter{
		w:     fw.w,
		padTo: int(sectorSize),
	}
	if err := fw.writeDirEntries(pw, fw.root); err != nil {
		return err
	}
	if err := pw.Flush(); err != nil {
		return err
	}

	// data area
	if _, err := fw.dataTmp.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err := io.Copy(fw.w, fw.dataTmp); err != nil {
		return err
	}

	fw.dataTmp.Close()
	return os.Remove(fw.dataTmp.Name())
}
