package fat

import (
	"testing"
	"time"
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
