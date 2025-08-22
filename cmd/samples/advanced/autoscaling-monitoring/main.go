package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pborman/uuid"
	"go.uber.org/cadence/client"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
	"github.com/uber-go/tally"
	"github.com/uber-go/tally/prometheus"
	"go.uber.org/zap"
)

const (
	ApplicationName = "autoscaling-monitoring"
)

func main() {
	// Parse command line arguments
	var mode string
	flag.StringVar(&mode, "m", "worker", "Mode: worker or trigger")
	flag.Parse()

	// Load configuration
	configFile := "config/autoscaling.yaml"
	config := loadConfiguration(configFile)

	// Setup common helper with our configuration
	var h common.SampleHelper
	h.Config = config.ToCommonConfiguration()

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

	// Set up metrics scope for helper
	h.WorkerMetricScope = scope
	h.ServiceMetricScope = scope

	switch mode {
	case "worker":
		// Start metrics server only in worker mode
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
		startWorkers(&h, &config)
	case "trigger":
		startWorkflow(&h, &config)
	default:
		fmt.Printf("Unknown mode: %s\n", mode)
		os.Exit(1)
	}
}

func startWorkers(h *common.SampleHelper, config *AutoscalingConfiguration) {
	startWorkersWithAutoscaling(h, config)
}

func startWorkflow(h *common.SampleHelper, config *AutoscalingConfiguration) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              fmt.Sprintf("autoscaling_%s", uuid.New()),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute * 10,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}

	// Use configuration values
	workflows := config.Autoscaling.LoadGeneration.Workflows
	workflowDelay := config.Autoscaling.LoadGeneration.WorkflowDelay
	activitiesPerWorkflow := config.Autoscaling.LoadGeneration.ActivitiesPerWorkflow
	batchDelay := config.Autoscaling.LoadGeneration.BatchDelay
	minProcessingTime := config.Autoscaling.LoadGeneration.MinProcessingTime
	maxProcessingTime := config.Autoscaling.LoadGeneration.MaxProcessingTime

	// Start multiple workflows with delays
	for i := 0; i < workflows; i++ {
		workflowOptions.ID = fmt.Sprintf("autoscaling_%d_%s", i, uuid.New())
		h.StartWorkflow(workflowOptions, autoscalingWorkflowName, activitiesPerWorkflow, batchDelay, minProcessingTime, maxProcessingTime)

		// Add delay between workflows (except for the last one)
		if i < workflows-1 {
			time.Sleep(time.Duration(workflowDelay) * time.Millisecond)
		}
	}

	fmt.Printf("Started %d autoscaling workflows with %d activities each\n", workflows, activitiesPerWorkflow)
	fmt.Println("Monitor the worker performance and autoscaling behavior in Grafana:")
	fmt.Println("http://localhost:3000/d/dehkspwgabvuoc/cadence-client")
}
