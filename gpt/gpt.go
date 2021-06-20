// Package gpt provides a minimal reader for partition tables in GPT (GUID
// partition tables) format, just enough for the rootdev package to match block
// devices to root=PARTUUID= kernel parameters.
package gpt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
)

// PartitionUUIDs returns the ids of all GPT partitions on the disk.  Currently,
// only the first partition id will be returned, as that is all that gokrazy
// currently can ever match.
func PartitionUUIDs(r io.ReaderAt) []string {
	type partitionEntry struct {
		TypeGUID   [16]byte
		GUID       [16]byte
		FirstLBA   uint64
		LastLBA    uint64
		Attributes uint64
		Name       [72]byte
	}

	// 512 bytes MBR
	// 512 bytes GPT header
	// 512 bytes GPT partition entries
	buf := make([]byte, 3*512)
	if _, err := r.ReadAt(buf, 0); err != nil {
		return nil
	}

	var firstPartition partitionEntry
	if err := binary.Read(bytes.NewReader(buf[2*512:]), binary.LittleEndian, &firstPartition); err != nil {
		return nil
	}

	return []string{
		GUIDFromBytes(firstPartition.GUID[:]),
	}
}

// GUIDFromBytes returns the canonical string representation of the specified
// GUID.
func GUIDFromBytes(b []byte) string {
	// See Intel EFI specification, Appendix A: GUID and Time Formats
	// https://www.intel.de/content/dam/doc/product-specification/efi-v1-10-specification.pdf
	var (
		timeLow                 uint32
		timeMid                 uint16
		timeHighAndVersion      uint16
		clockSeqHighAndReserved uint8
		clockSeqLow             uint8
		node                    [6]byte
	)
	timeLow = binary.LittleEndian.Uint32(b[0:4])
	timeMid = binary.LittleEndian.Uint16(b[4:6])
	timeHighAndVersion = binary.LittleEndian.Uint16(b[6:8])
	clockSeqHighAndReserved = b[8]
	clockSeqLow = b[9]
	copy(node[:], b[10:])
	return fmt.Sprintf("%08X-%04X-%04X-%02X%02X-%012X",
		timeLow,
		timeMid,
		timeHighAndVersion,
		clockSeqHighAndReserved,
		clockSeqLow,
		node)
}
