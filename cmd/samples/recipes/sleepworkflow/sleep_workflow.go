package main

import (
	"context"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const sleepWorkflowName = "sleepWorkflow"

// sleepWorkflow demonstrates workflow.Sleep followed by a main activity call
func sleepWorkflow(ctx workflow.Context, sleepDuration time.Duration) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Workflow started, will sleep", zap.String("duration", sleepDuration.String()))

	// Sleep for the specified duration
	err := workflow.Sleep(ctx, sleepDuration)
	if err != nil {
		logger.Error("Sleep failed", zap.Error(err))
		return err
	}

	logger.Info("Sleep finished, executing main activity")

	// Set activity options
	activityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result string
	err = workflow.ExecuteActivity(ctx, mainSleepActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Main activity failed", zap.Error(err))
		return err
	}

	logger.Info("Workflow completed", zap.String("Result", result))
	return nil
}

// mainSleepActivity is a simple activity for demonstration
func mainSleepActivity(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("mainSleepActivity executed")
	return "Main sleep activity completed", nil
}
