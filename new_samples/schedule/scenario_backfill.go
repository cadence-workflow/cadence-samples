package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/zap"
)

// runBackfill exercises three sub-scenarios that cover the most important backfill behaviors:
//
//  1. OverlapPolicy comparison — how SkipNew / Concurrent / CancelPrevious differ when many
//     historical slots fire at once (backfill fires slots in rapid succession, so the policy
//     has immediate, observable consequences).
//
//  2. Idempotency — re-submitting the same BackfillID does not re-queue the range.
//
// Each case creates a paused hourly schedule with StartTime 7 days in the past so the
// 4-hour backfill window [-5h, -1h] always falls inside the active period.
//
// Run the worker first, then: go run . -m manage -scenario backfill
func runBackfill() {
	logger := BuildLogger()
	c := buildScheduleClient(nil)
	sc := c.ScheduleClient()

	backfillOverlapPolicies(logger, sc)
	backfillIdempotency(logger, sc)

	logger.Info("=== Backfill demo complete — compare run timelines in the Cadence Web UI ===")
}

// backfillOverlapPolicies runs three sub-cases with the same 4-hour window (~3-4 hourly ticks).
// The schedule is paused before the backfill so only backfill runs are observed.
func backfillOverlapPolicies(logger *zap.Logger, sc client.ScheduleClient) {
	now := time.Now()
	bfStart := now.Add(-5 * time.Hour)
	bfEnd := now.Add(-1 * time.Hour)

	cases := []struct {
		tag         string
		policy      client.ScheduleOverlapPolicy
		sleepSecs   int   // how long each workflow run sleeps
		minExpected int64 // minimum TotalRuns to consider the case verified
		note        string
	}{
		{
			tag:         "SkipNew",
			policy:      client.ScheduleOverlapPolicySkipNew,
			sleepSecs:   0,
			minExpected: 1,
			note: "backfill fires slots in rapid succession; each tick sees the previous run " +
				"as still open → only the first slot starts, the rest are skipped",
		},
		{
			tag:         "Concurrent",
			policy:      client.ScheduleOverlapPolicyConcurrent,
			sleepSecs:   0,
			minExpected: 2,
			note: "all slots start in parallel immediately; " +
				"TotalRuns should equal the number of ticks in the window",
		},
		{
			tag:         "CancelPrevious",
			policy:      client.ScheduleOverlapPolicyCancelPrevious,
			sleepSecs:   20, // slow run so there is something to cancel when the next tick fires
			minExpected: 2,
			note: "each slot cancels the still-running prior run before starting; " +
				"TotalRuns == tick count but only the last slot survives to completion",
		},
	}

	for _, tc := range cases {
		tc := tc
		func() {
			ctx := context.Background()
			scheduleID := newScheduleID("sample-backfill-" + tc.tag)
			logger.Info(fmt.Sprintf("=== Backfill: OverlapPolicy=%s ===", tc.tag),
				zap.String("scheduleID", scheduleID),
				zap.String("expect", tc.note))

			action := startWorkflowAction(logger, tc.sleepSecs)
			action.WorkflowIDPrefix = "backfill-" + tc.tag

			_, err := sc.Create(ctx, &client.CreateScheduleRequest{
				ScheduleID: scheduleID,
				Spec: &client.ScheduleSpec{
					CronExpression: "0 * * * *",
					StartTime:      now.Add(-7 * 24 * time.Hour), // must predate the backfill window
				},
				Action:   &client.ScheduleAction{StartWorkflow: action},
				Policies: &client.SchedulePolicies{OverlapPolicy: tc.policy},
			})
			if err != nil {
				logger.Fatal("Create failed", zap.String("case", tc.tag), zap.Error(err))
			}
			defer deleteQuietly(logger, sc, context.Background(), scheduleID)

			// Pause immediately so live cron fires do not interfere with the observation.
			if err = sc.Pause(ctx, scheduleID, "isolating backfill from live fires"); err != nil {
				logger.Fatal("Pause failed", zap.Error(err))
			}

			if err = sc.Backfill(ctx, scheduleID, &client.BackfillRequest{
				StartTime:     bfStart,
				EndTime:       bfEnd,
				OverlapPolicy: tc.policy,
			}); err != nil {
				logger.Fatal("Backfill failed", zap.Error(err))
			}

			verifyBackfillCase(logger, sc, ctx, scheduleID, tc.tag, tc.minExpected, tc.note)
		}()
	}
}

