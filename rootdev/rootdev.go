// Package rootdev provides functions to locate the root device from which
// gokrazy was booted.
//
// All functions provided by this package work only once /proc is mounted.
package rootdev

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

var cmdlineFile = "/proc/cmdline" // for testing

// Matches https://github.com/gokrazy/gokrazy#sd-card-contents
const (
	Boot  = 1
	Root2 = 2
	Root3 = 3
	Perm  = 4
)

// The ubd0=/dev/loop0p2 form is used when running on User Mode Linux.
var (
	rootRe = regexp.MustCompile(
		`(?:root|ubd0)=(/dev/(?:mmcblk[01]p|sda|loop0p|nvme0n1p))([23])`)

	uuidRe = regexp.MustCompile(
		`(?:root|ubd0)=(PARTUUID=[0-9a-fA-F]+)-([023]+)`)
)

func findPartUUID(uuid string) (string, error) {
	var dev string
	err := filepath.Walk("/sys/block", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Printf("findPartUUID: %v", err)
			return nil
		}
		if info.Mode()&os.ModeSymlink == 0 {
			return nil
		}
		devname := "/dev/" + filepath.Base(path)
		f, err := os.Open(devname)
		if err != nil {
			log.Printf("findPartUUID: %v", err)
			return nil
		}
		defer f.Close()
		if _, err := f.Seek(440, io.SeekStart); err != nil {
			var se syscall.Errno
			if errors.As(err, &se) && se == syscall.EINVAL {
				// Seek()ing empty loop devices results in EINVAL.
				return nil
			}
			log.Printf("findPartUUID: %v(%T)", err, err.(*os.PathError).Err)
			return nil
		}
		var diskSig struct {
			ID      uint32
			Trailer uint16
		}
		if err := binary.Read(f, binary.LittleEndian, &diskSig); err != nil {
			log.Printf("findPartUUID: %v", err)
			return nil
		}
		if fmt.Sprintf("%08x", diskSig.ID) == uuid && diskSig.Trailer == 0 {
			dev = devname
			// TODO: abort early with sentinel error code
			return nil
		}
		return nil
	})
	if err != nil {
		return "", err
	}
	if dev == "" {
		return "", fmt.Errorf("PARTUUID=%s not found", uuid)
	}
	return dev, nil
}

func findRaw() (dev string, partition string) {
	cmdline, err := ioutil.ReadFile(cmdlineFile)
	if err != nil {
		panic(err)
	}

	matches := uuidRe.FindStringSubmatch(string(cmdline))
	if len(matches) == 3 {
		return matches[1], matches[2]
	}

	matches = rootRe.FindStringSubmatch(string(cmdline))
	if len(matches) != 3 {
		panic(fmt.Sprintf("rootdev.find: kernel command line %q did not match %v",
			strings.TrimSpace(string(cmdline)),
			rootRe))
	}
	return matches[1], matches[2]
}

func findDev() (dev string, partition string) {
	dev, partition = findRaw()
	if strings.HasPrefix(dev, "PARTUUID=") {
		var err error
		dev, err = findPartUUID(strings.TrimPrefix(dev, "PARTUUID="))
		if err != nil {
			panic(err)
		}
		return dev, strings.TrimPrefix(partition, "0")
	}
	return dev, partition
}

// BlockDevice returns the file system path identifying the block device from
// which gokrazy was booted.
func BlockDevice() string {
	dev, _ := findDev()
	return strings.TrimSuffix(dev, "p")
}

// ActiveRootPartition returns 2 or 3, depending on which partition is the
// currently active root file system.
func ActiveRootPartition() int {
	_, partition := findDev()
	switch partition {
	case "2":
		return 2
	case "3":
		return 3
	default:
		panic(fmt.Sprintf("root partition %q is unexpectedly neither 2 nor 3", partition))
	}
}

// InactiveRootPartition returns 2 or 3, depending on which partition is
// currently inactive (i.e. the target for updates).
func InactiveRootPartition() int {
	switch ActiveRootPartition() {
	case 2:
		return 3
	case 3:
		return 2
	default:
		panic(fmt.Sprintf("root partition %d is unexpectedly neither 2 nor 3", ActiveRootPartition()))
	}
}

// Partition returns the file system path identifying the specified partition on
// the root device from which gokrazy was booted.
//
// E.g. Partition(2) = /dev/mmcblk0p2
func Partition(number int) string {
	dev, _ := findDev()
	if (strings.HasPrefix(dev, "/dev/mmcblk") ||
		strings.HasPrefix(dev, "/dev/loop") ||
		strings.HasPrefix(dev, "/dev/nvme")) &&
		!strings.HasSuffix(dev, "p") {
		dev += "p"
	}
	return dev + strconv.Itoa(number)
}

// PartitionCmdline returns the cmdline identifier (e.g. PARTUUID=aabbccdd-02,
// or /dev/mmcblk0p2) identifying the specified partition on the root device
// from which gokrazy was booted.
//
// E.g. PartitionCmdline(2) = PARTUUID=aabbccdd-02
func PartitionCmdline(number int) string {
	dev, _ := findRaw()
	if (strings.HasPrefix(dev, "/dev/mmcblk") ||
		strings.HasPrefix(dev, "/dev/loop") ||
		strings.HasPrefix(dev, "/dev/nvme")) &&
		!strings.HasSuffix(dev, "p") {
		dev += "p"
	}
	if strings.HasPrefix(dev, "PARTUUID=") {
		return dev + fmt.Sprintf("-%02d", number)
	}
	return dev + strconv.Itoa(number)
}

// PARTUUID returns the partition UUID of the block device from which gokrazy
// was booted, if any (or the empty string).
func PARTUUID() string {
	dev, _ := findRaw()
	if !strings.HasPrefix(dev, "PARTUUID=") {
		return ""
	}
	return strings.TrimPrefix(dev, "PARTUUID=")
}
