// Custom worker.go for Data samples - NOT generated
// This sample requires custom DataConverters in worker options, one per sample.

package main

import (
	"fmt"
	"os"

	"github.com/uber-go/tally"
	apiv1 "github.com/uber/cadence-idl/go/proto/api/v1"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/compatibility"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/peer"
	yarpchostport "go.uber.org/yarpc/peer/hostport"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	HostPort       = "127.0.0.1:7833"
	Domain         = "cadence-samples"
	ClientName     = "cadence-samples-worker"
	CadenceService = "cadence-frontend"

	// Each sample uses its own task list so it can have its own DataConverter.
	TaskListCompression = "cadence-samples-data-compression"
	TaskListEncryption  = "cadence-samples-data-encryption"
	TaskListS3          = "cadence-samples-data-s3"
)

// StartWorker starts one worker per DataConverter sample and prints startup stats for each.
func StartWorker() {
	logger := BuildLogger()
	cadenceClient := BuildCadenceClient()

	startCompressionWorker(logger, cadenceClient)
	startEncryptionWorker(logger, cadenceClient)
	startS3OffloadWorker(logger, cadenceClient)

	printCompressionStats()
	printEncryptionStats()
	printS3OffloadStats()
}

func startCompressionWorker(logger *zap.Logger, cadenceClient workflowserviceclient.Interface) {
	dataConverter := NewCompressedJSONDataConverter()
	workerOptions := worker.Options{
		Logger:        logger,
		MetricsScope:  tally.NewTestScope(TaskListCompression, nil),
		DataConverter: dataConverter,
	}

	w := worker.New(cadenceClient, Domain, TaskListCompression, workerOptions)
	w.RegisterWorkflowWithOptions(CompressionDataConverterWorkflow, workflow.RegisterOptions{Name: "cadence_samples.CompressionDataConverterWorkflow"})
	w.RegisterActivityWithOptions(CompressionDataConverterActivity, activity.RegisterOptions{Name: "cadence_samples.CompressionDataConverterActivity"})

	if err := w.Start(); err != nil {
		panic("Failed to start compression worker: " + err.Error())
	}
	logger.Info("Started compression worker", zap.String("task_list", TaskListCompression))
}

func startEncryptionWorker(logger *zap.Logger, cadenceClient workflowserviceclient.Interface) {
	key := LoadEncryptionKey()
	dataConverter, err := NewEncryptedJSONDataConverter(key)
	if err != nil {
		panic("Failed to create encryption data converter: " + err.Error())
	}
	workerOptions := worker.Options{
		Logger:        logger,
		MetricsScope:  tally.NewTestScope(TaskListEncryption, nil),
		DataConverter: dataConverter,
	}

	w := worker.New(cadenceClient, Domain, TaskListEncryption, workerOptions)
	w.RegisterWorkflowWithOptions(EncryptionDataConverterWorkflow, workflow.RegisterOptions{Name: "cadence_samples.EncryptionDataConverterWorkflow"})
	w.RegisterActivityWithOptions(EncryptionDataConverterActivity, activity.RegisterOptions{Name: "cadence_samples.EncryptionDataConverterActivity"})

	if err := w.Start(); err != nil {
		panic("Failed to start encryption worker: " + err.Error())
	}
	logger.Info("Started encryption worker", zap.String("task_list", TaskListEncryption))
}

func startS3OffloadWorker(logger *zap.Logger, cadenceClient workflowserviceclient.Interface) {
	store := NewLocalFSBlobStore()
	dataConverter := NewS3OffloadDataConverter(store, "cadence-samples-data-s3", defaultThresholdBytes)
	workerOptions := worker.Options{
		Logger:        logger,
		MetricsScope:  tally.NewTestScope(TaskListS3, nil),
		DataConverter: dataConverter,
	}

	w := worker.New(cadenceClient, Domain, TaskListS3, workerOptions)
	w.RegisterWorkflowWithOptions(S3OffloadDataConverterWorkflow, workflow.RegisterOptions{Name: "cadence_samples.S3OffloadDataConverterWorkflow"})
	w.RegisterActivityWithOptions(S3OffloadDataConverterActivity, activity.RegisterOptions{Name: "cadence_samples.S3OffloadDataConverterActivity"})

	if err := w.Start(); err != nil {
		panic("Failed to start S3 offload worker: " + err.Error())
	}
	logger.Info("Started S3 offload worker", zap.String("task_list", TaskListS3))
}

