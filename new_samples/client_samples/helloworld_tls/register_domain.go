package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/uber-common/cadence-samples/new_samples/worker"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

func main() {
	logger := worker.BuildLogger()
	logger.Info("Registering default domain for cadence-vishwa with TLS...")

	withTLSDialOption, err := buildTLSDialOption()
	if err != nil {
		logger.Fatal("Failed to build TLS dial option", zap.Error(err))
	}

	cadenceClient := worker.BuildCadenceClient(withTLSDialOption)

	// Register the domain
	domain := "default"
	retentionDays := int32(7)
	emitMetric := true

	req := &shared.RegisterDomainRequest{
		Name:                                   &domain,
		Description:                            stringPtr("Default domain for cadence samples"),
		WorkflowExecutionRetentionPeriodInDays: &retentionDays,
		EmitMetric:                             &emitMetric,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err = cadenceClient.RegisterDomain(ctx, req)
	if err != nil {
		// Check if domain already exists
		if _, ok := err.(*shared.DomainAlreadyExistsError); ok {
			logger.Info("Domain already exists", zap.String("domain", domain))
			return
		}
		logger.Fatal("Failed to register domain", zap.Error(err))
	}

	logger.Info("Successfully registered domain", zap.String("domain", domain))
}

func buildTLSDialOption() (grpc.DialOption, error) {
	// Load client certificate
	clientCert, err := tls.LoadX509KeyPair("credentials/client.crt", "credentials/client.key")
	if err != nil {
		return nil, fmt.Errorf("failed to load client certificate: %w", err)
	}

	// Load server CA
	caCert, err := os.ReadFile("credentials/keytest.crt")
	if err != nil {
		return nil, fmt.Errorf("failed to load server CA certificate: %w", err)
	}

	caCertPool := x509.NewCertPool()
	if !caCertPool.AppendCertsFromPEM(caCert) {
		return nil, fmt.Errorf("failed to append CA certificate")
	}

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{clientCert},
		MinVersion:         tls.VersionTLS12,
	}

	creds := credentials.NewTLS(tlsConfig)
	return grpc.DialerCredentials(creds), nil
}

func stringPtr(s string) *string {
	return &s
}
