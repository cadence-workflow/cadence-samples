package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// CancelWorkflow demonstrates cancellation handling in Cadence workflows.
// It shows how to:
// - Handle workflow cancellation
// - Use WaitForCancellation for graceful activity shutdown
// - Run cleanup activities using disconnected context
func CancelWorkflow(ctx workflow.Context) (retError error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute * 30,
		HeartbeatTimeout:       time.Second * 5,
		WaitForCancellation:    true,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("CancelWorkflow started")

	defer func() {
		if cadence.IsCanceledError(retError) {
			// When workflow is canceled, it has to get a new disconnected context to execute any activities
			newCtx, _ := workflow.NewDisconnectedContext(ctx)
			err := workflow.ExecuteActivity(newCtx, CleanupActivity).Get(ctx, nil)
			if err != nil {
				logger.Error("Cleanup activity failed", zap.Error(err))
				retError = err
				return
			}
			retError = nil
			logger.Info("Workflow completed with cleanup.")
		}
	}()

	var result string
	err := workflow.ExecuteActivity(ctx, ActivityToBeCanceled).Get(ctx, &result)
	if err != nil && !cadence.IsCanceledError(err) {
		logger.Error("Error from ActivityToBeCanceled", zap.Error(err))
		return err
	}
	logger.Info(fmt.Sprintf("ActivityToBeCanceled returns %v, %v", result, err))

	// Execute activity using a canceled ctx,
	// activity won't be scheduled and a canceled error will be returned
	err = workflow.ExecuteActivity(ctx, ActivityToBeSkipped).Get(ctx, nil)
	if err != nil && !cadence.IsCanceledError(err) {
		logger.Error("Error from ActivityToBeSkipped", zap.Error(err))
	}

	return err
}

// ActivityToBeCanceled is an activity that heartbeats until canceled.
// To cancel: use CLI 'cadence --domain cadence-samples workflow cancel --wid <WorkflowID>'
func ActivityToBeCanceled(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ActivityToBeCanceled started - waiting for cancellation")
	logger.Info("To cancel: cadence --env development --domain cadence-samples workflow cancel --wid <WorkflowID>")

	for {
		select {
		case <-time.After(1 * time.Second):
			logger.Info("heartbeating...")
			activity.RecordHeartbeat(ctx, "")
		case <-ctx.Done():
			logger.Info("context is cancelled")
			return "I am canceled by Done", ctx.Err()
		}
	}
}

// CleanupActivity runs after workflow cancellation to perform cleanup.
func CleanupActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("CleanupActivity started - performing cleanup after cancellation")
	return nil
}

// ActivityToBeSkipped demonstrates that activities are skipped when workflow is canceled.
func ActivityToBeSkipped(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("ActivityToBeSkipped - this should not run if workflow is canceled")
	return nil
}