// printCompressionStats displays gzip compression statistics for the sample payload.
func printCompressionStats() {
	largePayload := CreateLargePayload()
	originalSize, compressedSize, compressionPercentage, err := GetPayloadSizeInfo(largePayload, NewCompressedJSONDataConverter())
	if err != nil {
		fmt.Printf("Error calculating compression stats: %v\n", err)
		return
	}

	fmt.Printf("\n=== Compression Sample Statistics ===\n")
	fmt.Printf("Original JSON size:  %d bytes (%.2f KB)\n", originalSize, float64(originalSize)/1024.0)
	fmt.Printf("Compressed size:     %d bytes (%.2f KB)\n", compressedSize, float64(compressedSize)/1024.0)
	fmt.Printf("Compression ratio:   %.2f%% reduction\n", compressionPercentage)
	fmt.Printf("Space saved:         %d bytes (%.2f KB)\n", originalSize-compressedSize, float64(originalSize-compressedSize)/1024.0)
	fmt.Printf("Start workflow: cadence --domain %s workflow start --tl %s --workflow_type cadence_samples.CompressionDataConverterWorkflow --et 60\n", Domain, TaskListCompression)
	fmt.Printf("=====================================\n\n")
}

// printEncryptionStats displays AES-256-GCM encryption statistics for the sample record.
func printEncryptionStats() {
	record := CreateSensitiveCustomerRecord()
	converter, err := NewEncryptedJSONDataConverter(demoEncryptionKey)
	if err != nil {
		fmt.Printf("Error creating encryption converter for stats: %v\n", err)
		return
	}
	plaintextSize, ciphertextSize, preview, err := GetEncryptionSizeInfo(record, converter)
	if err != nil {
		fmt.Printf("Error calculating encryption stats: %v\n", err)
		return
	}

	fmt.Printf("\n=== Encryption Sample Statistics ===\n")
	fmt.Printf("Plaintext JSON size:  %d bytes\n", plaintextSize)
	fmt.Printf("Ciphertext size:      %d bytes (overhead: %d bytes nonce+tag)\n", ciphertextSize, ciphertextSize-plaintextSize)
	fmt.Printf("Ciphertext preview:   %s\n", preview)
	fmt.Printf("Start workflow: cadence --domain %s workflow start --tl %s --workflow_type cadence_samples.EncryptionDataConverterWorkflow --et 60\n", Domain, TaskListEncryption)
	fmt.Printf("====================================\n\n")
}

// printS3OffloadStats displays claim-check offload statistics for the sample payload.
func printS3OffloadStats() {
	payload := CreateS3LargePayload()
	jsonSize, cadenceBytes, err := GetS3OffloadSizeInfo(payload, defaultThresholdBytes)
	if err != nil {
		fmt.Printf("Error calculating S3 offload stats: %v\n", err)
		return
	}

	fmt.Printf("\n=== S3 Offload Sample Statistics ===\n")
	fmt.Printf("Full payload JSON size:    %d bytes (%.2f KB)\n", jsonSize, float64(jsonSize)/1024.0)
	fmt.Printf("Stored in BlobStore:       %d bytes (%.2f KB)\n", jsonSize, float64(jsonSize)/1024.0)
	fmt.Printf("Stored in Cadence history: %d bytes (claim-check reference only)\n", cadenceBytes)
	fmt.Printf("Reduction in Cadence:      %.1f%%\n", 100.0*(1.0-float64(cadenceBytes)/float64(jsonSize)))
	fmt.Printf("BlobStore location:        %s/cadence-samples-data-s3/\n", os.TempDir())
	fmt.Printf("Start workflow: cadence --domain %s workflow start --tl %s --workflow_type cadence_samples.S3OffloadDataConverterWorkflow --et 60\n", Domain, TaskListS3)
	fmt.Printf("=====================================\n\n")
}

func BuildCadenceClient(dialOptions ...grpc.DialOption) workflowserviceclient.Interface {
	grpcTransport := grpc.NewTransport()
	myChooser := peer.NewSingle(
		yarpchostport.Identify(HostPort),
		grpcTransport.NewDialer(dialOptions...),
	)
	outbound := grpcTransport.NewOutbound(myChooser)

	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: ClientName,
		Outbounds: yarpc.Outbounds{
			CadenceService: {Unary: outbound},
		},
	})
	if err := dispatcher.Start(); err != nil {
		panic("Failed to start dispatcher: " + err.Error())
	}

	clientConfig := dispatcher.ClientConfig(CadenceService)

	return compatibility.NewThrift2ProtoAdapter(
		apiv1.NewDomainAPIYARPCClient(clientConfig),
		apiv1.NewWorkflowAPIYARPCClient(clientConfig),
		apiv1.NewWorkerAPIYARPCClient(clientConfig),
		apiv1.NewVisibilityAPIYARPCClient(clientConfig),
	)
}

func BuildLogger() *zap.Logger {
	config := zap.NewDevelopmentConfig()
	config.Level.SetLevel(zapcore.InfoLevel)

	var err error
	logger, err := config.Build()
	if err != nil {
		panic("Failed to setup logger: " + err.Error())
	}

	return logger
}
