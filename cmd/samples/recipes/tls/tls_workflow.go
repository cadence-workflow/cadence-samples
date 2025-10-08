package main

import (
	"context"
	"errors"
	"time"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const tlsWorkflowName = "tlsWorkflow"

// tlsWorkflow demonstrates TLS connection testing workflow
func tlsWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("TLS Workflow started")

	// Set activity options
	activityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	// Test 1: Setup certificates
	logger.Info("Setting up certificates for testing")
	var certsDir string
	err := workflow.ExecuteActivity(ctx, setupCertificatesActivity).Get(ctx, &certsDir)
	if err != nil {
		logger.Error("Certificate setup failed", zap.Error(err))
		return err
	}

	// Test 2: Test TLS connection with valid certificates
	logger.Info("Testing TLS connection with valid certificates")
	var validResult string
	err = workflow.ExecuteActivity(ctx, testTLSConnectionActivity, certsDir+"/client.crt", certsDir+"/client.key", certsDir+"/ca.crt").Get(ctx, &validResult)
	if err != nil {
		logger.Error("Valid TLS connection test failed", zap.Error(err))
	} else {
		logger.Info("Valid TLS connection test result", zap.String("Result", validResult))
	}

	// Test 3: Test TLS connection with missing certificates
	logger.Info("Testing TLS connection with missing certificates")
	var invalidResult string
	err = workflow.ExecuteActivity(ctx, testTLSConnectionActivity, "certs/missing.crt", "certs/missing.key", "certs/missing-ca.crt").Get(ctx, &invalidResult)
	if err != nil {
		logger.Info("Missing certificates test failed as expected", zap.Error(err))
	} else {
		logger.Info("Missing certificates test result", zap.String("Result", invalidResult))
	}

	// Test 4: Test standard non-TLS connection
	logger.Info("Testing standard non-TLS connection")
	var standardResult string
	err = workflow.ExecuteActivity(ctx, testStandardConnectionActivity).Get(ctx, &standardResult)
	if err != nil {
		logger.Error("Standard connection test failed", zap.Error(err))
	} else {
		logger.Info("Standard connection test result", zap.String("Result", standardResult))
	}

	// Cleanup certificates
	logger.Info("Cleaning up certificates")
	var cleanupResult string
	err = workflow.ExecuteActivity(ctx, cleanupCertificatesActivity, certsDir).Get(ctx, &cleanupResult)
	if err != nil {
		logger.Error("Certificate cleanup failed", zap.Error(err))
	} else {
		logger.Info("Certificate cleanup completed", zap.String("Result", cleanupResult))
	}

	logger.Info("TLS Workflow completed successfully")
	return nil
}

// setupCertificatesActivity creates certificates for testing
func setupCertificatesActivity(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Setting up certificates for testing")

	certsDir := setupCertificatesForTesting()
	if certsDir == "" {
		return "", errors.New("failed to setup certificates")
	}

	logger.Info("Certificates setup completed", zap.String("CertsDir", certsDir))
	return certsDir, nil
}

// testTLSConnectionActivity tests TLS connection with given certificates
func testTLSConnectionActivity(ctx context.Context, clientCert, clientKey, caCert string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Testing TLS connection",
		zap.String("ClientCert", clientCert),
		zap.String("ClientKey", clientKey),
		zap.String("CACert", caCert))

	// Pre-check certificate files exist
	if !fileExists(clientCert) || !fileExists(clientKey) || !fileExists(caCert) {
		return "FAILED - Certificate file not found", errors.New("certificate file not found")
	}

	nonFatalLogger := createNonFatalLogger()

	// Create TLS-enabled client manually
	client, err := createTLSClient(nonFatalLogger, clientCert, clientKey, caCert)
	if err != nil {
		errorMsg := "FAILED - " + getSimpleError(err)
		logger.Info("TLS connection failed", zap.String("Error", errorMsg))
		return errorMsg, err
	}
	_ = client

	result := "SUCCESS - TLS connection established"
	logger.Info("TLS connection succeeded")
	return result, nil
}

// testStandardConnectionActivity tests standard non-TLS connection
func testStandardConnectionActivity(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Testing standard connection")

	nonFatalLogger := createNonFatalLogger()

	builder := common.NewBuilder(nonFatalLogger).
		SetHostPort("localhost:7933").
		SetDomain("sample-domain")
	// No SetTLSConfig call - TLS is disabled

	_, err := builder.BuildCadenceClient()
	if err != nil {
		errorMsg := "FAILED - " + getSimpleError(err)
		logger.Info("Standard connection failed", zap.String("Error", errorMsg))
		return errorMsg, err
	}

	result := "SUCCESS - Standard connection established"
	logger.Info("Standard connection succeeded")
	return result, nil
}

// cleanupCertificatesActivity removes certificate files
func cleanupCertificatesActivity(ctx context.Context, certsDir string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Cleaning up certificates", zap.String("CertsDir", certsDir))

	cleanupCertificates(certsDir)

	result := "Certificates cleaned up successfully"
	logger.Info("Certificate cleanup completed")
	return result, nil
}
