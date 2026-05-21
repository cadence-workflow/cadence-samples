package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/cadence/worker"
	"go.uber.org/zap"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
)

func registerWorkflowAndActivity(h *common.SampleHelper) {
	h.RegisterWorkflowWithAlias(scheduledWorkflow, scheduledWorkflowName)
	h.RegisterActivity(scheduledActivity)
}

func startWorkers(h *common.SampleHelper) {
	workerOptions := worker.Options{
		MetricsScope: h.WorkerMetricScope,
		Logger:       h.Logger,
		FeatureFlags: client.FeatureFlags{
			WorkflowExecutionAlreadyCompletedErrorEnabled: true,
		},
	}
	h.StartWorkers(h.Config.DomainName, ApplicationName, workerOptions)
}

func manageSchedule(h *common.SampleHelper) {
	cadenceClient, err := h.Builder.BuildCadenceClient()
	if err != nil {
		h.Logger.Fatal("Failed to build cadence client", zap.Error(err))
	}

	sc := cadenceClient.ScheduleClient()
	ctx := context.Background()
	scheduleID := fmt.Sprintf("sample-schedule-%d", time.Now().UnixNano())

	h.Logger.Info("=== Step 1: Create schedule ===", zap.String("scheduleID", scheduleID))
	_, err = sc.Create(ctx, &client.CreateScheduleRequest{
		ScheduleID: scheduleID,
		Spec: &client.ScheduleSpec{
			CronExpression: "0 * * * *",
		},
		Action: &client.ScheduleAction{
			StartWorkflow: &client.ScheduleStartWorkflowAction{
				WorkflowType:                    scheduledWorkflowName,
				TaskList:                        ApplicationName,
				ExecutionStartToCloseTimeout:    15 * time.Second,
				DecisionTaskStartToCloseTimeout: 10 * time.Second,
			},
		},
	})
	if err != nil {
		h.Logger.Fatal("Create failed", zap.Error(err))
	}
	h.Logger.Info("Schedule created")

	h.Logger.Info("=== Step 2: Describe schedule ===")
	desc, err := sc.Describe(ctx, scheduleID)
	if err != nil {
		h.Logger.Fatal("Describe failed", zap.Error(err))
	}
	if desc.Spec != nil {
		h.Logger.Info("Describe result - spec", zap.String("cron", desc.Spec.CronExpression))
	}
	if desc.Action != nil && desc.Action.StartWorkflow != nil {
		h.Logger.Info("Describe result - action", zap.String("workflowType", desc.Action.StartWorkflow.WorkflowType))
	}
	if desc.State != nil {
		h.Logger.Info("Describe result - state", zap.Bool("paused", desc.State.Paused))
	}

	h.Logger.Info("=== Step 3: Update schedule (change cron to every 2 hours) ===")
	err = sc.Update(ctx, &client.UpdateScheduleRequest{
		ScheduleID: scheduleID,
		Spec:       &client.ScheduleSpec{CronExpression: "0 */2 * * *"},
	})
	if err != nil {
		h.Logger.Fatal("Update failed", zap.Error(err))
	}
	desc, err = sc.Describe(ctx, scheduleID)
	if err != nil {
		h.Logger.Fatal("Describe after Update failed", zap.Error(err))
	}
	if desc.Spec != nil {
		h.Logger.Info("Update result", zap.String("newCron", desc.Spec.CronExpression))
	}

	h.Logger.Info("=== Step 4: Pause schedule ===")
	if err = sc.Pause(ctx, scheduleID, "manual pause for demo"); err != nil {
		h.Logger.Fatal("Pause failed", zap.Error(err))
	}
	desc, err = sc.Describe(ctx, scheduleID)
	if err != nil {
		h.Logger.Fatal("Describe after Pause failed", zap.Error(err))
	}
	if desc.State != nil {
		var reason string
		if desc.State.PauseInfo != nil {
			reason = desc.State.PauseInfo.Reason
		}
		h.Logger.Info("Pause result",
			zap.Bool("paused", desc.State.Paused),
			zap.String("reason", reason),
		)
	}

	h.Logger.Info("=== Step 5: Unpause schedule ===")
	if err = sc.Unpause(ctx, scheduleID, "resuming after demo pause", client.ScheduleCatchUpPolicySkip); err != nil {
		h.Logger.Fatal("Unpause failed", zap.Error(err))
	}
	desc, err = sc.Describe(ctx, scheduleID)
	if err != nil {
		h.Logger.Fatal("Describe after Unpause failed", zap.Error(err))
	}
	if desc.State != nil {
		h.Logger.Info("Unpause result", zap.Bool("paused", desc.State.Paused))
	}

	h.Logger.Info("=== Step 6: Backfill past 3 hours (triggers one run) ===")
	now := time.Now()
	err = sc.Backfill(ctx, scheduleID, &client.BackfillRequest{
		StartTime:     now.Add(-4 * time.Hour),
		EndTime:       now.Add(-1 * time.Hour),
		OverlapPolicy: client.ScheduleOverlapPolicySkipNew,
	})
	if err != nil {
		h.Logger.Fatal("Backfill failed", zap.Error(err))
	}
	h.Logger.Info("Backfill submitted — waiting for worker to pick up the run...")
	time.Sleep(3 * time.Second)

	h.Logger.Info("=== Step 7: List schedules in domain ===")
	listResp, err := sc.List(ctx, 100, nil)
	if err != nil {
		h.Logger.Fatal("List failed", zap.Error(err))
	}
	found := false
	for _, s := range listResp.Schedules {
		h.Logger.Info("Listed schedule",
			zap.String("scheduleID", s.ScheduleID),
			zap.String("workflowType", s.WorkflowType),
			zap.Bool("paused", s.State != nil && s.State.Paused),
		)
		if s.ScheduleID == scheduleID {
			found = true
		}
	}
	if !found {
		h.Logger.Warn("Our schedule was not found in List (indexing may be delayed)")
	}

	h.Logger.Info("=== Step 8: Delete schedule ===")
	if err = sc.Delete(ctx, scheduleID); err != nil {
		h.Logger.Fatal("Delete failed", zap.Error(err))
	}
	h.Logger.Info("Schedule deleted", zap.String("scheduleID", scheduleID))
}

func main() {
	var mode string
	flag.StringVar(&mode, "m", "manage", "Mode: worker | manage")
	flag.Parse()

	var h common.SampleHelper
	h.SetupServiceConfig()

	switch mode {
	case "worker":
		registerWorkflowAndActivity(&h)
		startWorkers(&h)
		// The workers are supposed to be long running process that should not exit.
		// Use select{} to block indefinitely for samples, you can quit by CMD+C.
		select {}
	case "manage":
		manageSchedule(&h)
	default:
		h.Logger.Fatal("Unknown mode", zap.String("mode", mode), zap.String("valid", "worker|manage"))
	}
}
