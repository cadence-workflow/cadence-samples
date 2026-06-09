package main

import (
	"context"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	scheduledWorkflowName = "scheduledWorkflow"
)

// scheduledWorkflow is triggered by the schedule on each cron tick.
//
// sleepSeconds == 0 → fast run; sleepSeconds > 0 → run stays open longer than the cron
// interval, which the overlap-policy scenario needs to force runs to overlap.
func scheduledWorkflow(ctx workflow.Context, sleepSeconds int) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Scheduled workflow started", zap.Int("sleepSeconds", sleepSeconds))

	if sleepSeconds > 0 {
		if err := workflow.Sleep(ctx, time.Duration(sleepSeconds)*time.Second); err != nil {
			return err
		}
	}

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	if err := workflow.ExecuteActivity(ctx, scheduledActivity).Get(ctx, nil); err != nil {
		logger.Error("Activity failed", zap.Error(err))
		return err
	}
	logger.Info("Scheduled workflow completed")
	return nil
}

// scheduledActivity is the unit of work executed on each schedule trigger.
func scheduledActivity(ctx context.Context) error {
	activity.GetLogger(ctx).Info("Scheduled activity executed", zap.Time("at", time.Now()))
	return nil
}
