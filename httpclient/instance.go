package httpclient

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/gokrazy/internal/config"
	"github.com/gokrazy/internal/tlsflag"
	"github.com/gokrazy/internal/updateflag"
)

func For(cfg *config.Struct) (_ *http.Client, foundMatchingCertificate bool, updateBaseURL *url.URL, _ error) {
	schema := "http"
	certPath, _, err := tlsflag.CertificatePathsFor(cfg.Hostname)
	if err != nil {
		return nil, false, nil, err
	}
	if certPath != "" {
		schema = "https"
	}

	update, err := cfg.Update.WithFallbackToHostSpecific(cfg.Hostname)
	if err != nil {
		return nil, false, nil, err
	}

	if update.HTTPPort == "" {
		update.HTTPPort = "80"
	}

	if update.HTTPSPort == "" {
		update.HTTPSPort = "443"
	}

	updateBaseURL, err = updateflag.BaseURL(update.HTTPPort, schema, update.Hostname, update.HTTPPassword)
	if err != nil {
		return nil, false, nil, err
	}

	hc, fmc, err := GetTLSHttpClientByTLSFlag(tlsflag.GetUseTLS(), tlsflag.GetInsecure(), updateBaseURL)
	if err != nil {
		return nil, false, nil, fmt.Errorf("getting http client by tls flag: %v", err)
	}
	return hc, fmc, updateBaseURL, nil
}

func GetHTTPClientForInstance(instance string) (*http.Client, *url.URL, error) {
	_, updateHostname := updateflag.GetUpdateTarget(instance)
	pw, err := config.HostnameSpecific(updateHostname).ReadFile("http-password.txt")
	if err != nil {
		return nil, nil, err
	}

	port, err := config.HostnameSpecific(updateHostname).ReadFile("http-port.txt")
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, err

	}

	cert, err := config.HostnameSpecific(updateHostname).ReadFile("cert.pem")
	if err != nil && !os.IsNotExist(err) {
		return nil, nil, err

	}

	tlsFlag := ""
	scheme := "http"
	if cert != "" {
		tlsFlag = "self-signed"
		scheme = "https"
	}

	if port == "" {
		if scheme == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}

	updateBaseUrl, err := updateflag.BaseURL(port, scheme, updateHostname, pw)
	if err != nil {
		return nil, nil, err
	}

	httpClient, _, err := GetTLSHttpClientByTLSFlag(tlsFlag, false, updateBaseUrl)
	if err != nil {
		return nil, nil, err
	}

	return httpClient, updateBaseUrl, err
}
