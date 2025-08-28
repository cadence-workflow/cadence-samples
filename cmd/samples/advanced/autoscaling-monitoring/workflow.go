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
func AutoscalingWorkflow(ctx workflow.Context, activitiesPerWorkflow int, batchDelay int, minProcessingTime, maxProcessingTime int) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Autoscaling workflow started", zap.Int("activitiesPerWorkflow", activitiesPerWorkflow))

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute * 20,
		StartToCloseTimeout:    time.Minute * 20,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// Generate load by executing activities in parallel
	var futures []workflow.Future

	// Execute activities in batches to create varying load
	for i := 0; i < activitiesPerWorkflow; i++ {
		future := workflow.ExecuteActivity(ctx, LoadGenerationActivity, i, minProcessingTime, maxProcessingTime)
		futures = append(futures, future)

		// Add some delay between batches to simulate real-world patterns
		// Use batch delay from configuration
		if i > 0 && i % 10 == 0 {
			workflow.Sleep(ctx, time.Duration(batchDelay)*time.Millisecond)
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
