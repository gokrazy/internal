// Package config allows tools such as the gokr-packer or breakglass reading
// host-specific configuration, such as the HTTP password.
package config

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	userConfigDir = func() string {
		userConfigDir, err := os.UserConfigDir()
		if err != nil {
			log.Fatalf("https://golang.org/pkg/os/#UserConfigDir failed: %v", err)
		}
		return userConfigDir
	}()

	// Typically ~/.config/gokrazy on Linux
	gokrazyConfigDir = filepath.Join(userConfigDir, "gokrazy")
)

func Gokrazy() string { return gokrazyConfigDir }

type HostnameDir string

func (h HostnameDir) ReadFile(configBaseName string) (string, error) {
	b, err := ioutil.ReadFile(filepath.Join(string(h), configBaseName))
	if err != nil {
		// fall back to global path
		b, err = ioutil.ReadFile(filepath.Join(gokrazyConfigDir, configBaseName))
		if err != nil {
			return "", err
		}
	}
	return strings.TrimSpace(string(b)), nil
}

func HostnameSpecific(hostname string) HostnameDir {
	return HostnameDir(filepath.Join(gokrazyConfigDir, "hosts", hostname))
}
