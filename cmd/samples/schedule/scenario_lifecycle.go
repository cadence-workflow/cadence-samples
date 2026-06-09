package main

import (
	"context"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/zap"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
)

// runLifecycle walks the full schedule lifecycle with rich, fully-populated requests and
// verifies the round-trip via Describe. It folds in several SDK plan cases:
//   - 1. full-field create → describe round-trip
//   - 2. Create is not idempotent
//   - 3. Update is describe-then-update (mutate current state; untouched fields are preserved)
//   - 4. Pause (reason), Backfill (runs fire), Delete (absent from List afterward)
//   - 5. List entry fields
func runLifecycle(h *common.SampleHelper) {
	c := buildClient(h)
	sc := c.ScheduleClient()
	ctx := context.Background()
	scheduleID := newScheduleID("sample-lifecycle")
	defer deleteQuietly(h, sc, context.Background(), scheduleID)

	// ── 1. Create with every settable field ────────────────────────────────
	h.Logger.Info("=== Create (full field) ===", zap.String("scheduleID", scheduleID))
	now := time.Now()
	action := startWorkflowAction(h, 0)
	action.WorkflowIDPrefix = "lifecycle-run"
	// Action-level Memo: you pass a native value, the SDK encodes it. On Describe it comes
	// back as raw bytes (map[string][]byte) which you decode — see actionMemoPresent below.
	action.Memo = map[string]interface{}{"actionNote": "round-trips via describe"}

	_, err := sc.Create(ctx, &client.CreateScheduleRequest{
		ScheduleID: scheduleID,
		Spec: &client.ScheduleSpec{
			CronExpression: "0 * * * *", // hourly
			StartTime:      now,
			EndTime:        now.Add(30 * 24 * time.Hour),
			Jitter:         5 * time.Second,
		},
		Action: &client.ScheduleAction{StartWorkflow: action},
		Policies: &client.SchedulePolicies{
			OverlapPolicy:    client.ScheduleOverlapPolicySkipNew,
			CatchUpPolicy:    client.ScheduleCatchUpPolicyOne,
			CatchUpWindow:    time.Hour,
			PauseOnFailure:   true,
			BufferLimit:      10,
			ConcurrencyLimit: 5,
		},
		// Schedule-level Memo IS returned (raw bytes) by Describe — see decode in the
		// dataconverter scenario. Here we just round-trip its presence.
		Memo: map[string]interface{}{"owner": "schedule-sample", "purpose": "lifecycle-demo"},
	})
	if err != nil {
		h.Logger.Fatal("Create failed", zap.Error(err))
	}

	// Describe and log the full round-trip, calling out the documented gaps.
	desc := mustDescribe(h, sc, ctx, scheduleID)
	logDescribe(h, "after create", desc)

	// ── §2.5-C Create is not idempotent ──────────────────────────────────────
	h.Logger.Info("=== Create again with same ID (expected to fail — not idempotent) ===")
	if _, err = sc.Create(ctx, &client.CreateScheduleRequest{
		ScheduleID: scheduleID,
		Spec:       &client.ScheduleSpec{CronExpression: "0 * * * *"},
		Action:     &client.ScheduleAction{StartWorkflow: startWorkflowAction(h, 0)},
	}); err != nil {
		h.Logger.Info("Duplicate create rejected as expected", zap.String("error", err.Error()))
	} else {
		h.Logger.Warn("Duplicate create unexpectedly SUCCEEDED — Create should not be idempotent")
	}

	// ── §2.2 Update is describe-then-update (siblings preserved) ──────────────
	// The Update API takes a callback. The SDK calls DescribeSchedule for you, hands you the
	// current state pre-populated in *client.ScheduleUpdate, and you mutate only what you want
	// to change. The SDK diffs against the described baseline and sends just those changes, so
	// sibling sub-fields (PauseOnFailure, BufferLimit, ...) and untouched top-level fields (the
	// Action) survive — no need to re-send the whole schedule, no accidental resets.
	h.Logger.Info("=== Update (describe-then-update: change cron + one policy sub-field) ===")
	if err = sc.Update(ctx, scheduleID, func(u *client.ScheduleUpdate) error {
		u.Spec.CronExpression = "0 */2 * * *"                              // every 2h
		u.Policies.OverlapPolicy = client.ScheduleOverlapPolicyConcurrent // flip ONE sub-field
		return nil
	}); err != nil {
		h.Logger.Fatal("Update failed", zap.Error(err))
	}
	desc = mustDescribe(h, sc, ctx, scheduleID)
	logDescribe(h, "after describe-then-update", desc)
	h.Logger.Info("Note: PauseOnFailure / BufferLimit / ConcurrencyLimit are PRESERVED — " +
		"describe-then-update changes only the sub-fields you touch; the siblings survive " +
		"(this is the opposite of a blind full-replacement). The Action is untouched too.")

	// Update again, changing ONLY the action-level Memo via the SetActionMemo helper (which
	// encodes native Go values the same way Create does). Spec and Policies are left alone and
	// therefore preserved — demonstrating a targeted action change.
	h.Logger.Info("=== Update (action memo only, via SetActionMemo) ===")
	if err = sc.Update(ctx, scheduleID, func(u *client.ScheduleUpdate) error {
		return u.SetActionMemo(map[string]interface{}{"actionNote": "updated via describe-then-update"})
	}); err != nil {
		h.Logger.Fatal("Update (action memo) failed", zap.Error(err))
	}
	desc = mustDescribe(h, sc, ctx, scheduleID)
	logDescribe(h, "after action-memo update", desc)

	// ── §2.3 Pause (reason is reflected in PauseInfo) ────────────────────────
	const pauseReason = "lifecycle demo pause"
	h.Logger.Info("=== Pause ===", zap.String("reason", pauseReason))
	if err = sc.Pause(ctx, scheduleID, pauseReason); err != nil {
		h.Logger.Fatal("Pause failed", zap.Error(err))
	}
	desc = mustDescribe(h, sc, ctx, scheduleID)
	if desc.State != nil && desc.State.PauseInfo != nil {
		h.Logger.Info("Paused",
			zap.Bool("paused", desc.State.Paused),
			zap.String("reason", desc.State.PauseInfo.Reason))
	}

	// ── §2.3 Unpause ─────────────────────────────────────────────────────────
	h.Logger.Info("=== Unpause ===")
	if err = sc.Unpause(ctx, scheduleID, "resuming after demo", client.ScheduleCatchUpPolicySkip); err != nil {
		h.Logger.Fatal("Unpause failed", zap.Error(err))
	}
	desc = mustDescribe(h, sc, ctx, scheduleID)
	h.Logger.Info("Unpaused", zap.Bool("paused", desc.State != nil && desc.State.Paused))

	// ── §2.3 Backfill a past range (inclusive of both endpoints) ─────────────
	// Cron is now every 2h; backfill a 6-hour window so several boundaries fire.
	h.Logger.Info("=== Backfill past 6h (watch the worker pick up runs) ===")
	if err = sc.Backfill(ctx, scheduleID, &client.BackfillRequest{
		StartTime:     now.Add(-6 * time.Hour),
		EndTime:       now.Add(-1 * time.Hour),
		OverlapPolicy: client.ScheduleOverlapPolicySkipNew,
	}); err != nil {
		h.Logger.Fatal("Backfill failed", zap.Error(err))
	}
	h.Logger.Info("Backfill submitted — sleeping briefly so the worker can run the backfilled executions")
	time.Sleep(5 * time.Second)

	// ── §2.4 List entry fields ───────────────────────────────────────────────
	h.Logger.Info("=== List (entry fields) ===")
	if entry := findInList(h, sc, ctx, scheduleID); entry != nil {
		h.Logger.Info("Found our schedule in List",
			zap.String("scheduleID", entry.ScheduleID),
			zap.String("workflowType", entry.WorkflowType),
			zap.String("cron", entry.CronExpression),
			zap.Bool("paused", entry.State != nil && entry.State.Paused))
	} else {
		h.Logger.Warn("Our schedule not yet visible in List (indexing may lag)")
	}

	// ── §2.3 Delete, then confirm it disappears from List ────────────────────
	h.Logger.Info("=== Delete ===")
	if err = sc.Delete(ctx, scheduleID); err != nil {
		h.Logger.Fatal("Delete failed", zap.Error(err))
	}
	// Deletion is async; poll List until the schedule is gone (or give up after a bit).
	gone := false
	for deadline := time.Now().Add(15 * time.Second); time.Now().Before(deadline); time.Sleep(time.Second) {
		if findInList(h, sc, ctx, scheduleID) == nil {
			gone = true
			break
		}
	}
	if gone {
		h.Logger.Info("Confirmed: deleted schedule no longer appears in List")
	} else {
		h.Logger.Warn("Deleted schedule still in List after 15s (deletion is async / indexing may lag)")
	}
}

