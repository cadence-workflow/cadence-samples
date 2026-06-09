package main

import (
	"context"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/zap"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
)

// runOverlap demonstrates ScheduleOverlapPolicy (SDK plan §2.6).
//
// Each schedule fires every 3s but its workflow stays open for 8s (sleepSeconds=8), so a
// new fire always lands while the previous run is still active — forcing the overlap policy
// to take effect. Observe the actual behavior in the worker logs or the Cadence Web UI
// (this sample logs the *expected* behavior; counting live runs is out of scope here).
//
// Run the worker (`-m worker`) in another terminal first, then `-m manage -scenario overlap`.
func runOverlap(h *common.SampleHelper) {
	c := buildClient(h)
	sc := c.ScheduleClient()

	cases := []struct {
		name        string
		policy      client.ScheduleOverlapPolicy
		concurrency int32
		expect      string
	}{
		{"SkipNew", client.ScheduleOverlapPolicySkipNew, 0,
			"new fires are SKIPPED while a run is active → roughly one run at a time"},
		{"Concurrent", client.ScheduleOverlapPolicyConcurrent, 3,
			"multiple runs execute simultaneously, capped at ConcurrencyLimit=3"},
		{"CancelPrevious", client.ScheduleOverlapPolicyCancelPrevious, 0,
			"each new fire CANCELS the still-running previous run before starting"},
	}

	const observeWindow = 18 * time.Second
	for _, tc := range cases {
		ctx := context.Background()
		scheduleID := newScheduleID("sample-overlap-" + tc.name)
		h.Logger.Info("=== Overlap: "+tc.name+" ===",
			zap.String("scheduleID", scheduleID),
			zap.String("expect", tc.expect))

		action := startWorkflowAction(h, 8) // 8s runs vs 3s cron → guaranteed overlap
		_, err := sc.Create(ctx, &client.CreateScheduleRequest{
			ScheduleID: scheduleID,
			Spec:       &client.ScheduleSpec{CronExpression: "@every 3s"},
			Action:     &client.ScheduleAction{StartWorkflow: action},
			Policies: &client.SchedulePolicies{
				OverlapPolicy:    tc.policy,
				ConcurrencyLimit: tc.concurrency,
			},
		})
		if err != nil {
			h.Logger.Fatal("Create failed", zap.String("case", tc.name), zap.Error(err))
		}

		h.Logger.Info("Letting it run — watch the worker logs / Web UI for overlap behavior",
			zap.Duration("window", observeWindow))
		time.Sleep(observeWindow)

		if desc, derr := sc.Describe(ctx, scheduleID); derr == nil && desc.Info != nil {
			h.Logger.Info("Observed (best-effort)", zap.Int64("totalRuns", desc.Info.TotalRuns))
		}
		deleteQuietly(h, sc, ctx, scheduleID)
	}
	h.Logger.Info("Overlap demo complete. Compare run timelines per policy in the Cadence Web UI.")
}
