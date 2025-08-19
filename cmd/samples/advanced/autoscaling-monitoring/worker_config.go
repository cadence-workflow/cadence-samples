package main

import (
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
)

// startWorkersWithAutoscaling starts workers with autoscaling configuration
func startWorkersWithAutoscaling(h *common.SampleHelper, config *AutoscalingConfiguration) {
	// Configure worker options with autoscaling-friendly settings from config
	workerOptions := worker.Options{
		MetricsScope: h.WorkerMetricScope,
		Logger:       h.Logger,
		AutoScalerOptions: worker.AutoScalerOptions{
			Enabled:         true,
			PollerMinCount:  config.Autoscaling.PollerMinCount,
			PollerMaxCount:  config.Autoscaling.PollerMaxCount,
			PollerInitCount: config.Autoscaling.PollerInitCount,
		},
		FeatureFlags: client.FeatureFlags{
			WorkflowExecutionAlreadyCompletedErrorEnabled: true,
		},
	}

	h.Logger.Info("Starting workers with autoscaling configuration",
		zap.Bool("AutoScalerEnabled", workerOptions.AutoScalerOptions.Enabled),
		zap.Int("PollerMinCount", workerOptions.AutoScalerOptions.PollerMinCount),
		zap.Int("PollerMaxCount", workerOptions.AutoScalerOptions.PollerMaxCount),
		zap.Int("PollerInitCount", workerOptions.AutoScalerOptions.PollerInitCount))

	// Use worker.NewV2 for autoscaling support
	w, err := worker.NewV2(h.Service, h.Config.DomainName, ApplicationName, workerOptions)
	if err != nil {
		h.Logger.Fatal("Failed to create worker with autoscaling", zap.Error(err))
	}

	// Register workflows and activities
	registerWorkflowAndActivityForAutoscaling(w)

	// Start the worker
	err = w.Run()
	if err != nil {
		h.Logger.Fatal("Failed to run worker", zap.Error(err))
	}
}

// registerWorkflowAndActivityForAutoscaling registers the workflow and activities
func registerWorkflowAndActivityForAutoscaling(w worker.Worker) {
	w.RegisterWorkflowWithOptions(AutoscalingWorkflow, workflow.RegisterOptions{Name: autoscalingWorkflowName})
	w.RegisterActivityWithOptions(LoadGenerationActivity, activity.RegisterOptions{Name: loadGenerationActivityName})
}