// mustDescribe describes a schedule, failing fast on error.
func mustDescribe(h *common.SampleHelper, sc client.ScheduleClient, ctx context.Context, id string) *client.DescribeScheduleResponse {
	desc, err := sc.Describe(ctx, id)
	if err != nil {
		h.Logger.Fatal("Describe failed", zap.String("scheduleID", id), zap.Error(err))
	}
	return desc
}

// findInList returns the list entry for id, or nil if not present.
func findInList(h *common.SampleHelper, sc client.ScheduleClient, ctx context.Context, id string) *client.ScheduleListEntry {
	resp, err := sc.List(ctx, 100, nil)
	if err != nil {
		h.Logger.Fatal("List failed", zap.Error(err))
	}
	for _, e := range resp.Schedules {
		if e.ScheduleID == id {
			return e
		}
	}
	return nil
}

// logDescribe prints the interesting fields of a Describe response, calling out the
// fields the server is known not to populate yet (so they read as findings, not bugs).
func logDescribe(h *common.SampleHelper, when string, desc *client.DescribeScheduleResponse) {
	if desc == nil {
		return
	}
	if desc.Spec != nil {
		h.Logger.Info("Describe "+when+": spec",
			zap.String("cron", desc.Spec.CronExpression),
			zap.Time("startTime", desc.Spec.StartTime),
			zap.Time("endTime", desc.Spec.EndTime),
			zap.Duration("jitter", desc.Spec.Jitter))
	}
	if desc.Action != nil && desc.Action.StartWorkflow != nil {
		h.Logger.Info("Describe "+when+": action",
			zap.String("workflowType", desc.Action.StartWorkflow.WorkflowType),
			zap.String("taskList", desc.Action.StartWorkflow.TaskList),
			zap.Bool("actionMemoPresent", len(desc.Action.StartWorkflow.Memo) > 0)) // expected true — returned on read
	}
	if desc.Policies != nil {
		h.Logger.Info("Describe "+when+": policies",
			zap.Int("overlapPolicy", int(desc.Policies.OverlapPolicy)),
			zap.Int("catchUpPolicy", int(desc.Policies.CatchUpPolicy)),
			zap.Duration("catchUpWindow", desc.Policies.CatchUpWindow),
			zap.Bool("pauseOnFailure", desc.Policies.PauseOnFailure),
			zap.Int32("bufferLimit", desc.Policies.BufferLimit),
			zap.Int32("concurrencyLimit", desc.Policies.ConcurrencyLimit))
	}
	h.Logger.Info("Describe "+when+": schedule-level memo present (raw bytes, decodable with your DataConverter)",
		zap.Bool("memoPresent", len(desc.Memo) > 0))
	if desc.Info != nil {
		// CreateTime/LastUpdateTime (zero) and OngoingBackfills (nil) are documented gaps.
		h.Logger.Info("Describe "+when+": info",
			zap.Int64("totalRuns", desc.Info.TotalRuns),
			zap.Time("createTime_BUG1_zero", desc.Info.CreateTime),
			zap.Time("lastUpdateTime_BUG1_zero", desc.Info.LastUpdateTime),
			zap.Int("ongoingBackfills_BUG3_nil", len(desc.Info.OngoingBackfills)))
	}
}
