package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pborman/uuid"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/zap"
	"google.golang.org/grpc/credentials"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/cadence/compatibility"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/peer"
	"go.uber.org/yarpc/peer/hostport"
	"go.uber.org/yarpc/transport/grpc"
)

const (
	ApplicationName = "tlsTaskList"
	TLSWorkflowName = "tlsWorkflow"
)

func startWorkers(h *common.SampleHelper) {
	workerOptions := worker.Options{
		MetricsScope: h.WorkerMetricScope,
		Logger:       h.Logger,
		FeatureFlags: client.FeatureFlags{
			WorkflowExecutionAlreadyCompletedErrorEnabled: true,
		},
	}
	h.StartWorkers(h.Config.DomainName, ApplicationName, workerOptions)
}

func startWorkflow(h *common.SampleHelper) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "tls_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    5 * time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}
	h.StartWorkflow(workflowOptions, TLSWorkflowName)
}

func registerWorkflowAndActivity(h *common.SampleHelper) {
	h.RegisterWorkflowWithAlias(tlsWorkflow, TLSWorkflowName)
	h.RegisterActivity(setupCertificatesActivity)
	h.RegisterActivity(testTLSConnectionActivity)
	h.RegisterActivity(testStandardConnectionActivity)
	h.RegisterActivity(cleanupCertificatesActivity)
}

func main() {
	var mode string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger.")
	flag.Parse()

	var h common.SampleHelper
	h.SetupServiceConfig()

	switch mode {
	case "worker":
		registerWorkflowAndActivity(&h)
		startWorkers(&h)
		select {}
	case "trigger":
		startWorkflow(&h)
	}
}

// setupCertificatesForTesting handles the complete certificate setup workflow for testing
// Returns the certificates directory path on success, empty string on failure
func setupCertificatesForTesting() string {
	fmt.Println("Creating fresh TLS certificates...")
	certsDir := "cmd/samples/recipes/tls/certs"

	if err := createCertificates(certsDir); err != nil {
		fmt.Printf("❌ Failed to create certificates: %v\n", err)
		return ""
	}

	fmt.Println("✅ Certificates created successfully")
	fmt.Println()
	return certsDir
}

// createCertificates creates fresh TLS certificates for testing
func createCertificates(certsDir string) error {
	// Ensure certs directory exists
	if err := os.MkdirAll(certsDir, 0755); err != nil {
		return fmt.Errorf("failed to create certs directory: %v", err)
	}

	// Clean up any existing certificates first
	cleanupCertificates(certsDir)

	// Generate CA private key
	if err := runOpenSSLCommand("genrsa", "-out", filepath.Join(certsDir, "ca.key"), "4096"); err != nil {
		return fmt.Errorf("failed to generate CA key: %v", err)
	}

	// Generate CA certificate
	if err := runOpenSSLCommand("req", "-new", "-x509", "-key", filepath.Join(certsDir, "ca.key"),
		"-sha256", "-subj", "/C=US/ST=CA/O=Cadence/CN=CadenceCA", "-days", "3650",
		"-out", filepath.Join(certsDir, "ca.crt")); err != nil {
		return fmt.Errorf("failed to generate CA certificate: %v", err)
	}

	// Generate server private key
	if err := runOpenSSLCommand("genrsa", "-out", filepath.Join(certsDir, "server.key"), "4096"); err != nil {
		return fmt.Errorf("failed to generate server key: %v", err)
	}

	// Generate server certificate request
	if err := runOpenSSLCommand("req", "-new", "-key", filepath.Join(certsDir, "server.key"),
		"-subj", "/C=US/ST=CA/O=Cadence/CN=localhost",
		"-out", filepath.Join(certsDir, "server.csr")); err != nil {
		return fmt.Errorf("failed to generate server CSR: %v", err)
	}

	// Sign server certificate
	if err := runOpenSSLCommand("x509", "-req", "-in", filepath.Join(certsDir, "server.csr"),
		"-CA", filepath.Join(certsDir, "ca.crt"), "-CAkey", filepath.Join(certsDir, "ca.key"),
		"-CAcreateserial", "-out", filepath.Join(certsDir, "server.crt"),
		"-days", "365", "-sha256"); err != nil {
		return fmt.Errorf("failed to sign server certificate: %v", err)
	}

	// Generate client private key
	if err := runOpenSSLCommand("genrsa", "-out", filepath.Join(certsDir, "client.key"), "4096"); err != nil {
		return fmt.Errorf("failed to generate client key: %v", err)
	}

	// Generate client certificate request
	if err := runOpenSSLCommand("req", "-new", "-key", filepath.Join(certsDir, "client.key"),
		"-subj", "/C=US/ST=CA/O=Cadence/CN=client",
		"-out", filepath.Join(certsDir, "client.csr")); err != nil {
		return fmt.Errorf("failed to generate client CSR: %v", err)
	}

	// Sign client certificate
	if err := runOpenSSLCommand("x509", "-req", "-in", filepath.Join(certsDir, "client.csr"),
		"-CA", filepath.Join(certsDir, "ca.crt"), "-CAkey", filepath.Join(certsDir, "ca.key"),
		"-CAcreateserial", "-out", filepath.Join(certsDir, "client.crt"),
		"-days", "365", "-sha256"); err != nil {
		return fmt.Errorf("failed to sign client certificate: %v", err)
	}

	// Clean up CSR files
	os.Remove(filepath.Join(certsDir, "server.csr"))
	os.Remove(filepath.Join(certsDir, "client.csr"))

	return nil
}

