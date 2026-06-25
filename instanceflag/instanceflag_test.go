package instanceflag_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gokrazy/internal/config"
	"github.com/gokrazy/internal/instanceflag"
)

func TestFlags(t *testing.T) {
	for _, tt := range []struct {
		name           string
		parent         string
		wantName       string
		wantPath       string
		wantConfigPath string
	}{
		{
			name:           "scan2drive",
			parent:         "/nonexistant/parent",
			wantName:       "scan2drive",
			wantPath:       "/nonexistant/parent/scan2drive",
			wantConfigPath: "/nonexistant/parent/scan2drive/config.json",
		},

		// absolute instance name (e.g. -i $PWD)
		{
			name:           "/nonexistant/other/scan2drive",
			parent:         "/nonexistant/parent",
			wantName:       "scan2drive",
			wantPath:       "/nonexistant/other/scan2drive",
			wantConfigPath: "/nonexistant/other/scan2drive/config.json",
		},

		// relative instance name (e.g. -i ../scan2drive)
		{
			name:           "../hello",
			parent:         "/nonexistant/parent",
			wantName:       "hello",
			wantPath:       "../hello",
			wantConfigPath: "../hello/config.json",
		},
	} {
		f := instanceflag.Flags{
			Name:   tt.name,
			Parent: tt.parent,
		}
		if got, want := f.InstanceName(), tt.wantName; got != want {
			t.Errorf("Flags{Name: %s, Parent: %s}.InstanceName() = %q, want %q", tt.name, tt.parent, got, want)
		}
		if got, want := f.InstancePath(), tt.wantPath; got != want {
			t.Errorf("Flags{Name: %s, Parent: %s}.InstancePath() = %q, want %q", tt.name, tt.parent, got, want)
		}
		if got, want := f.InstanceConfigPath(), tt.wantConfigPath; got != want {
			t.Errorf("Flags{Name: %s, Parent: %s}.InstanceConfigPath() = %q, want %q", tt.name, tt.parent, got, want)
		}
	}
}

func TestFlagsFromPWD(t *testing.T) {
	nonInstanceDir := t.TempDir()
	instanceParentDir := t.TempDir()
	instanceDir := filepath.Join(instanceParentDir, "frompwd")
	instanceConfigJSON := filepath.Join(instanceDir, "config.json")
	cfg := config.NewStruct("frompwd")
	cfg.Packages = []string{"github.com/gokrazy/hello"}
	b, err := cfg.FormatForFile()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(instanceDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(instanceConfigJSON, b, 0644); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		wd             string
		parent         string
		wantName       string
		wantPath       string
		wantConfigPath string
	}{
		{
			wd:             nonInstanceDir,
			parent:         "/nonexistant/parent",
			wantName:       "hello",
			wantPath:       "/nonexistant/parent/hello",
			wantConfigPath: "/nonexistant/parent/hello/config.json",
		},

		{
			wd:             instanceDir,
			parent:         instanceParentDir,
			wantName:       "frompwd",
			wantPath:       instanceDir,
			wantConfigPath: instanceConfigJSON,
		},
	} {
		t.Chdir(tt.wd)
		f := instanceflag.Flags{
			Name:   instanceflag.InstanceDefault(),
			Parent: tt.parent,
		}
		if got, want := f.InstanceName(), tt.wantName; got != want {
			t.Errorf("Flags{Name: %s, Parent: %s}.InstanceName() = %q, want %q", instanceflag.InstanceDefault(), tt.parent, got, want)
		}
		if got, want := f.InstancePath(), tt.wantPath; got != want {
			t.Errorf("Flags{Name: %s, Parent: %s}.InstancePath() = %q, want %q", instanceflag.InstanceDefault(), tt.parent, got, want)
		}
		if got, want := f.InstanceConfigPath(), tt.wantConfigPath; got != want {
			t.Errorf("Flags{Name: %s, Parent: %s}.InstanceConfigPath() = %q, want %q", instanceflag.InstanceDefault(), tt.parent, got, want)
		}
	}
}
