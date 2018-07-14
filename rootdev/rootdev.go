// Package rootdev provides functions to locate the root device from which
// gokrazy was booted.
package rootdev

import (
	"fmt"
	"io/ioutil"
	"regexp"
)

var rootDeviceRe = regexp.MustCompile(`root=(/dev/(?:mmcblk0p|sda))`)

// MustFind returns the device from which gokrazy was booted. It is safe to
// append a partition number to the resulting string. MustFind works once /proc
// is mounted.
func MustFind() string {
	cmdline, err := ioutil.ReadFile("/proc/cmdline")
	if err != nil {
		panic(err)
	}

	matches := rootDeviceRe.FindStringSubmatch(string(cmdline))
	if len(matches) != 2 {
		panic(fmt.Sprintf("mustFindRootDevice: kernel command line %q did not match %v", string(cmdline), rootDeviceRe))
	}
	return matches[1]
}
