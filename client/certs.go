package capybara

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"

	"google.golang.org/grpc/credentials"
)

// loadTLSCredentials will load the CA who signed the server's certificate
// and verify its authenticity.
func loadTLSCredentials(path string) (credentials.TransportCredentials, error) {
	// Load the CA certificate that signed the server's cert
	pemServerCA, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	// Append the CA cert to a new cert pool
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(pemServerCA) {
		return nil, fmt.Errorf("failed to add server CA's certificate")
	}

	// Create and return TLS credentials from the cert pool
	return credentials.NewTLS(&tls.Config{RootCAs: cp}), nil
}
