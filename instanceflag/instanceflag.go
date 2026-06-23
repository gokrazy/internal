package instanceflag

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

// ParentDirDefault returns the default value for the --parent_dir flag,
// i.e. $GOKRAZY_PARENT_DIR or os.UserHomeDir/gokrazy.
func ParentDirDefault() string {
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

func instanceDefault(parentDir string) string {
	def := os.Getenv("GOKRAZY_INSTANCE")
	if def == "" {
		def = instanceFromPWD(parentDir)
	}
	if def == "" {
		def = "hello"
	}
	return def
}

func instanceFromPWD(parentDir string) string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	wdAbs, err := filepath.Abs(wd)
	if err != nil {
		return ""
	}

	parentAbs, err := filepath.Abs(parentDir)
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

// InstancePath returns the name of the directory
// containing the gokrazy config.json,
// e.g. /home/michael/gokrazy/scan2drive/config.json
func (f *Flags) InstancePath() string {
	return filepath.Join(f.Parent, f.Name)
}

// InstanceConfigPath returns the full path to config.json.
func (f *Flags) InstanceConfigPath() string {
	return filepath.Join(f.InstancePath(), "config.json")
}

func RegisterPflags(fs *pflag.FlagSet) *Flags {
	parent := ParentDirDefault()
	f := &Flags{
		Parent: parent,
		Name:   instanceDefault(parent),
	}

	fs.StringVarP(&f.Name,
		"instance",
		"i",
		f.Name,
		`instance, identified by hostname`)

	fs.StringVar(&f.Parent,
		"parent_dir",
		f.Parent,
		`parent directory: contains one subdirectory per instance`)

	return f
}
