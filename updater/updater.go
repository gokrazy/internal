package updater

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"hash/crc32"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

var ErrUpdateHandlerNotImplemented = errors.New("update handler not implemented")

type countingWriter int64

func (cw *countingWriter) Write(p []byte) (n int, err error) {
	*cw += countingWriter(len(p))
	return len(p), nil
}

type Target struct {
	BaseURL    string
	HTTPClient *http.Client

	supports []string
}

func NewTarget(baseURL string, httpClient *http.Client) (*Target, error) {
	supports, err := targetSupports(baseURL, httpClient)
	if err != nil {
		return nil, err
	}
	return &Target{
		BaseURL:    baseURL,
		HTTPClient: httpClient,
		supports:   supports,
	}, nil
}

func (t *Target) Supports(feature string) bool {
	for _, f := range t.supports {
		if f == feature {
			return true
		}
	}
	return false
}

// suffix is one of boot, root, mbr, bootonly.
func (t *Target) StreamTo(suffix string, r io.Reader) error {
	start := time.Now()
	updateHash := t.Supports("updatehash")
	var hash hash.Hash
	if updateHash {
		hash = crc32.NewIEEE()
	} else {
		hash = sha256.New()
	}
	var cw countingWriter
	req, err := http.NewRequest(http.MethodPut, t.BaseURL+"update/"+suffix, io.TeeReader(io.TeeReader(r, hash), &cw))
	if err != nil {
		return err
	}
	if updateHash {
		req.Header.Set("X-Gokrazy-Update-Hash", "crc32")
	}
	resp, err := t.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected HTTP status code: got %d, want %d (body %q)", got, want, string(body))
	}
	remoteHash, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if bytes.HasPrefix(remoteHash, []byte("<!DOCTYPE html>")) {
		return ErrUpdateHandlerNotImplemented
	}
	decoded := make([]byte, hex.DecodedLen(len(remoteHash)))
	n, err := hex.Decode(decoded, remoteHash)
	if err != nil {
		return err
	}
	if got, want := decoded[:n], hash.Sum(nil); !bytes.Equal(got, want) {
		return fmt.Errorf("unexpected checksum: got %x, want %x", got, want)
	}
	duration := time.Since(start)
	// TODO: return this
	log.Printf("%d bytes in %v, i.e. %f MiB/s", int64(cw), duration, float64(cw)/duration.Seconds()/1024/1024)
	return nil
}

func (t *Target) Switch() error {
	resp, err := t.HTTPClient.Post(t.BaseURL+"update/switch", "", nil)
	if err != nil {
		return err
	}
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected HTTP status code: got %d, want %d (body %q)", got, want, string(body))
	}
	return nil
}

func (t *Target) Reboot() error {
	resp, err := t.HTTPClient.Post(t.BaseURL+"reboot", "", nil)
	if err != nil {
		return err
	}
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected HTTP status code: got %d, want %d (body %q)", got, want, string(body))
	}
	return nil
}

func targetSupports(baseURL string, client *http.Client) ([]string, error) {
	resp, err := client.Get(baseURL + "update/features")
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusNotFound {
		// Target device does not support /features handler yet, so no features
		// are supported.
		return nil, nil
	}
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected HTTP status code: got %d, want %d (body %q)", got, want, string(body))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return strings.Split(strings.TrimSpace(string(body)), ","), nil
}
