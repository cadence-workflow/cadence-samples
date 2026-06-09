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
//
// It takes a single sleepSeconds input so the same workflow can serve every scenario:
//   - sleepSeconds == 0 → a fast run (lifecycle / catch-up / pagination demos).
//   - sleepSeconds  > 0 → a run that stays open longer than the cron interval, which is
//     what the overlap-policy demo (scenario_overlap.go) needs to force runs to overlap.
//
// The schedule passes this value via ScheduleStartWorkflowAction.Input, which the SDK
// delivers as the workflow's input. See encodeWorkflowInput in helpers.go for the
// matching encoder.
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
