package updater

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
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

func StreamTo(baseUrl string, r io.Reader, client *http.Client) error {
	start := time.Now()
	hash := sha256.New()
	var cw countingWriter
	req, err := http.NewRequest(http.MethodPut, baseUrl, io.TeeReader(io.TeeReader(r, hash), &cw))
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
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
		return fmt.Errorf("unexpected SHA256 hash: got %x, want %x", got, want)
	}
	duration := time.Since(start)
	// TODO: return this
	log.Printf("%d bytes in %v, i.e. %f MiB/s", int64(cw), duration, float64(cw)/duration.Seconds()/1024/1024)
	return nil
}

func UpdateRoot(baseUrl string, r io.Reader, client *http.Client) error {
	return StreamTo(baseUrl+"update/root", r, client)
}

func UpdateBoot(baseUrl string, r io.Reader, client *http.Client) error {
	return StreamTo(baseUrl+"update/boot", r, client)
}

func UpdateMBR(baseUrl string, r io.Reader, client *http.Client) error {
	return StreamTo(baseUrl+"update/mbr", r, client)
}

func Switch(baseUrl string, client *http.Client) error {
	resp, err := client.Post(baseUrl+"update/switch", "", nil)
	if err != nil {
		return err
	}
	if got, want := resp.StatusCode, http.StatusOK; got != want {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("unexpected HTTP status code: got %d, want %d (body %q)", got, want, string(body))
	}
	return nil
}

func Reboot(baseUrl string, client *http.Client) error {
	resp, err := client.Post(baseUrl+"reboot", "", nil)
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

func TargetSupports(baseUrl, feature string, client *http.Client) (bool, error) {
	supports, err := targetSupports(baseUrl, client)
	if err != nil {
		return false, err
	}
	for _, f := range supports {
		if f == feature {
			return true, nil
		}
	}
	return false, nil
}
