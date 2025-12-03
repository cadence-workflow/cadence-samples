package main

import (
	"context"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// SleepWorkflow demonstrates workflow.Sleep for pausing execution.
func SleepWorkflow(ctx workflow.Context, sleepDuration time.Duration) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("SleepWorkflow started", zap.Duration("sleepDuration", sleepDuration))

	// Sleep for the specified duration
	err := workflow.Sleep(ctx, sleepDuration)
	if err != nil {
		logger.Error("Sleep failed", zap.Error(err))
		return err
	}

	logger.Info("Sleep finished, executing activity")

	// Set activity options
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result string
	err = workflow.ExecuteActivity(ctx, MainSleepActivity).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed", zap.Error(err))
		return err
	}

	logger.Info("SleepWorkflow completed", zap.String("result", result))
	return nil
}

// MainSleepActivity is executed after the sleep completes.
func MainSleepActivity(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("MainSleepActivity executed")
	return "Activity completed after sleep", nil
}

