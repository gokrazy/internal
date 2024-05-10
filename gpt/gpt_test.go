package gpt

import (
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPartitionUUIDs(t *testing.T) {
	f, err := os.Open("testdata/snapshot.gpt.bin")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	got := PartitionUUIDs(f)
	want := []string{
		"80687DB2-F3F9-427A-8199-165DB4B50001",
		"80687DB2-F3F9-427A-8199-165DB4B50002",
		"80687DB2-F3F9-427A-8199-165DB4B50003",
		"80687DB2-F3F9-427A-8199-165DB4B50004",
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Fatalf("unexpected partition UUIDs: diff (-want +got):\n%s", diff)
	}
}

func TestGUIDFromBytes(t *testing.T) {
	b := [16]byte{
		162, 160, 208, 235, 229, 185, 51, 68, 135, 192, 104, 182, 183, 38, 153, 199,
	}
	got := GUIDFromBytes(b[:])
	const want = "EBD0A0A2-B9E5-4433-87C0-68B6B72699C7"
	if got != want {
		t.Errorf("GUIDFromBytes(%x) = %q, want %q", b, got, want)
	}
}
