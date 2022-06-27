package cmd

import (
	"crypto/ed25519"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
)

// GenerateServerCertEd will generate a new certificate and private key. This
// function will return an error if a key or a cert already exist as to avoid
// any data loss.
func GenerateServerCertEd(c *Conf, log zerolog.Logger, local bool) error {
	var caCertPath, caKeyPath = "certs/ca-cert.pem", "certs/ca-key.pem"

	// Create directories if needed
	if err := os.MkdirAll(filepath.Dir(c.Server.TLS.CertPath), os.ModePerm); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(c.Server.TLS.KeyPath), os.ModePerm); err != nil {
		return fmt.Errorf("create directory: %w", err)
	}

	// Error if cert or pk are already present
	if _, err := os.Stat(c.Server.TLS.CertPath); err == nil {
		return fmt.Errorf("file already exists: %s", c.Server.TLS.CertPath)
	}

	if _, err := os.Stat(c.Server.TLS.KeyPath); err == nil {
		return fmt.Errorf("file already exists: %s", c.Server.TLS.KeyPath)
	}

	if _, err := os.Stat(caCertPath); err == nil {
		return fmt.Errorf("file already exists: %s", caCertPath)
	}

	if _, err := os.Stat(caKeyPath); err == nil {
		return fmt.Errorf("file already exists: %s", caKeyPath)
	}

	// Dummy certificate authority
	caCert := &x509.Certificate{
		SerialNumber: big.NewInt(2022),
		Subject: pkix.Name{
			Organization: []string{"Capybara Ltd."},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// Generate a new private key
	caPub, caPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate key: %w", err)
	}

	log.Info().Msg("generated CA ed25519 private key")

	// Generate a new x509 cert with the previously created CA and private key
	caBytes, err := x509.CreateCertificate(rand.Reader, caCert, caCert, caPub, caPriv)
	if err != nil {
		return fmt.Errorf("generate certificate: %w", err)
	}

	log.Info().Msg("generated CA certificate with ed25519 private key")

	// Write the cert to the configured path
	fdca, err := os.OpenFile(caCertPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create file: %s: %w", caCertPath, err)
	}
	defer fdca.Close()

	if err = pem.Encode(fdca, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	}); err != nil {
		return fmt.Errorf("pem encode cert: %w", err)
	}

	log.Info().Str("file", c.Server.TLS.CertPath).Msg("wrote certificate to file")

	// Write the private key to the configured path
	fdk, err := os.OpenFile(caKeyPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create file: %s: %w", caKeyPath, err)
	}
	defer fdk.Close()

	b, err := x509.MarshalPKCS8PrivateKey(caPriv)
	if err != nil {
		return fmt.Errorf("marshal private key: %w", err)
	}

	if err = pem.Encode(fdk, &pem.Block{Type: "PRIVATE KEY", Bytes: b}); err != nil {
		return fmt.Errorf("write file: %s: %w", caKeyPath, err)
	}

	log.Info().Str("file", caKeyPath).Msg("wrote private key to file")

	// Generate server key and CSR
	servPub, servPriv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return fmt.Errorf("generate ed25519 key: %w", err)
	}

	log.Info().Msg("generated server ed25519 private key")

	template := x509.CertificateRequest{
		Subject: pkix.Name{
			CommonName: "Capybara Server",
		},
		SignatureAlgorithm: x509.PureEd25519,
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, servPriv)
	if err != nil {
		return fmt.Errorf("generate csr: %w", err)
	}

	csr, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return fmt.Errorf("parse csr: %w", err)
	}

	log.Info().Msg("generated server csr")

	clientCRTTemplate := x509.Certificate{
		Signature:          csr.Signature,
		SignatureAlgorithm: csr.SignatureAlgorithm,

		PublicKeyAlgorithm: csr.PublicKeyAlgorithm,
		PublicKey:          csr.PublicKey,

		SerialNumber: big.NewInt(2),
		Issuer:       caCert.Subject,
		Subject:      csr.Subject,
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage: []x509.ExtKeyUsage{
			x509.ExtKeyUsageClientAuth,
			x509.ExtKeyUsageServerAuth,
		},
	}

	if local {
		clientCRTTemplate.IPAddresses = []net.IP{net.ParseIP("127.0.0.1")}
	}

	// Sign CSR with CA private key
	clientCRTRaw, err := x509.CreateCertificate(rand.Reader, &clientCRTTemplate, caCert, servPub, caPriv)
	if err != nil {
		return fmt.Errorf("create crt: %w", err)
	}

	log.Info().Msg("signed csr with CA ed25519 private key")

	// Write cert on disk
	fdcrt, err := os.OpenFile(c.Server.TLS.CertPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create file: %s: %w", c.Server.TLS.CertPath, err)
	}
	defer fdcrt.Close()

	err = pem.Encode(fdcrt, &pem.Block{Type: "CERTIFICATE", Bytes: clientCRTRaw})
	if err != nil {
		return fmt.Errorf("pem encode crt: %w", err)
	}

	log.Info().Str("file", c.Server.TLS.CertPath).Msg("wrote server certificate to file")

	// Write private key on disk
	fdks, err := os.OpenFile(c.Server.TLS.KeyPath, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0600)
	if err != nil {
		return fmt.Errorf("create file: %s: %w", c.Server.TLS.KeyPath, err)
	}
	defer fdks.Close()

	b, err = x509.MarshalPKCS8PrivateKey(servPriv)
	if err != nil {
		return fmt.Errorf("marshal server private key: %w", err)
	}

	if err = pem.Encode(fdks, &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: b,
	}); err != nil {
		return fmt.Errorf("pem encode key: %w", err)
	}

	log.Info().Str("file", c.Server.TLS.KeyPath).Msg("wrote server ed25519 private key to file")

	return nil
}
