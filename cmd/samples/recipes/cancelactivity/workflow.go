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

/**
 * This is the cancel activity workflow sample.
 */

// ApplicationName is the task list for this sample
const ApplicationName = "cancelGroup"

// sampleCancelWorkflow workflow decider
func sampleCancelWorkflow(ctx workflow.Context) (retError error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute * 30,
		HeartbeatTimeout:       time.Second * 5,
		WaitForCancellation:    true,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("cancel workflow started")

	defer func() {
		if cadence.IsCanceledError(retError) {
			// When workflow is canceled, it has to get a new disconnected context to execute any activities
			newCtx, _ := workflow.NewDisconnectedContext(ctx)
			err := workflow.ExecuteActivity(newCtx, cleanupActivity).Get(ctx, nil)
			if err != nil {
				logger.Error("Cleanup activity failed", zap.Error(err))
				retError = err
				return
			}
			retError = nil
			logger.Info("Workflow completed.")
		}
	}()

	var result string
	err := workflow.ExecuteActivity(ctx, activityToBeCanceled).Get(ctx, &result)
	if err != nil && !cadence.IsCanceledError(err) {
		logger.Error("Error from activityToBeCanceled", zap.Error(err))
		return err
	}
	logger.Info(fmt.Sprintf("activityToBeCanceled returns %v, %v", result, err))

	// Execute activity using a canceled ctx,
	// activity won't be scheduled and an cancelled error will be returned
	err = workflow.ExecuteActivity(ctx, activityToBeSkipped).Get(ctx, nil)
	if err != nil && !cadence.IsCanceledError(err) {
		logger.Error("Error from activityToBeSkipped", zap.Error(err))
	}

	return err
}

func activityToBeCanceled(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("activity started, to cancel workflow, use ./cancelactivity -m cancel -w <WorkflowID> or CLI: 'cadence --do samples-domain wf cancel -w <WorkflowID>' to cancel")
	for {
		select {
		case <-time.After(1 * time.Second):
			logger.Info("heartbeating...")
			activity.RecordHeartbeat(ctx, "")
		case <-ctx.Done():
			logger.Info("context is cancelled")
			// returned canceled error here so that in workflow history we can see ActivityTaskCanceled event
			// or if not cancelled, return timeout error
			return "I am canceled by Done", ctx.Err()
		}
	}
}

func cleanupActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("cleanupActivity started")
	return nil
}

func activityToBeSkipped(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("this activity will be skipped due to cancellation")
	return nil
}
