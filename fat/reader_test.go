package fat

import (
	"io"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestUnmarshalTimeDate(t *testing.T) {
	t.Parallel()

	arbitrary := time.Date(2017, 9, 6, 8, 13, 28, 0, time.UTC)
	arbitraryC := common{modTime: arbitrary}

	for _, entry := range []struct {
		t, d uint16
		want time.Time
	}{
		{
			t:    arbitraryC.Time(),
			d:    arbitraryC.Date(),
			want: arbitrary,
		},
		{
			d:    0x2B14,
			want: time.Date(2001, 8, 20, 0, 0, 0, 0, time.UTC),
		},
		{
			t:    0x5401,
			d:    0x0021, // minimum date
			want: time.Date(1980, 1, 1, 10, 32, 2, 0, time.UTC),
		},
		{
			t:    0x5401,
			d:    0xFC46, // maximum date
			want: time.Date(2106, 2, 6, 10, 32, 2, 0, time.UTC),
		},
	} {
		entry := entry // copy
		t.Run(entry.want.String(), func(t *testing.T) {
			t.Parallel()
			got := unmarshalTimeDate(entry.t, entry.d)
			if !got.Equal(entry.want) {
				t.Fatalf("unexpected time: got %v, want %v", got, entry.want)
			}
		})
	}
}

func TestExtents(t *testing.T) {
	t.Parallel()

	tmp, err := ioutil.TempFile("", "example")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	fw, err := NewWriter(tmp)
	if err != nil {
		t.Fatal(err)
	}

	w, err := fw.File("/resolv.conf", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write([]byte("nameserver 8.8.8.8")); err != nil {
		t.Fatal(err)
	}

	bCmdline := []byte("root=/dev/xda")
	w, err = fw.File("/cmdline.txt", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(bCmdline); err != nil {
		t.Fatal(err)
	}

	bEntry := []byte("options root=/dev/xda")
	w, err = fw.File("/loader/entries/gokrazy.conf", time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.Write(bEntry); err != nil {
		t.Fatal(err)
	}

	if err := fw.Flush(); err != nil {
		t.Fatal(err)
	}

	rd, err := NewReader(tmp)
	if err != nil {
		t.Fatal(err)
	}

	{
		offset, length, err := rd.Extents("/cmdline.txt")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tmp.Seek(offset, io.SeekStart); err != nil {
			t.Fatal(err)
		}
		got := make([]byte, length)
		if _, err := io.ReadFull(tmp, got); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(bCmdline, got); diff != "" {
			t.Fatalf("unexpected cmdline.txt contents: diff (-want +got):\n%s", diff)
		}
	}

	{
		offset, length, err := rd.Extents("/loader/entries/gokrazy.conf")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tmp.Seek(offset, io.SeekStart); err != nil {
			t.Fatal(err)
		}
		got := make([]byte, length)
		if _, err := io.ReadFull(tmp, got); err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(bEntry, got); diff != "" {
			t.Fatalf("unexpected gokrazy.conf contents: diff (-want +got):\n%s", diff)
		}
	}
}
