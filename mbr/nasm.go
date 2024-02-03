//go:build ignore
// +build ignore

package main

import (
	"fmt"
	"go/format"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
)

func main() {
	nasm := exec.Command("nasm", "bootloader.asm", "-o", "bootloader.img")
	nasm.Stderr = os.Stderr
	if err := nasm.Run(); err != nil {
		log.Fatalf("%v: %v", nasm.Args, err)
	}
	b, err := os.ReadFile("bootloader.img")
	if err != nil {
		log.Fatal(err)
	}
	b = []byte(fmt.Sprintf("package mbr\nvar mbr = %#v", b))
	b, err = format.Source(b)
	if err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("GENERATED_mbr.go", b, 0644); err != nil {
		log.Fatalf("WriteFile: %v", err)
	}
}
