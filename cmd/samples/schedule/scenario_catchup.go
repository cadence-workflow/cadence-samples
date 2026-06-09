package main

import (
	"context"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/zap"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
)

// runCatchUp demonstrates catch-up behavior on Unpause (SDK plan §2.7).
//
// A schedule fires every 3s. We pause it, let several fire times elapse while paused, then
// unpause with a catch-up policy override and observe how many missed runs the server
// replays:
//   - Skip → discards all missed fires (≈ no catch-up runs)
//   - All  → replays every missed fire within the catch-up window
//
// Run the worker first, then `-m manage -scenario catchup`. Counting is best-effort via
// Info.TotalRuns; the authoritative view is the worker logs / Web UI.
func runCatchUp(h *common.SampleHelper) {
	c := buildClient(h)
	sc := c.ScheduleClient()

	cases := []struct {
		name   string
		policy client.ScheduleCatchUpPolicy
		expect string
	}{
		{"Skip", client.ScheduleCatchUpPolicySkip, "missed fires discarded → ~0 catch-up runs on unpause"},
		{"All", client.ScheduleCatchUpPolicyAll, "every missed fire replayed on unpause (within catch-up window)"},
	}

	const pausedFor = 12 * time.Second
	for _, tc := range cases {
		ctx := context.Background()
		scheduleID := newScheduleID("sample-catchup-" + tc.name)
		h.Logger.Info("=== Catch-up: "+tc.name+" ===",
			zap.String("scheduleID", scheduleID),
			zap.String("expect", tc.expect))

		_, err := sc.Create(ctx, &client.CreateScheduleRequest{
			ScheduleID: scheduleID,
			Spec:       &client.ScheduleSpec{CronExpression: "@every 3s"},
			Action:     &client.ScheduleAction{StartWorkflow: startWorkflowAction(h, 0)},
			Policies: &client.SchedulePolicies{
				// Generous catch-up window so "All" can actually replay the missed fires.
				CatchUpWindow: time.Hour,
			},
		})
		if err != nil {
			h.Logger.Fatal("Create failed", zap.String("case", tc.name), zap.Error(err))
		}

		// Pause immediately, then let fire times accumulate while paused.
		if err = sc.Pause(ctx, scheduleID, "accumulating missed fires for catch-up demo"); err != nil {
			h.Logger.Fatal("Pause failed", zap.Error(err))
		}
		runsBefore := totalRuns(h, sc, ctx, scheduleID)
		h.Logger.Info("Paused — waiting so fire times are missed",
			zap.Duration("pausedFor", pausedFor), zap.Int64("runsBeforePause", runsBefore))
		time.Sleep(pausedFor)

		// Unpause with the per-call catch-up override.
		h.Logger.Info("Unpausing with override", zap.String("catchUpPolicy", tc.name))
		if err = sc.Unpause(ctx, scheduleID, "resume with "+tc.name, tc.policy); err != nil {
			h.Logger.Fatal("Unpause failed", zap.Error(err))
		}
		time.Sleep(6 * time.Second) // let any catch-up runs start

		runsAfter := totalRuns(h, sc, ctx, scheduleID)
		h.Logger.Info("Catch-up result (best-effort)",
			zap.Int64("runsBeforePause", runsBefore),
			zap.Int64("runsAfterUnpause", runsAfter),
			zap.Int64("delta", runsAfter-runsBefore),
			zap.String("expected", tc.expect))
		deleteQuietly(h, sc, ctx, scheduleID)
	}
	h.Logger.Info("Catch-up demo complete.")
}

// totalRuns returns Info.TotalRuns for a schedule, or 0 if unavailable.
func totalRuns(h *common.SampleHelper, sc client.ScheduleClient, ctx context.Context, id string) int64 {
	desc, err := sc.Describe(ctx, id)
	if err != nil || desc.Info == nil {
		return 0
	}
	return desc.Info.TotalRuns
}
