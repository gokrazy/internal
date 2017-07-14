package fat_test

import (
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

	w, err := fw.File("etc/resolv.conf", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("nameserver 8.8.8.8")); err != nil {
		t.Fatal(err)
	}

	if err := fw.Flush(); err != nil {
		t.Fatal(err)
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
