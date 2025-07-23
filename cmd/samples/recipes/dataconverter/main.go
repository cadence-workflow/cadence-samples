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
	ApplicationName = "dataConverterTaskList"
)

func startWorkers(h *common.SampleHelper) {
	workerOptions := worker.Options{
		MetricsScope: h.WorkerMetricScope,
		Logger:       h.Logger,
		FeatureFlags: client.FeatureFlags{
			WorkflowExecutionAlreadyCompletedErrorEnabled: true,
		},
		DataConverter: NewJSONDataConverter(),
	}
	h.StartWorkers(h.Config.DomainName, ApplicationName, workerOptions)
}

func startWorkflow(h *common.SampleHelper) {
	input := MyPayload{Msg: "hello", Count: 1}
	workflowOptions := client.StartWorkflowOptions{
		ID:                              "dataconverter_" + uuid.New(),
		TaskList:                        ApplicationName,
		ExecutionStartToCloseTimeout:    time.Minute,
		DecisionTaskStartToCloseTimeout: time.Minute,
	}
	h.StartWorkflow(workflowOptions, DataConverterWorkflowName, input)
}

func registerWorkflowAndActivity(h *common.SampleHelper) {
	h.RegisterWorkflowWithAlias(dataConverterWorkflow, DataConverterWorkflowName)
	h.RegisterActivity(dataConverterActivity)
}

func main() {
	var mode string
	flag.StringVar(&mode, "m", "trigger", "Mode is worker or trigger.")
	flag.Parse()

	var h common.SampleHelper
	h.DataConverter = NewJSONDataConverter()
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
