package fat

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"time"
)

// Reader is a minimalistic FAT16B reader, which only aims to be
// compatible with file systems created by Writer.
type Reader struct {
	r                 io.ReadSeeker
	sectorSize        uint16
	sectorsPerCluster uint8
	reservedSectors   uint16
	rootDirEntries    uint16
	fatSectors        uint16
}

// NewReader creates a new FAT16B Reader by reading file system
// metadata.
func NewReader(r io.ReadSeeker) (*Reader, error) {
	rd := &Reader{
		r: r,
	}

	// Skip jumpCode and OEM
	if _, err := r.Seek(3+8, io.SeekStart); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &rd.sectorSize); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &rd.sectorsPerCluster); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &rd.reservedSectors); err != nil {
		return nil, err
	}

	// Skip number of FAT copies
	if _, err := r.Seek(1, io.SeekCurrent); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &rd.rootDirEntries); err != nil {
		return nil, err
	}

	// Skip number of sectors and media type
	if _, err := r.Seek(2+1, io.SeekCurrent); err != nil {
		return nil, err
	}

	if err := binary.Read(r, binary.LittleEndian, &rd.fatSectors); err != nil {
		return nil, err
	}

	return rd, nil
}

func (r *Reader) fullSectors(bytes int64) int64 {
	sectorSize := int64(r.sectorSize)
	clusters := bytes / sectorSize
	if bytes%sectorSize > 0 {
		clusters++
	}
	return clusters
}

type dirEntry struct {
	Name         [8]byte
	Ext          [3]byte
	Attr         uint8
	Reserved     [10]byte
	Time         uint16
	Date         uint16
	FirstCluster uint16
	Size         uint32
}

// Extents returns the offset and length of the file identified by path.
//
// This function is useful only on FAT file systems where all files
// are stored un-fragmented, such as file systems generated by Writer.
func (r *Reader) Extents(path string) (offset int64, length int64, err error) {
	dirOffset := int64((r.reservedSectors + r.fatSectors)) * int64(r.sectorSize)
	dataOffset := dirOffset + r.fullSectors(int64(r.rootDirEntries)*32)*int64(r.sectorSize)

	numDirEntries := int(r.rootDirEntries)

	components := strings.Split(path[1:], "/")
	for _, component := range components {
		for i := 0; i < numDirEntries; i++ {
			if _, err := r.r.Seek(dirOffset+int64(i*32), io.SeekStart); err != nil {
				return 0, 0, err
			}

			var entry dirEntry

			if err := binary.Read(r.r, binary.LittleEndian, &entry); err != nil {
				return 0, 0, err
			}

			// unused slot
			if entry.Name[0] == 0 {
				continue
			}

			var name string
			if idx := bytes.IndexByte(entry.Name[:], ' '); idx > -1 {
				name = string(entry.Name[:idx])
			} else {
				name = string(entry.Name[:])
			}
			if entry.Ext[0] != ' ' {
				name += "." + string(entry.Ext[:])
			}

			// TODO: read long file names entries instead (with fallback for older installations)
			primary, ext := shortFileName(component, make(map[string]bool))
			shortName := strings.TrimSpace(primary)
			if ext != "" {
				shortName += "." + strings.TrimSpace(ext)
			}
			if strings.ToLower(name) != strings.ToLower(shortName) &&
				name != component /* for backwards compatibility */ {
				continue
			}

			offset := dataOffset + int64(entry.FirstCluster-2)*int64(r.sectorsPerCluster)*int64(r.sectorSize)
			if entry.Attr == attrDirectory {
				dirOffset = offset
				break
			}

			return offset, int64(entry.Size), nil
		}
	}

	return 0, 0, fmt.Errorf("%q not found", path)
}

func unmarshalTimeDate(t, d uint16) time.Time {
	year := (d >> 9) & 0x7F
	month := (d >> 5) & 0x0F
	day := d & 0x1F

	hour := (t >> 11) & 0x1F
	minute := (t >> 5) & 0x3F
	second := t & 0x1F

	return time.Date(1980+int(year), time.Month(month), int(day), int(hour), int(minute), int(second)*2, 0, time.UTC)
}

// ModTime returns the modification time of the file identified by path.
//
// TODO: implement support for subdirectories
func (r *Reader) ModTime(path string) (time.Time, error) {
	dirOffset := int64((r.reservedSectors + r.fatSectors)) * int64(r.sectorSize)

	numDirEntries := int(r.rootDirEntries)

	components := strings.Split(path[1:], "/")
	for _, component := range components {
		for i := 0; i < numDirEntries; i++ {
			if _, err := r.r.Seek(dirOffset+int64(i*32), io.SeekStart); err != nil {
				return time.Time{}, err
			}

			var entry dirEntry

			if err := binary.Read(r.r, binary.LittleEndian, &entry); err != nil {
				return time.Time{}, err
			}

			// unused slot
			if entry.Name[0] == 0 {
				continue
			}

			var name string
			if idx := bytes.IndexByte(entry.Name[:], ' '); idx > -1 {
				name = string(entry.Name[:idx])
			} else {
				name = string(entry.Name[:])
			}
			if entry.Ext[0] != ' ' {
				name += "." + string(entry.Ext[:])
			}

			// TODO: read long file names entries instead (with fallback for older installations)
			primary, ext := shortFileName(component, make(map[string]bool))
			shortName := strings.TrimSpace(primary)
			if ext != "" {
				shortName += "." + strings.TrimSpace(ext)
			}
			if strings.ToLower(name) != strings.ToLower(shortName) &&
				name != component /* for backwards compatibility */ {
				continue
			}

			return unmarshalTimeDate(entry.Time, entry.Date), nil
		}
	}

	return time.Time{}, fmt.Errorf("%q not found", path)
}
