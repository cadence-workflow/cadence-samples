package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pborman/uuid"
	"go.uber.org/cadence/client"
	"gopkg.in/yaml.v2"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
)

const (
	ApplicationName = "autoscaling-monitoring"
)

// Global configuration
var config AutoscalingConfiguration

func main() {
	var mode string
	var configFile string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker, trigger, or server.")
	flag.StringVar(&configFile, "config", "config/autoscaling.yaml", "Configuration file path.")
	flag.Parse()

	// Load configuration
	loadConfiguration(configFile)

	// Setup common helper with our configuration
	var h common.SampleHelper
	h.Config = config.Configuration
	h.SetupServiceConfig()

	switch mode {
	case "worker":
		startWorkers(&h)

		// The workers are supposed to be long running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by CMD+C.
		select {}
	case "trigger":
		startWorkflow(&h)
	case "server":
		startPrometheusServer(&h)
	default:
		fmt.Printf("Unknown mode: %s\n", mode)
		os.Exit(1)
	}
}

// loadConfiguration loads the autoscaling configuration from file
func loadConfiguration(configFile string) {
	// Read config file
	configData, err := ioutil.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Failed to read config file: %v, using defaults\n", err)
		config = DefaultAutoscalingConfiguration()
		return
	}

	// Parse config
	if err := yaml.Unmarshal(configData, &config); err != nil {
		fmt.Printf("Error parsing configuration: %v, using defaults\n", err)
		config = DefaultAutoscalingConfiguration()
		return
	}

	fmt.Printf("Loaded configuration from %s\n", configFile)
}

func startWorkers(h *common.SampleHelper) {
	startWorkersWithAutoscaling(h)
}

func startWorkflow(h *common.SampleHelper) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "autoscaling_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute * 10,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}

	// Use iterations from configuration
	iterations := config.Autoscaling.LoadGeneration.Iterations
	h.StartWorkflow(workflowOptions, autoscalingWorkflowName, iterations)

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
