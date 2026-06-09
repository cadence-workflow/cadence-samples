package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/cadence/encoded"
	"go.uber.org/zap"
)

// buildScheduleClient builds a Cadence client.Client wrapping BuildCadenceClient().
// Pass a non-nil dc to use a custom DataConverter (e.g. for the dataconverter scenario).
func buildScheduleClient(dc encoded.DataConverter) client.Client {
	var opts *client.Options
	if dc != nil {
		opts = &client.Options{DataConverter: dc}
	}
	return client.NewClient(BuildCadenceClient(), Domain, opts)
}

// newScheduleID returns a process-unique schedule ID so reruns never collide.
func newScheduleID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// encodeWorkflowInput encodes the scheduled workflow's input via the default DataConverter.
// The resulting bytes go into ScheduleStartWorkflowAction.Input (the SDK passes them verbatim).
func encodeWorkflowInput(logger *zap.Logger, sleepSeconds int) []byte {
	data, err := encoded.GetDefaultDataConverter().ToData(sleepSeconds)
	if err != nil {
		logger.Fatal("Failed to encode workflow input", zap.Error(err))
	}
	return data
}

// startWorkflowAction builds the common StartWorkflow action used by every scenario.
func startWorkflowAction(logger *zap.Logger, sleepSeconds int) *client.ScheduleStartWorkflowAction {
	return &client.ScheduleStartWorkflowAction{
		WorkflowType:                    scheduledWorkflowName,
		TaskList:                        TaskListName,
		Input:                           encodeWorkflowInput(logger, sleepSeconds),
		ExecutionStartToCloseTimeout:    60 * time.Second,
		DecisionTaskStartToCloseTimeout: 10 * time.Second,
	}
}

// deleteQuietly best-effort deletes a schedule and logs (but does not fail) on error.
func deleteQuietly(logger *zap.Logger, sc client.ScheduleClient, ctx context.Context, scheduleID string) {
	if err := sc.Delete(ctx, scheduleID); err != nil {
		logger.Warn("cleanup: delete failed", zap.String("scheduleID", scheduleID), zap.Error(err))
		return
	}
	logger.Info("cleanup: schedule deleted", zap.String("scheduleID", scheduleID))
}
