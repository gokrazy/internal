// +build ignore

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func main() {
	var buf bytes.Buffer
	nasm := exec.Command("nasm", "bootloader.asm", "-o", "/dev/stdout")
	nasm.Stdout = &buf
	nasm.Stderr = os.Stderr
	if err := nasm.Run(); err != nil {
		log.Fatalf("%v: %v", nasm.Args, err)
	}
	if err := ioutil.WriteFile("GENERATED_mbr.go", []byte(fmt.Sprintf("package mbr\nvar mbr = %#v", buf.Bytes())), 0644); err != nil {
		log.Fatalf("WriteFile: %v", err)
	}
}
