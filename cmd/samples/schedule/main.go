package main

import (
	"flag"
	"strings"

	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/zap"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
)

func registerWorkflowAndActivity(h *common.SampleHelper) {
	h.RegisterWorkflowWithAlias(scheduledWorkflow, scheduledWorkflowName)
	h.RegisterActivity(scheduledActivity)
}

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

// scenarios maps the -scenario flag to its driver. Each driver builds its own client and
// is self-contained (creates and cleans up its own schedules).
var scenarios = map[string]func(*common.SampleHelper){
	"lifecycle":     runLifecycle,     // full lifecycle round-trip
	"overlap":       runOverlap,       // overlap policies (needs worker, sleeping runs)
	"catchup":       runCatchUp,       // catch-up on unpause (needs worker)
	"pagination":    runPagination,    // List pagination via NextPageToken
	"dataconverter": runDataConverter, // §2.8 — custom DataConverter for Memo, both levels (no worker needed)
}

func scenarioNames() string {
	names := make([]string, 0, len(scenarios))
	for n := range scenarios {
		names = append(names, n)
	}
	return strings.Join(names, " | ")
}

func main() {
	var mode, scenario string
	flag.StringVar(&mode, "m", "manage", "Mode: worker | manage")
	flag.StringVar(&scenario, "scenario", "lifecycle", "manage scenario: "+scenarioNames())
	flag.Parse()

	var h common.SampleHelper
	h.SetupServiceConfig()

	switch mode {
	case "worker":
		registerWorkflowAndActivity(&h)
		startWorkers(&h)
		// The workers are supposed to be long running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by CMD+C.
		select {}
	case "manage":
		run, ok := scenarios[scenario]
		if !ok {
			h.Logger.Fatal("Unknown scenario", zap.String("scenario", scenario), zap.String("valid", scenarioNames()))
		}
		run(&h)
	default:
		h.Logger.Fatal("Unknown mode", zap.String("mode", mode), zap.String("valid", "worker|manage"))
	}
}
