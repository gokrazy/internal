package httpclient

import (
	"net/http"
	"net/url"
	"os"

	"github.com/gokrazy/internal/config"
	"github.com/gokrazy/internal/updateflag"
)

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
