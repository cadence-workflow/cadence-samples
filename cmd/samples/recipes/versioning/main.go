package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) worker.Worker {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope:      h.WorkerMetricScope,
		Logger:            h.Logger,
		WorkerStopTimeout: 1 * time.Second,
	}
	return h.StartWorkers(h.Config.DomainName, ApplicationName, workerOptions)
}

func startWorkflow(h *common.SampleHelper) {
	// Allow to run only one Versioned workflow at a time
	workflowOptions := client.StartWorkflowOptions{
		ID:                              VersionedWorkflowID,
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Hour,
		DecisionTaskStartToCloseTimeout: time.Minute,
		WorkflowIDReusePolicy:           client.WorkflowIDReusePolicyAllowDuplicate,
	}
	h.StartWorkflow(workflowOptions, VersionedWorkflowName, 0)
}

// stopWorkflow sends a signal to the workflow to stop it gracefully.
func stopWorkflow(h *common.SampleHelper) {
	h.Logger.Info("Stopping workflow")
	h.SignalWorkflow(VersionedWorkflowID, StopSignalName, "")
}

func main() {
	var mode string
	var version string

	flag.StringVar(&mode, "m", "trigger", "Mode is worker (version flag is required), trigger (start a new workflow, only one allowed), stop (stop a running workflow). Default is trigger.")
	flag.StringVar(&version, "v", "", "Version of the workflow to run, supported versions are 1, 2, 3, or 4. Required in worker mode.")

	flag.Parse()

	var h common.SampleHelper
	h.SetupServiceConfig()

	switch mode {
	case "worker":
		switch version {
		case "1", "v1":
			SetupHelperForVersionedWorkflowV1(&h)

		case "2", "v2":
			SetupHelperForVersionedWorkflowV2(&h)

		case "3", "v3":
			SetupHelperForVersionedWorkflowV3(&h)

		case "4", "v4":
			SetupHelperForVersionedWorkflowV4(&h)

		case "":
			fmt.Printf("-v flag is required for worker mode. Use -v 1, -v 2, -v 3, or -v 4 to specify the version.\n")
			os.Exit(1)

		default:
			fmt.Printf("Invalid version specified:%s . Use -v 1, -v 2, -v 3, or -v 4.", version)
			os.Exit(1)
		}

		startWorkers(&h)

		// The workers are supposed to be long-running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by CMD+C.
		select {}

	case "trigger":
		startWorkflow(&h)

	case "stop":
		stopWorkflow(&h)

	default:
		fmt.Printf("Invalid mode specified: %s. Use -m worker, -m trigger, -m stop.\n", mode)
		os.Exit(1)
	}
}
