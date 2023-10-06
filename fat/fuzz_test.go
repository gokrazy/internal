package fat_test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gokrazy/internal/fat"
)

func FuzzSizes(f *testing.F) {
	f.Fuzz(func(t *testing.T, inp []byte) {
		if len(inp)%4 != 0 {
			return
		}
		nInp := len(inp) / 4
		validationInp := inp
		for cnt := 0; cnt < nInp; cnt++ {
			fileSize := binary.LittleEndian.Uint32(validationInp[:4])
			validationInp = validationInp[4:]
			if fileSize > 1*1024*1024 {
				return // do not generate files over 1 MB
			}
		}

		tmp, err := ioutil.TempFile("", "example")
		if err != nil {
			t.Fatal(err)
		}
		defer os.Remove(tmp.Name())
		fw, err := fat.NewWriter(tmp)
		if err != nil {
			t.Fatal(err)
		}

		for cnt := 0; cnt < nInp; cnt++ {
			fileSize := binary.LittleEndian.Uint32(inp[:4])
			inp = inp[4:]

			log.Printf("writing /%d.txt with %d x bytes", cnt, fileSize)
			w, err := fw.File(fmt.Sprintf("/%d.txt", cnt), time.Now())
			if err != nil {
				t.Fatal(err)
			}
			if _, err := w.Write(bytes.Repeat([]byte("x"), int(fileSize))); err != nil {
				t.Fatal(err)
			}
		}

		if err := fw.Flush(); err != nil {
			t.Fatal(err)
		}

		// dosfsck verifies it can access the entire file system, but our FAT writer
		// might not fill up the entire file system, resulting in a too-short file:
		size, err := tmp.Seek(0, io.SeekCurrent)
		if err != nil {
			t.Fatal(err)
		}
		if pad := fw.TotalSectors*512 - int(size); pad > 0 {
			if _, err := tmp.Write(bytes.Repeat([]byte{0}, pad)); err != nil {
				t.Fatal(err)
			}
		}

		if err := tmp.Close(); err != nil {
			t.Fatal(err)
		}

		cmd := exec.Command("dosfsck", "-v", tmp.Name())
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			cp := exec.Command("cp", tmp.Name(), "/tmp/broken.fat")
			cp.Stdout = os.Stdout
			cp.Stderr = os.Stderr
			cp.Run()
			t.Fatal(err)
		}
	})
}
