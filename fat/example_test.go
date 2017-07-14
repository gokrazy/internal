package fat_test

import (
	"io/ioutil"
	"log"
	"time"

	"github.com/gokrazy/internal/fat"
)

func Example() {
	tmp, err := ioutil.TempFile("", "example")
	if err != nil {
		log.Fatal(err)
	}

	fw, err := fat.NewWriter(tmp)
	if err != nil {
		log.Fatal(err)
	}

	w, err := fw.File("etc/resolv.conf", time.Now())
	if err != nil {
		log.Fatal(err)
	}
	if _, err := w.Write([]byte("nameserver 8.8.8.8")); err != nil {
		log.Fatal(err)
	}

	if err := fw.Flush(); err != nil {
		log.Fatal(err)
	}

	if err := tmp.Close(); err != nil {
		log.Fatal(err)
	}

	log.Printf("mount -o loop %s /mnt/loop", tmp.Name())
}
