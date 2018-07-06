// Package mbr provides a configured version of Sebastian Plotz’s minimal
// stage1-only Linux bootloader, to be written to a Master Boot Record.
package mbr

import (
	"bytes"
	"encoding/binary"
)

//go:generate go run nasm.go

// write to byte offset 433
type bootloaderParams struct {
	CurrentLBA uint32 // e.g. 8218
	CmdlineLBA uint32 // e.g. 8218
}

func Configure(vmlinuzLba, cmdlineLba uint32) [446]byte {
	buf := bytes.NewBuffer(make([]byte, 0, 446))
	// buf.Write never fails
	buf.Write(mbr[:433])
	params := bootloaderParams{
		CurrentLBA: vmlinuzLba,
		CmdlineLBA: cmdlineLba,
	}
	binary.Write(buf, binary.LittleEndian, &params)
	var b [446]byte
	copy(b[:], buf.Bytes())
	return b
}
