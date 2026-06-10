package main

import (
	"context"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/zap"
)

// runOverlap is a live demonstration of ScheduleOverlapPolicy.
//
// Each schedule fires every 3 seconds but its workflow sleeps for 8 seconds, so a new
// fire always lands while the previous run is still active — forcing the policy to act.
//
// This scenario is observational: watch the worker terminal or Cadence Web UI to see
// the policies take effect in real time. For automated verification of overlap policy
// correctness, see scenario_backfill.go which tests the same three policies with
// deterministic, countable results.
//
// Run the worker (`-m worker`) in another terminal first, then `-m manage -scenario overlap`.
func runOverlap() {
	logger := BuildLogger()
	c := buildScheduleClient(nil)
	sc := c.ScheduleClient()

	const fireWindow = 15 * time.Second

	cases := []struct {
		name        string
		policy      client.ScheduleOverlapPolicy
		concurrency int32
		workerHint  string
	}{
		{
			name:        "SkipNew",
			policy:      client.ScheduleOverlapPolicySkipNew,
			concurrency: 0,
			workerHint:  "watch for gaps between 'Scheduled workflow started' — most ticks are silently dropped while a run is open",
		},
		{
			name:        "Concurrent",
			policy:      client.ScheduleOverlapPolicyConcurrent,
			concurrency: 3,
			workerHint:  "watch for multiple simultaneous 'Scheduled workflow started'; at most 3 run at once (ConcurrencyLimit=3)",
		},
		{
			name:        "CancelPrevious",
			policy:      client.ScheduleOverlapPolicyCancelPrevious,
			concurrency: 0,
			workerHint:  "watch for 'workflow cancelled' immediately before each new 'Scheduled workflow started'",
		},
	}

	for _, tc := range cases {
		tc := tc
		func() {
			ctx := context.Background()
			scheduleID := newScheduleID("sample-overlap-" + tc.name)
			logger.Info("=== Overlap: "+tc.name+" ===",
				zap.String("scheduleID", scheduleID),
				zap.Duration("fireWindow", fireWindow),
				zap.String("workerHint", tc.workerHint))
			defer deleteQuietly(logger, sc, context.Background(), scheduleID)

			action := startWorkflowAction(logger, 8) // 8s run > 3s cron → guaranteed overlap
			action.WorkflowIDPrefix = "overlap-" + tc.name

			if _, err := sc.Create(ctx, &client.CreateScheduleRequest{
				ScheduleID: scheduleID,
				Spec:       &client.ScheduleSpec{CronExpression: "@every 3s"},
				Action:     &client.ScheduleAction{StartWorkflow: action},
				Policies: &client.SchedulePolicies{
					OverlapPolicy:    tc.policy,
					ConcurrencyLimit: tc.concurrency,
				},
			}); err != nil {
				logger.Fatal("Create failed", zap.String("case", tc.name), zap.Error(err))
			}

			time.Sleep(fireWindow)

			// Pause before reading to avoid racing with a live fire.
			if err := sc.Pause(ctx, scheduleID, "end of observation window"); err != nil {
				logger.Warn("Pause failed", zap.Error(err))
			}
			time.Sleep(2 * time.Second) // let in-flight starts register

			if desc, err := sc.Describe(ctx, scheduleID); err == nil && desc.Info != nil {
				logger.Info("Observed TotalRuns (informational — exact count varies with timing)",
					zap.String("policy", tc.name),
					zap.Int64("totalRuns", desc.Info.TotalRuns))
			}
		}()
	}

	logger.Info("Overlap demo complete. Compare run timelines per policy in the Cadence Web UI.")
}
