package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/cadence/encoded"
	"go.uber.org/zap"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
)

// buildClient builds a Cadence client from the sample helper, failing fast on error.
func buildClient(h *common.SampleHelper) client.Client {
	c, err := h.Builder.BuildCadenceClient()
	if err != nil {
		h.Logger.Fatal("Failed to build cadence client", zap.Error(err))
	}
	return c
}

// newScheduleID returns a process-unique schedule ID so reruns never collide.
func newScheduleID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// encodeWorkflowInput encodes the scheduled workflow's input (sleepSeconds) exactly the
// way the worker decodes it — through the default DataConverter. The resulting bytes go
// into ScheduleStartWorkflowAction.Input, which the SDK passes through verbatim (it does
// not encode Input for you; see the field doc on ScheduleStartWorkflowAction).
func encodeWorkflowInput(h *common.SampleHelper, sleepSeconds int) []byte {
	data, err := encoded.GetDefaultDataConverter().ToData(sleepSeconds)
	if err != nil {
		h.Logger.Fatal("Failed to encode workflow input", zap.Error(err))
	}
	return data
}

// startWorkflowAction builds the common StartWorkflow action used by every scenario.
// sleepSeconds controls how long each triggered run stays open (see scheduledWorkflow).
func startWorkflowAction(h *common.SampleHelper, sleepSeconds int) *client.ScheduleStartWorkflowAction {
	return &client.ScheduleStartWorkflowAction{
		WorkflowType:                    scheduledWorkflowName,
		TaskList:                        ApplicationName,
		Input:                           encodeWorkflowInput(h, sleepSeconds),
		ExecutionStartToCloseTimeout:    60 * time.Second,
		DecisionTaskStartToCloseTimeout: 10 * time.Second,
	}
}

// deleteQuietly best-effort deletes a schedule and logs (but does not fail) on error.
// Used in defer cleanup so a scenario always tidies up after itself.
func deleteQuietly(h *common.SampleHelper, sc client.ScheduleClient, ctx context.Context, scheduleID string) {
	if err := sc.Delete(ctx, scheduleID); err != nil {
		h.Logger.Warn("cleanup: delete failed", zap.String("scheduleID", scheduleID), zap.Error(err))
		return
	}
	h.Logger.Info("cleanup: schedule deleted", zap.String("scheduleID", scheduleID))
}
