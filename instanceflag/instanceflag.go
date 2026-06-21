package instanceflag

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/pflag"
)

func parentDirDefault() string {
	def := os.Getenv("GOKRAZY_PARENT_DIR")
	if def == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			homeDir = fmt.Sprintf("os.UserHomeDir failed: %v", err)
		}
		def = filepath.Join(homeDir, "gokrazy")
	}
	return def
}

func instanceDefault() string {
	def := os.Getenv("GOKRAZY_INSTANCE")
	if def == "" {
		def = instanceFromPWD()
	}
	if def == "" {
		def = "hello"
	}
	return def
}

func instanceFromPWD() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	wdAbs, err := filepath.Abs(wd)
	if err != nil {
		return ""
	}

	parentAbs, err := filepath.Abs(global.Parent)
	if err != nil {
		return ""
	}

	if !strings.HasPrefix(wdAbs, parentAbs+"/") {
		return ""
	}

	// Process is running in an instance directory (a
	// subdirectory of the parent dir), so default the instance
	// flag to that same subdirectory.
	instance := strings.TrimPrefix(wdAbs, parentAbs+"/")
	if idx := strings.IndexRune(instance, '/'); idx > -1 {
		instance = instance[:idx]
	}
	return instance
}

// Flags contains command-line flag values related to the gokrazy instance.
type Flags struct {
	Name   string // --instance / -i
	Parent string // --parent_dir
}

var global Flags

func init() {
	global.Parent = parentDirDefault()
	global.Name = instanceDefault()
}

func RegisterPflags(fs *pflag.FlagSet) *Flags {
	fs.StringVarP(&global.Name,
		"instance",
		"i",
		instanceDefault(),
		`instance, identified by hostname`)

	fs.StringVar(&global.Parent,
		"parent_dir",
		parentDirDefault(),
		`parent directory: contains one subdirectory per instance`)

	return &global
}

func SetInstance(i string) {
	global.Name = i
}

func SetParentDir(p string) {
	global.Parent = p
}

func Instance() string {
	return global.Name
}

var parentDirOnce sync.Once

func ParentDir() string {
	parentDirOnce.Do(func() {
		if !strings.Contains(global.Parent, "./") &&
			!strings.Contains(global.Parent, "../") &&
			!strings.Contains(global.Parent, "/..") {
			return
		}
		if abs, err := filepath.Abs(global.Parent); err == nil {
			global.Parent = abs
		}
	})
	return global.Parent
}
