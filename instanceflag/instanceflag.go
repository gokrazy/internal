package instanceflag

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

var (
	instance = func() string {
		def := os.Getenv("GOKRAZY_INSTANCE")
		if def == "" {
			def = "hello"
		}
		return def
	}()
	parentDir = func() string {
		def := os.Getenv("GOKRAZY_PARENT_DIR")
		if def == "" {
			homeDir, err := os.UserHomeDir()
			if err != nil {
				homeDir = fmt.Sprintf("os.UserHomeDir failed: %v", err)
			}
			def = filepath.Join(homeDir, "gokrazy")
		}
		return def
	}()
)

func RegisterPflags(fs *pflag.FlagSet) {
	fs.StringVarP(&instance,
		"instance",
		"i",
		instance,
		`instance, identified by hostname`)

	fs.StringVar(&parentDir,
		"parent_dir",
		parentDir,
		`parent directory: contains one subdirectory per instance`)

}

func SetInstance(i string) {
	instance = i
}

func SetParentDir(p string) {
	parentDir = p
}

func Instance() string {
	return instance
}

func ParentDir() string {
	return parentDir
}