package fat_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/gokrazy/internal/fat"
)

func TestDosfsck(t *testing.T) {
	tmp, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	fw, err := fat.NewWriter(tmp)
	if err != nil {
		t.Fatal(err)
	}

	w, err := fw.File("/empty.txt", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("nameserver 8.8.8.8")); err != nil {
		t.Fatal(err)
	}

	w, err = fw.File("/etc/resolv.conf", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("nameserver 8.8.8.8")); err != nil {
		t.Fatal(err)
	}

	w, err = fw.File("/EFI/BOOT/bootx64.efi", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	vmlinuz := make([]byte, 10*1024*1024)
	if _, err := w.Write(vmlinuz); err != nil {
		t.Fatal(err)
	}

	w, err = fw.File("/s.txt", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("short file name")); err != nil {
		t.Fatal(err)
	}

	w, err = fw.File("/s.conf", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("short file name with long extension")); err != nil {
		t.Fatal(err)
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
		t.Fatal(err)
	}
}
