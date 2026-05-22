package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"go.uber.org/yarpc/transport/grpc"
	"google.golang.org/grpc/credentials"
)

// BuildTLSDialOption creates a gRPC dial option with TLS credentials for secure
// connection to a Cadence server with mutual TLS enabled.
func BuildTLSDialOption() (grpc.DialOption, error) {
	// Load client certificate for mutual TLS
	clientCert, err := tls.LoadX509KeyPair("credentials/client.crt", "credentials/client.key")
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	// Load server CA certificate
	caCert, err := os.ReadFile("credentials/keytest.crt")
	if err != nil {
		return nil, fmt.Errorf("failed to load server CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, // For development only - verify certs in production
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{clientCert},
		MinVersion:         tls.VersionTLS12,
	}

	creds := credentials.NewTLS(tlsConfig)
	return grpc.DialerCredentials(creds), nil
}
