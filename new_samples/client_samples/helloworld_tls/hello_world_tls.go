package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/uber-common/cadence-samples/new_samples/worker"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"
)

func main() {
	withTLSDialOption, err := withTLSDialOption()
	if err != nil {
		panic(err)
	}

	cadenceClient := worker.BuildCadenceClient(withTLSDialOption)
	logger := worker.BuildLogger()

	domain := "default"
	tasklist := "default-tasklist"
	workflowID := uuid.New().String()
	requestID := uuid.New().String()
	executionTimeout := int32(60)
	closeTimeout := int32(60)

	workflowType := "cadence_samples.HelloWorldWorkflow"
	input := []byte(`{"message": "Uber"}`)

	req := shared.StartWorkflowExecutionRequest{
		Domain:     &domain,
		WorkflowId: &workflowID,
		WorkflowType: &shared.WorkflowType{
			Name: &workflowType,
		},
		TaskList: &shared.TaskList{
			Name: &tasklist,
		},
		Input:                               input,
		ExecutionStartToCloseTimeoutSeconds: &executionTimeout,
		TaskStartToCloseTimeoutSeconds:      &closeTimeout,
		RequestId:                           &requestID,
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	resp, err := cadenceClient.StartWorkflowExecution(ctx, &req)
	if err != nil {
		logger.Error("Failed to create workflow", zap.Error(err))
		panic("Failed to create workflow.")
	}

	logger.Info("successfully started HelloWorld workflow", zap.String("runID", resp.GetRunId()))
}

func withTLSDialOption() (grpc.DialOption, error) {
	// Present client cert for mutual TLS (if enabled on server)
	clientCert, err := tls.LoadX509KeyPair("credentials/client.crt", "credentials/client.key")
	if err != nil {
		return nil, fmt.Errorf("Failed to load client certificate: %v", zap.Error(err))
	}

	// Load server CA
	caCert, err := os.ReadFile("credentials/keytest.crt")
	if err != nil {
		return nil, fmt.Errorf("Failed to load server CA certificate: %v", zap.Error(err))
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	tlsConfig := tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{clientCert},
	}
	// Create TLS credentials from the TLS configuration
	creds := credentials.NewTLS(&tlsConfig)
	// Create a gRPC dial option with TLS credentials for secure connection
	grpc.DialerCredentials(creds)
	// Return the gRPC dial option configured with TLS credentials
	return grpc.DialerCredentials(creds), nil
}
