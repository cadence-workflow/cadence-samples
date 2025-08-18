package main

import (
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	autoscalingWorkflowName = "autoscalingWorkflow"
)

// AutoscalingWorkflow demonstrates a workflow that can generate load
// to test worker poller autoscaling
func AutoscalingWorkflow(ctx workflow.Context, iterations int) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Autoscaling workflow started", zap.Int("iterations", iterations))

	// Record workflow start metrics
	RecordWorkflowStarted("autoscaling-worker-1")

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Generate load by executing activities in parallel
	var futures []workflow.Future

	// Execute activities in batches to create varying load
	for i := 0; i < iterations; i++ {
		future := workflow.ExecuteActivity(ctx, LoadGenerationActivity, i)
		futures = append(futures, future)

		// Add some delay between batches to simulate real-world patterns
		// Use batch delay from configuration
		if i > 0 && i%10 == 0 {
			batchDelay := time.Duration(config.Autoscaling.LoadGeneration.BatchDelay) * time.Second
			workflow.Sleep(ctx, batchDelay)
		}
	}

	// Wait for all activities to complete
	for i, future := range futures {
		var result error
		if err := future.Get(ctx, &result); err != nil {
			logger.Error("Activity failed", zap.Int("taskID", i), zap.Error(err))
			return err
		}
	}

	logger.Info("Autoscaling workflow completed", zap.Int("totalActivities", len(futures)))
	return nil
}
