package httpclient

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gokrazy/internal/config"
)

func GetTLSHttpClientByTLSFlag(tlsFlag *string, baseUrl *url.URL) (*http.Client, bool, error) {
	if *tlsFlag == "" {
		return http.DefaultClient, false, nil
	}
	rootCAs, err := x509.SystemCertPool()
	if err != nil {
		log.Printf("initializing x509 system cert pool failed (%v), falling back to empty cert pool", err)
	}
	if rootCAs == nil {
		rootCAs = x509.NewCertPool()
	}

	foundMatchingCertificate := false
	// Append user specified certificate(s)
	if *tlsFlag != "self-signed" {
		usrCert := strings.Split(*tlsFlag, ",")[0]
		certBytes, err := ioutil.ReadFile(usrCert)
		if err != nil {
			return nil, false, fmt.Errorf("reading user specified certificate %s: %v", usrCert, err)
		}
		rootCAs.AppendCertsFromPEM(certBytes)
	} else {
		// Try to find a certificate in the local host config
		hostConfig := config.HostnameSpecific(baseUrl.Host)
		certPath := filepath.Join(string(hostConfig), "cert.pem")
		if _, err := os.Stat(certPath); !os.IsNotExist(err) {
			foundMatchingCertificate = true
			log.Printf("Using certificate %s", certPath)
			certBytes, err := ioutil.ReadFile(certPath)
			if err != nil {
				return nil, false, fmt.Errorf("reading certificate %s: %v", certPath, err)
			}
			rootCAs.AppendCertsFromPEM(certBytes)
		}
	}

	return GetTLSHttpClient(rootCAs), foundMatchingCertificate, nil
}

func GetTLSHttpClient(trustStore *x509.CertPool) *http.Client {
	httpTransport := http.DefaultTransport.(*http.Transport).Clone()
	httpTransport.TLSClientConfig = &tls.Config{
		RootCAs: trustStore,
	}

	return &http.Client{
		Transport: httpTransport,
	}
}

func GetRemoteScheme(baseUrl *url.URL) (string, error) {
	// probe for https redirect, before sending credentials via http
	probeClient := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse // do not follow redirects
		},
	}
	probeResp, err := probeClient.Get("http://" + baseUrl.Host)
	if err != nil {
		return "", fmt.Errorf("probing url for https: %v", err)
	}
	probeLocation, err := probeResp.Location()
	if err != nil {
		return "", fmt.Errorf("getting probe url for https: %v", err)
	}
	return probeLocation.Scheme, nil
}
