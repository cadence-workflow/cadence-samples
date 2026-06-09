package main

import (
	"context"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/zap"
)

// runCatchUp demonstrates catch-up behavior on Unpause.
//
// A schedule fires every 3s. We pause it, let several fire times elapse while paused, then
// unpause with a catch-up policy override and observe how many missed runs the server replays:
//   - Skip → discards all missed fires
//   - All  → replays every missed fire within the catch-up window
//
// Run the worker first, then `-m manage -scenario catchup`.
func runCatchUp() {
	logger := BuildLogger()
	c := buildScheduleClient(nil)
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
		logger.Info("=== Catch-up: "+tc.name+" ===",
			zap.String("scheduleID", scheduleID),
			zap.String("expect", tc.expect))

		_, err := sc.Create(ctx, &client.CreateScheduleRequest{
			ScheduleID: scheduleID,
			Spec:       &client.ScheduleSpec{CronExpression: "@every 3s"},
			Action:     &client.ScheduleAction{StartWorkflow: startWorkflowAction(logger, 0)},
			Policies: &client.SchedulePolicies{
				CatchUpWindow: time.Hour,
			},
		})
		if err != nil {
			logger.Fatal("Create failed", zap.String("case", tc.name), zap.Error(err))
		}

		if err = sc.Pause(ctx, scheduleID, "accumulating missed fires for catch-up demo"); err != nil {
			logger.Fatal("Pause failed", zap.Error(err))
		}
		runsBefore := totalRuns(sc, ctx, scheduleID)
		logger.Info("Paused — waiting so fire times are missed",
			zap.Duration("pausedFor", pausedFor), zap.Int64("runsBeforePause", runsBefore))
		time.Sleep(pausedFor)

		logger.Info("Unpausing with override", zap.String("catchUpPolicy", tc.name))
		if err = sc.Unpause(ctx, scheduleID, "resume with "+tc.name, tc.policy); err != nil {
			logger.Fatal("Unpause failed", zap.Error(err))
		}
		time.Sleep(6 * time.Second)

		runsAfter := totalRuns(sc, ctx, scheduleID)
		logger.Info("Catch-up result (best-effort)",
			zap.Int64("runsBeforePause", runsBefore),
			zap.Int64("runsAfterUnpause", runsAfter),
			zap.Int64("delta", runsAfter-runsBefore),
			zap.String("expected", tc.expect))
		deleteQuietly(logger, sc, ctx, scheduleID)
	}
	logger.Info("Catch-up demo complete.")
}

func totalRuns(sc client.ScheduleClient, ctx context.Context, id string) int64 {
	desc, err := sc.Describe(ctx, id)
	if err != nil || desc.Info == nil {
		return 0
	}
	return desc.Info.TotalRuns
}
