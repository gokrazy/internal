package tlsflag

import (
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gokrazy/internal/config"
	"github.com/gokrazy/internal/httpclient"
)

var (
	useTLS   string
	insecure bool
)

func RegisterFlags(fs *flag.FlagSet) {
	fs.StringVar(&useTLS,
		"tls",
		"",
		`TLS certificate for the web interface (-tls=<certificate or full chain path>,<private key path>).
Use -tls=self-signed to generate a self-signed RSA4096 certificate using the hostname specified with -hostname. In this case, the certificate and key will be placed in your local config folder (on Linux: ~/.config/gokrazy/<hostname>/).
WARNING: When reusing a hostname, no new certificate will be generated and the stored one will be used.
When updating a running instance, the specified certificate will be used to verify the connection. Otherwise the updater will load the hostname-specific certificate from your local config folder in addition to the system trust store.
You can also create your own certificate-key-pair (e.g. by using https://github.com/FiloSottile/mkcert) and place them into your local config folder.`)

	fs.BoolVar(&insecure, "insecure",
		false,
		"Ignore TLS stripping detection.")
}

func Insecure() bool {
	return insecure
}

type ErrNotYetCreated struct {
	HostConfigPath string
	CertPath       string
	KeyPath        string
}

func (e *ErrNotYetCreated) Error() string {
	return "self-signed certificate not yet created"
}

func CertificatePathsFor(hostname string) (certPath string, keyPath string, _ error) {
	hostConfigPath := config.HostnameSpecific(hostname)
	certPath = filepath.Join(string(hostConfigPath), "cert.pem")
	keyPath = filepath.Join(string(hostConfigPath), "key.pem")
	exist := true
	if _, err := os.Stat(certPath); os.IsNotExist(err) {
		exist = false
	}
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		exist = false
	}

	switch useTLS {
	case "self-signed":
		// If the user set -tls=self-signed, treat non-existing certificates as
		// an error.

		if !exist {
			return "", "", &ErrNotYetCreated{
				HostConfigPath: string(hostConfigPath),
				CertPath:       certPath,
				KeyPath:        keyPath,
			}
		}

		// TODO: Check validity dates of existing certificate

	case "off":
		// User specified -tls=off explicitly.
		return "", "", nil

	case "":
		// If the user did not set -tls, return the cert/key path locations only
		// if they exist.

		if !exist {
			return "", "", nil
		}

	default:
		parts := strings.Split(useTLS, ",")
		certPath = parts[0]
		if len(parts) > 1 {
			keyPath = parts[1]
		} else {
			return "", "", fmt.Errorf("no private key supplied")
		}
		// TODO: Check validity
	}
	return certPath, keyPath, nil
}

func GetTLSHttpClient(updateBaseUrl *url.URL) (*http.Client, bool, error) {
	return httpclient.GetTLSHttpClientByTLSFlag(useTLS, insecure, updateBaseUrl)
}