// cleanupCertificates removes all certificate files
func cleanupCertificates(certsDir string) {
	certFiles := []string{"ca.key", "ca.crt", "ca.srl", "server.key", "server.crt", "server.csr", "client.key", "client.crt", "client.csr"}
	for _, file := range certFiles {
		os.Remove(filepath.Join(certsDir, file))
	}
}

// runOpenSSLCommand executes an openssl command with given arguments
func runOpenSSLCommand(args ...string) error {
	cmd := exec.Command("openssl", args...)
	cmd.Stdout = nil // Suppress output
	cmd.Stderr = nil // Suppress errors for cleaner output
	return cmd.Run()
}

// createTLSClient creates a Cadence client with TLS configuration
func createTLSClient(logger *zap.Logger, clientCertPath, clientKeyPath, caCertPath string) (client.Client, error) {
	// Load client certificate
	clientCert, err := tls.LoadX509KeyPair(clientCertPath, clientKeyPath)
	if err != nil {
		return nil, err
	}

	// Load CA certificate
	caCert, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Configure TLS
	tlsConfig := tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            caCertPool,
		Certificates:       []tls.Certificate{clientCert},
	}
	creds := credentials.NewTLS(&tlsConfig)

	// Create gRPC transport with TLS
	grpcTransport := grpc.NewTransport()
	dialer := grpcTransport.NewDialer(grpc.DialerCredentials(creds))
	outbound := grpcTransport.NewOutbound(peer.NewSingle(hostport.Identify("localhost:7933"), dialer))

	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: "cadence-client",
		Outbounds: yarpc.Outbounds{
			"cadence-frontend": {Unary: outbound},
		},
	})

	if err := dispatcher.Start(); err != nil {
		return nil, err
	}

	// Create service client
	clientConfig := dispatcher.ClientConfig("cadence-frontend")
	service := compatibility.NewThrift2ProtoAdapter(
		apiv1.NewDomainAPIYARPCClient(clientConfig),
		apiv1.NewWorkflowAPIYARPCClient(clientConfig),
		apiv1.NewWorkerAPIYARPCClient(clientConfig),
		apiv1.NewVisibilityAPIYARPCClient(clientConfig),
	)

	return client.NewClient(service, "sample-domain", &client.Options{}), nil
}

// createNonFatalLogger creates a logger that doesn't call os.Exit
func createNonFatalLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level = zap.NewAtomicLevelAt(zap.ErrorLevel) // Only show errors
	logger, _ := config.Build()
	return logger
}

// fileExists checks if a file exists
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err)
}

// getSimpleError extracts a simple error message for display
func getSimpleError(err error) string {
	errStr := err.Error()

	switch {
	case contains(errStr, "no such file or directory"):
		return "Certificate file not found"
	case contains(errStr, "connection refused"):
		return "Cadence server not running"
	case contains(errStr, "certificate verify failed"):
		return "Certificate verification failed"
	case contains(errStr, "tls: handshake failure"):
		return "TLS handshake failed"
	case contains(errStr, "dial tcp"):
		return "Cannot connect to server"
	default:
		return "Connection error: " + errStr
	}
}

// contains checks if a string contains a substring
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
