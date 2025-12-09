package main

import (
	"context"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// DynamicGreetingActivityName is the registered name for the activity.
// This demonstrates how to invoke activities by string name rather than function reference.
const DynamicGreetingActivityName = "cadence_samples.DynamicGreetingActivity"

type dynamicWorkflowInput struct {
	Message string `json:"message"`
}

// DynamicWorkflow demonstrates calling activities using string names for dynamic behavior.
// Instead of passing the function directly to ExecuteActivity, we pass the activity name.
// This is useful for plugin systems or configuration-driven workflows.
func DynamicWorkflow(ctx workflow.Context, input dynamicWorkflowInput) (string, error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("DynamicWorkflow started")

	var greetingMsg string
	// Note: We pass the activity NAME (string) instead of the function reference
	err := workflow.ExecuteActivity(ctx, DynamicGreetingActivityName, input.Message).Get(ctx, &greetingMsg)
	if err != nil {
		logger.Error("DynamicGreetingActivity failed", zap.Error(err))
		return "", err
	}

	logger.Info("Workflow result", zap.String("greeting", greetingMsg))
	return greetingMsg, nil
}

// DynamicGreetingActivity is a simple activity that returns a greeting message.
func DynamicGreetingActivity(ctx context.Context, message string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("DynamicGreetingActivity started.")
	return "Hello, " + message, nil
}

