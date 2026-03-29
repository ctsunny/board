package services

import (
	"crypto/tls"
	"fmt"

	"golang.org/x/crypto/acme/autocert"
)

// SetupTLS configures Let's Encrypt autocert and returns a tls.Config.
func SetupTLS(domain, certDir string) (*tls.Config, error) {
	if domain == "" {
		return nil, fmt.Errorf("domain must not be empty")
	}
	m := &autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(domain),
		Cache:      autocert.DirCache(certDir),
	}
	tlsCfg := m.TLSConfig()
	return tlsCfg, nil
}
