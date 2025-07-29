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
	ApplicationName   = "sleepTaskList"
	SleepWorkflowName = "sleepWorkflow"
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
	sleepDuration := 30 * time.Second
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "sleep_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}
	h.StartWorkflow(workflowOptions, SleepWorkflowName, sleepDuration)
}

func registerWorkflowAndActivity(h *common.SampleHelper) {
	h.RegisterWorkflowWithAlias(sleepWorkflow, SleepWorkflowName)
	h.RegisterActivity(mainSleepActivity)
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
