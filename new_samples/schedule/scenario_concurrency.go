package main

import (
	"context"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/zap"
)

// runConcurrency verifies that ConcurrencyLimit bounds the number of simultaneously
// running workflow executions triggered by a schedule.
//
// Setup: schedule fires every 3 seconds; each workflow sleeps for 20 seconds.
// The observation window (12s) is shorter than the workflow duration, so no run can
// complete before we read TotalRuns. After 12s with ConcurrencyLimit=2, TotalRuns
// must equal exactly 2 — any higher value means the limit was not enforced.
//
// No backfill needed: ConcurrencyLimit is server-enforced regardless of whether
// fires come from live cron, backfill, or manual triggers.
//
// Run the worker (`-m worker`) in another terminal first, then `-m manage -scenario concurrency`.
func runConcurrency() {
	logger := BuildLogger()
	c := buildScheduleClient(nil)
	sc := c.ScheduleClient()

	const (
		concurrencyLimit  = int32(2)
		workflowSleepSecs = 20            // longer than observeWindow — no run completes during the check
		observeWindow     = 12 * time.Second // @every 3s → 4 fires; only 2 should start
	)

	ctx := context.Background()
	scheduleID := newScheduleID("sample-concurrency")
	logger.Info("=== ConcurrencyLimit ===",
		zap.String("scheduleID", scheduleID),
		zap.Int32("concurrencyLimit", concurrencyLimit),
		zap.Int("workflowSleepSecs", workflowSleepSecs),
		zap.Duration("observeWindow", observeWindow))
	defer deleteQuietly(logger, sc, context.Background(), scheduleID)

	action := startWorkflowAction(logger, workflowSleepSecs)
	action.WorkflowIDPrefix = "concurrency"

	if _, err := sc.Create(ctx, &client.CreateScheduleRequest{
		ScheduleID: scheduleID,
		Spec:       &client.ScheduleSpec{CronExpression: "@every 3s"},
		Action:     &client.ScheduleAction{StartWorkflow: action},
		Policies: &client.SchedulePolicies{
			OverlapPolicy:    client.ScheduleOverlapPolicyConcurrent,
			ConcurrencyLimit: concurrencyLimit,
		},
	}); err != nil {
		logger.Fatal("Create failed", zap.Error(err))
	}

	logger.Info("Letting schedule fire",
		zap.Duration("window", observeWindow),
		zap.String("hint", "watch worker for 'Scheduled workflow started' — expect exactly 2"))
	time.Sleep(observeWindow)

	// Pause before reading to prevent a concurrent fire from racing with Describe.
	if err := sc.Pause(ctx, scheduleID, "reading TotalRuns"); err != nil {
		logger.Warn("Pause failed — TotalRuns read may race with a live fire", zap.Error(err))
	}
	time.Sleep(time.Second) // let any in-flight start register

	runs := totalRuns(sc, ctx, scheduleID)
	switch {
	case runs == int64(concurrencyLimit):
		logger.Info("  MATCH   ConcurrencyLimit enforced: exactly limit runs started, rest were held back",
			zap.Int64("totalRuns", runs),
			zap.Int32("concurrencyLimit", concurrencyLimit))
	case runs < int64(concurrencyLimit):
		logger.Warn("  fewer runs than ConcurrencyLimit — schedule may not have fired yet",
			zap.Int64("totalRuns", runs),
			zap.Int32("concurrencyLimit", concurrencyLimit),
			zap.String("hint", "ensure the worker is running"))
	default:
		logger.Warn("  MISMATCH more than ConcurrencyLimit runs started simultaneously",
			zap.Int64("totalRuns", runs),
			zap.Int32("concurrencyLimit", concurrencyLimit))
	}
}
