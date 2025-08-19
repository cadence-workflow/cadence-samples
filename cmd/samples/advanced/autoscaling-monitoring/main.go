package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pborman/uuid"
	"go.uber.org/cadence/client"
	"gopkg.in/yaml.v2"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
	"github.com/uber-go/tally"
	"go.uber.org/zap"
)

const (
	ApplicationName = "autoscaling-monitoring"
)

func main() {
	var mode string
	var configFile string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker, trigger, or server.")
	flag.StringVar(&configFile, "config", "config/autoscaling.yaml", "Configuration file path.")
	flag.Parse()

	// Load configuration
	config := loadConfiguration(configFile)

	// Setup common helper with our configuration
	var h common.SampleHelper
	h.Config = config.Configuration

	// Set up logging
	logger, err := zap.NewDevelopment()
	if err != nil {
		panic(fmt.Sprintf("Failed to setup logger: %v", err))
	}
	h.Logger = logger

	// Set up service client using our config
	h.Builder = common.NewBuilder(logger).
		SetHostPort(config.HostNameAndPort).
		SetDomain(config.DomainName)

	service, err := h.Builder.BuildServiceClient()
	if err != nil {
		panic(fmt.Sprintf("Failed to build service client: %v", err))
	}
	h.Service = service

	// Set up metrics scope (noop for now, Prometheus will be set up separately)
	h.WorkerMetricScope = tally.NoopScope
	h.ServiceMetricScope = tally.NoopScope

	switch mode {
	case "worker":
		startWorkers(&h, &config)
	case "trigger":
		startWorkflow(&h, &config)
	case "server":
		startPrometheusServer(&h)
	default:
		fmt.Printf("Unknown mode: %s\n", mode)
		os.Exit(1)
	}
}

// loadConfiguration loads the autoscaling configuration from file
func loadConfiguration(configFile string) AutoscalingConfiguration {
	// Read config file
	configData, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Failed to read config file: %v, using defaults\n", err)
		return DefaultAutoscalingConfiguration()
	}

	// Parse config
	var config AutoscalingConfiguration
	if err := yaml.Unmarshal(configData, &config); err != nil {
		fmt.Printf("Error parsing configuration: %v, using defaults\n", err)
		return DefaultAutoscalingConfiguration()
	}

	// Ensure base Configuration fields are populated
	if config.DomainName == "" {
		config.DomainName = "default"
	}
	if config.ServiceName == "" {
		config.ServiceName = "cadence-frontend"
	}
	if config.HostNameAndPort == "" {
		config.HostNameAndPort = "localhost:7833"
	}

	fmt.Printf("Loaded configuration from %s\n", configFile)
	return config
}

func startWorkers(h *common.SampleHelper, config *AutoscalingConfiguration) {
	startWorkersWithAutoscaling(h, config)
}

func startWorkflow(h *common.SampleHelper, config *AutoscalingConfiguration) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "autoscaling_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute * 10,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}

	// Use configuration values
	iterations := config.Autoscaling.LoadGeneration.Iterations
	batchDelay := config.Autoscaling.LoadGeneration.BatchDelay
	minProcessingTime := config.Autoscaling.LoadGeneration.MinProcessingTime
	maxProcessingTime := config.Autoscaling.LoadGeneration.MaxProcessingTime
	h.StartWorkflow(workflowOptions, autoscalingWorkflowName, iterations, batchDelay, minProcessingTime, maxProcessingTime)

	fmt.Printf("Started autoscaling workflow with %d iterations\n", iterations)
	fmt.Println("Monitor the worker performance and autoscaling behavior in Grafana:")
	fmt.Println("http://localhost:3000/d/dehkspwgabvuoc/cadence-client")
}

func startPrometheusServer(h *common.SampleHelper) {
	startPrometheusHTTPServer(h)

	// Block until interrupted
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)
	fmt.Println("Prometheus server started. Press Ctrl+C to stop...")
	<-done
	fmt.Println("Shutting down Prometheus server...")
}
