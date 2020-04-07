package rootdev

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
)

func setCmdline(t *testing.T, part string) {
	t.Helper()
	f, err := ioutil.TempFile("", "rootdevtest")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	t.Cleanup(func() { os.Remove(f.Name()) })
	if _, err := fmt.Fprintf(f, "console=tty1 %s ro\n", part); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	cmdlineFile = f.Name()
	// TODO: call find()
}

func TestBlockDevice(t *testing.T) {
	for _, tt := range []struct {
		cmdline string
		want    string
	}{
		{"root=/dev/mmcblk0p2", "/dev/mmcblk0"},
		{"root=/dev/sda2", "/dev/sda"},
		{"ubd0=/dev/loop0p3", "/dev/loop0"},
		// {"root=PARTUUID=2e18c40c-02", "/dev/sda"},
	} {
		t.Run(tt.cmdline, func(t *testing.T) {
			setCmdline(t, tt.cmdline)
			if got, want := BlockDevice(), tt.want; got != want {
				t.Errorf("MustFind() = %v, want %v", got, want)
			}
		})
	}
}

func TestActiveRootPartition(t *testing.T) {
	for _, tt := range []struct {
		cmdline string
		want    int
	}{
		{"root=/dev/mmcblk0p2", 2},
		{"root=/dev/sda2", 2},
		{"ubd0=/dev/loop0p3", 3},
		// {"root=PARTUUID=2e18c40c-02", 2},
	} {
		t.Run(tt.cmdline, func(t *testing.T) {
			setCmdline(t, tt.cmdline)
			if got, want := ActiveRootPartition(), tt.want; got != want {
				t.Errorf("ActiveRootPartition() = %v, want %v", got, want)
			}
		})
	}
}

func TestInactiveRootPartition(t *testing.T) {
	setCmdline(t, "root=/dev/mmcblk0p2")
	if got, want := InactiveRootPartition(), 3; got != want {
		t.Errorf("InactiveRootPartition() = %v, want %v", got, want)
	}
}

func TestPartition(t *testing.T) {
	for _, tt := range []struct {
		cmdline string
		want    string
	}{
		{"root=/dev/mmcblk0p2", "/dev/mmcblk0p3"},
		{"root=/dev/sda2", "/dev/sda3"},
		{"ubd0=/dev/loop0p3", "/dev/loop0p3"},
		// {"root=PARTUUID=2e18c40c-02", "/dev/sda3"},
	} {
		t.Run(tt.cmdline, func(t *testing.T) {
			setCmdline(t, tt.cmdline)
			const partNum = 3
			if got, want := Partition(partNum), tt.want; got != want {
				t.Errorf("Partition(%d) = %v, want %v", partNum, got, want)
			}
		})
	}
}

func TestPartitionCmdline(t *testing.T) {
	for _, tt := range []struct {
		cmdline string
		want    string
	}{
		{"root=/dev/mmcblk0p2", "/dev/mmcblk0p3"},
		{"root=/dev/sda2", "/dev/sda3"},
		{"ubd0=/dev/loop0p3", "/dev/loop0p3"},
		{"root=PARTUUID=2e18c40c-02", "PARTUUID=2e18c40c-03"},
	} {
		t.Run(tt.cmdline, func(t *testing.T) {
			setCmdline(t, tt.cmdline)
			const partNum = 3
			if got, want := PartitionCmdline(partNum), tt.want; got != want {
				t.Errorf("PartitionCmdline(%d) = %v, want %v", partNum, got, want)
			}
		})
	}
}