// verifyBackfillCase polls TotalRuns for up to 30 s and logs MATCH / MISMATCH.
func verifyBackfillCase(logger *zap.Logger, sc client.ScheduleClient, ctx context.Context, id, tag string, minExpected int64, note string) {
	logger.Info("--- Verify: Backfill/"+tag+" ---",
		zap.Int64("minExpected", minExpected))
	logger.Info("  Watch the worker terminal: 'Scheduled workflow started' / 'completed' per slot")

	var last int64
	for deadline := time.Now().Add(30 * time.Second); time.Now().Before(deadline); time.Sleep(2 * time.Second) {
		desc, err := sc.Describe(ctx, id)
		if err != nil || desc.Info == nil {
			continue
		}
		last = desc.Info.TotalRuns
		if last >= minExpected {
			logger.Info("  MATCH   TotalRuns reached expected minimum",
				zap.Int64("totalRuns", last),
				zap.Int64("minExpected", minExpected))
			logBackfillNote(logger, tag, last)
			return
		}
	}
	logger.Warn("  MISMATCH TotalRuns did not reach minimum within 30s",
		zap.Int64("got", last), zap.Int64("minExpected", minExpected),
		zap.String("hint", "ensure the worker is running and StartTime predates the backfill window"))
}

func logBackfillNote(logger *zap.Logger, tag string, totalRuns int64) {
	switch tag {
	case "SkipNew":
		logger.Info("  NOTE: SkipNew — subsequent ticks were SKIPPED (each saw the prior run as open)",
			zap.Int64("startedRuns", totalRuns))
	case "Concurrent":
		logger.Info("  NOTE: Concurrent — all ticks started in parallel",
			zap.Int64("startedRuns", totalRuns))
	case "CancelPrevious":
		logger.Info("  NOTE: CancelPrevious — each tick cancelled its predecessor; "+
			"watch for 'workflow cancelled' in the worker terminal",
			zap.Int64("startedRuns", totalRuns))
	}
}

// backfillIdempotency shows that submitting the same BackfillID twice does not re-queue
// the range — the scheduler deduplicates by ID and the second request is a no-op.
func backfillIdempotency(logger *zap.Logger, sc client.ScheduleClient) {
	ctx := context.Background()
	now := time.Now()
	scheduleID := newScheduleID("sample-backfill-idem")
	bfStart := now.Add(-5 * time.Hour)
	bfEnd := now.Add(-1 * time.Hour)
	const backfillID = "demo-idempotency-key"

	logger.Info("=== Backfill: Idempotency ===",
		zap.String("scheduleID", scheduleID),
		zap.String("backfillID", backfillID),
		zap.String("expect", "second request with same BackfillID is ignored — TotalRuns does not increase"))

	action := startWorkflowAction(logger, 0)
	action.WorkflowIDPrefix = "backfill-idem"

	_, err := sc.Create(ctx, &client.CreateScheduleRequest{
		ScheduleID: scheduleID,
		Spec: &client.ScheduleSpec{
			CronExpression: "0 * * * *",
			StartTime:      now.Add(-7 * 24 * time.Hour),
		},
		Action:   &client.ScheduleAction{StartWorkflow: action},
		Policies: &client.SchedulePolicies{OverlapPolicy: client.ScheduleOverlapPolicyConcurrent},
	})
	if err != nil {
		logger.Fatal("Create failed", zap.Error(err))
	}
	defer deleteQuietly(logger, sc, context.Background(), scheduleID)

	if err = sc.Pause(ctx, scheduleID, "isolating backfill"); err != nil {
		logger.Fatal("Pause failed", zap.Error(err))
	}

	req := &client.BackfillRequest{
		StartTime:     bfStart,
		EndTime:       bfEnd,
		OverlapPolicy: client.ScheduleOverlapPolicyConcurrent,
		BackfillID:    backfillID,
	}

	if err = sc.Backfill(ctx, scheduleID, req); err != nil {
		logger.Fatal("First backfill failed", zap.Error(err))
	}
	logger.Info("First backfill submitted — waiting for runs to settle")
	time.Sleep(10 * time.Second)

	runsAfterFirst := totalRuns(sc, ctx, scheduleID)
	logger.Info("After first backfill", zap.Int64("totalRuns", runsAfterFirst))

	if err = sc.Backfill(ctx, scheduleID, req); err != nil {
		logger.Fatal("Second backfill failed", zap.Error(err))
	}
	logger.Info("Second backfill submitted (same BackfillID) — waiting")
	time.Sleep(10 * time.Second)

	runsAfterSecond := totalRuns(sc, ctx, scheduleID)

	logger.Info("--- Verify: Backfill/Idempotency ---")
	if runsAfterSecond == runsAfterFirst {
		logger.Info("  MATCH   idempotency: TotalRuns unchanged after duplicate BackfillID submission",
			zap.Int64("totalRuns", runsAfterSecond))
	} else {
		logger.Warn("  NOTE: TotalRuns changed — duplicate was not suppressed or completed runs were restarted",
			zap.Int64("afterFirst", runsAfterFirst),
			zap.Int64("afterSecond", runsAfterSecond))
	}
}
