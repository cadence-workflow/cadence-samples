package main

import (
	"flag"
	"time"

	"github.com/pborman/uuid"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
)

const (
	tl = "localActivityFailure"
)

// This needs to be done as part of a bootstrap step when the process starts.
// The workers are supposed to be long running.
func startWorkers(h *common.SampleHelper) {
	// Configure worker options.
	workerOptions := worker.Options{
		MetricsScope: h.WorkerMetricScope,
		Logger:       h.Logger,
	}
	h.StartWorkers(h.Config.DomainName, "localActivityFailure", workerOptions)
}

func startNormalActivityTest(h *common.SampleHelper) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "localactivityfailure_" + uuid.New(),
		TaskList:                        tl,
		ExecutionStartToCloseTimeout:    time.Minute * 3,
		DecisionTaskStartToCloseTimeout: time.Minute,
		WorkflowIDReusePolicy:           client.WorkflowIDReusePolicyAllowDuplicate,
	}
	h.StartWorkflow(workflowOptions, testNormalActivity)
}

func startLocalActivityTest(h *common.SampleHelper) {
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "localactivityfailure_" + uuid.New(),
		TaskList:                        tl,
		ExecutionStartToCloseTimeout:    time.Minute * 3,
		DecisionTaskStartToCloseTimeout: time.Minute,
		WorkflowIDReusePolicy:           client.WorkflowIDReusePolicyAllowDuplicate,
	}
	h.StartWorkflow(workflowOptions, testLocalActivity)
}

func main() {
	var mode, workflowID, signal string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker, trigger or query.")
	flag.StringVar(&workflowID, "w", "", "WorkflowID")
	flag.StringVar(&signal, "s", "signal_data", "SignalData")
	flag.Parse()

	var h common.SampleHelper
	h.SetupServiceConfig()

	switch mode {
	case "worker":
		h.RegisterWorkflow(testLocalActivity)
		h.RegisterWorkflow(testNormalActivity)
		h.RegisterActivity(SumActivity)
		startWorkers(&h)

		// The workers are supposed to be long running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by CMD+C.
		select {}
	case "trigger-local":
		startNormalActivityTest(&h)
	case "trigger-normal":
		startLocalActivityTest(&h)
	}
}
