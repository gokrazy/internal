//go:build ignore
// +build ignore

package main

import (
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
}
