package main

import (
	"context"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// TracingWorkflow demonstrates distributed tracing in Cadence.
// Trace context is automatically propagated through workflow and activity execution.
func TracingWorkflow(ctx workflow.Context, name string) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("TracingWorkflow started")

	var result string
	err := workflow.ExecuteActivity(ctx, TracingActivity, name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", zap.Error(err))
		return err
	}

	logger.Info("Workflow completed.", zap.String("Result", result))
	return nil
}

// TracingActivity is a simple activity that returns a greeting.
func TracingActivity(ctx context.Context, name string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("TracingActivity started")
	return "Hello " + name + "!", nil
}

