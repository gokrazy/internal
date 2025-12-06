package updateflag_test

import (
	"testing"

	"github.com/gokrazy/internal/updateflag"
)

func TestBaseURL(t *testing.T) {
	for _, tt := range []struct {
		desc      string
		HTTPPort  string
		HTTPSPort string
		Schema    string
		Hostname  string
		Password  string
		wantURL   string
	}{
		{
			desc:      "default ports",
			HTTPPort:  "80",
			HTTPSPort: "443",
			Schema:    "http",
			Hostname:  "bakery",
			Password:  "secret",
			wantURL:   "http://gokrazy:secret@bakery/",
		},

		{
			desc:      "custom ports (HTTP)",
			HTTPPort:  "81",
			HTTPSPort: "444",
			Schema:    "http",
			Hostname:  "bakery",
			Password:  "secret",
			wantURL:   "http://gokrazy:secret@bakery:81/",
		},

		{
			desc:      "custom ports (HTTPS)",
			HTTPPort:  "81",
			HTTPSPort: "444",
			Schema:    "https",
			Hostname:  "bakery",
			Password:  "secret",
			wantURL:   "https://gokrazy:secret@bakery:444/",
		},
	} {
		t.Run(tt.desc, func(t *testing.T) {
			got, err := updateflag.Value{
				Update: "yes",
			}.BaseURL(tt.HTTPPort, tt.HTTPSPort, tt.Schema, tt.Hostname, tt.Password)
			if err != nil {
				t.Fatal(err)
			}
			if got.String() != tt.wantURL {
				t.Errorf("BaseURL(<TODO>): got %q, want %q", got, tt.wantURL)
			}
		})
	}
}
