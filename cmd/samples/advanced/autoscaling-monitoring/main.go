package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pborman/uuid"
	"go.uber.org/cadence/client"
	"gopkg.in/yaml.v2"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/prometheus"
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

	// Set up metrics scope with Tally Prometheus reporter
	var (
		safeCharacters  = []rune{'_'}
		sanitizeOptions = tally.SanitizeOptions{
			NameCharacters: tally.ValidCharacters{
				Ranges:     tally.AlphanumericRange,
				Characters: safeCharacters,
			},
			KeyCharacters: tally.ValidCharacters{
				Ranges:     tally.AlphanumericRange,
				Characters: safeCharacters,
			},
			ValueCharacters: tally.ValidCharacters{
				Ranges:     tally.AlphanumericRange,
				Characters: safeCharacters,
			},
			ReplacementCharacter: tally.DefaultReplacementCharacter,
		}
	)

	// Create Prometheus reporter
	reporter := prometheus.NewReporter(prometheus.Options{})

	// Create root scope with proper options
	scope, closer := tally.NewRootScope(tally.ScopeOptions{
		Tags:            map[string]string{"service": "autoscaling-monitoring"},
		SanitizeOptions: &sanitizeOptions,
		CachedReporter:  reporter,
	}, 10)
	defer closer.Close()

	// Set up HTTP handler for metrics endpoint
	if config.Prometheus != nil {
		go func() {
			http.Handle("/metrics", reporter.HTTPHandler())
			logger.Info("Starting Prometheus metrics server",
				zap.String("port", config.Prometheus.ListenAddress))
			if err := http.ListenAndServe(config.Prometheus.ListenAddress, nil); err != nil {
				logger.Error("Failed to start metrics server", zap.Error(err))
			}
		}()
	}

	// Set up metrics scope for helper
	h.WorkerMetricScope = scope
	h.ServiceMetricScope = scope

	switch mode {
	case "worker":
		startWorkers(&h, &config)
	case "trigger":
		startWorkflow(&h, &config)
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
