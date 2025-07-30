package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/pborman/uuid"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
)

const (
	ApplicationName = "dataConverterTaskList"
)

func startWorkers(h *common.SampleHelper) {
	workerOptions := worker.Options{
		MetricsScope: h.WorkerMetricScope,
		Logger:       h.Logger,
		FeatureFlags: client.FeatureFlags{
			WorkflowExecutionAlreadyCompletedErrorEnabled: true,
		},
		DataConverter: NewCompressedJSONDataConverter(),
	}
	h.StartWorkers(h.Config.DomainName, ApplicationName, workerOptions)
}

func startWorkflow(h *common.SampleHelper) {
	// Create a large payload to demonstrate compression benefits
	largeInput := CreateLargePayload()

	// Show compression statistics before starting workflow
	converter := NewCompressedJSONDataConverter()
	originalSize, compressedSize, compressionPercentage, err := GetPayloadSizeInfo(largeInput, converter)
	if err != nil {
		fmt.Printf("Error calculating compression stats: %v\n", err)
	} else {
		fmt.Printf("=== Compression Statistics ===\n")
		fmt.Printf("Original JSON size: %d bytes (%.2f KB)\n", originalSize, float64(originalSize)/1024.0)
		fmt.Printf("Compressed size: %d bytes (%.2f KB)\n", compressedSize, float64(compressedSize)/1024.0)
		fmt.Printf("Compression ratio: %.2f%% reduction\n", compressionPercentage)
		fmt.Printf("Space saved: %d bytes (%.2f KB)\n", originalSize-compressedSize, float64(originalSize-compressedSize)/1024.0)
		fmt.Printf("=============================\n\n")
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:                              "dataconverter_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}
	h.StartWorkflow(workflowOptions, LargeDataConverterWorkflowName, largeInput)
}

func registerWorkflowAndActivity(h *common.SampleHelper) {
	h.RegisterWorkflowWithAlias(largeDataConverterWorkflow, LargeDataConverterWorkflowName)
	h.RegisterActivity(largeDataConverterActivity)
}

func main() {
	var mode string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger.")
	flag.Parse()

	var h common.SampleHelper
	h.DataConverter = NewCompressedJSONDataConverter()
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
