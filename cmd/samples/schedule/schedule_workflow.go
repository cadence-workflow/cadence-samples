package main

import (
	"context"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	// ApplicationName is the task list shared by the worker and the scheduled workflow.
	ApplicationName = "scheduleGroup"

	// scheduledWorkflowName is the registered name used in CreateScheduleRequest.Action.
	scheduledWorkflowName = "scheduledWorkflow"
)

// scheduledWorkflow is triggered by the schedule on each cron tick.
func scheduledWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("Scheduled workflow started")
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
