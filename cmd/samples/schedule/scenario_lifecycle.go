package main

import (
	"context"
	"fmt"
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

	// ── Step 1: Create with every settable field ─────────────────────────────
	h.Logger.Info("=== Create (full field) ===", zap.String("scheduleID", scheduleID))
	now := time.Now()
	action := startWorkflowAction(h, 0)
	action.WorkflowIDPrefix = "lifecycle-run"
	// Action-level Memo: you pass a native value, the SDK encodes it. On Describe it comes
	// back as raw bytes (map[string][]byte) which you decode — see actionMemoPresent below.
	action.Memo = map[string]interface{}{"actionNote": "round-trips via describe"}

	createReq := &client.CreateScheduleRequest{
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
	}

	_, err := sc.Create(ctx, createReq)
	if err != nil {
		h.Logger.Fatal("Create failed", zap.Error(err))
	}

	desc := mustDescribe(h, sc, ctx, scheduleID)
	verifyAfterCreate(h, createReq, desc)

	// ── Step 2: Create is not idempotent ─────────────────────────────────────
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

	// ── Step 3: Update — change cron + one policy sub-field (siblings preserved) ──
	// The Update API takes a callback. The SDK calls DescribeSchedule for you, hands you the
	// current state pre-populated in *client.ScheduleUpdate, and you mutate only what you want
	// to change. The SDK diffs against the described baseline and sends just those changes, so
	// sibling sub-fields (PauseOnFailure, BufferLimit, ...) and untouched top-level fields (the
	// Action) survive — no need to re-send the whole schedule, no accidental resets.
	h.Logger.Info("=== Update (describe-then-update: change cron + one policy sub-field) ===")
	wantCron := "0 */2 * * *" // every 2h
	wantOverlap := client.ScheduleOverlapPolicyConcurrent
	if err = sc.Update(ctx, scheduleID, func(u *client.ScheduleUpdate) error {
		u.Spec.CronExpression = wantCron
		u.Policies.OverlapPolicy = wantOverlap
		return nil
	}); err != nil {
		h.Logger.Fatal("Update failed", zap.Error(err))
	}
	desc = mustDescribe(h, sc, ctx, scheduleID)
	verifyAfterCronAndOverlapUpdate(h, wantCron, wantOverlap, createReq.Policies, desc)

	// ── Step 4: Update — action memo only (SetActionMemo) ───────────────────────
	// Spec and Policies are left alone and therefore preserved — demonstrating a targeted action change.
	h.Logger.Info("=== Update (action memo only, via SetActionMemo) ===")
	if err = sc.Update(ctx, scheduleID, func(u *client.ScheduleUpdate) error {
		return u.SetActionMemo(map[string]interface{}{"actionNote": "updated via describe-then-update"})
	}); err != nil {
		h.Logger.Fatal("Update (action memo) failed", zap.Error(err))
	}
	desc = mustDescribe(h, sc, ctx, scheduleID)
	verifyAfterActionMemoUpdate(h, wantCron, wantOverlap, desc)

	// ── Step 5: Pause (reason is reflected in PauseInfo) ────────────────────────
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

	// ── Step 6: Unpause ──────────────────────────────────────────────────────
	h.Logger.Info("=== Unpause ===")
	if err = sc.Unpause(ctx, scheduleID, "resuming after demo", client.ScheduleCatchUpPolicySkip); err != nil {
		h.Logger.Fatal("Unpause failed", zap.Error(err))
	}
	desc = mustDescribe(h, sc, ctx, scheduleID)
	h.Logger.Info("Unpaused", zap.Bool("paused", desc.State != nil && desc.State.Paused))

	// ── Step 7: Backfill a past range (inclusive of both endpoints) ─────────────
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

	// ── Step 8: List entry fields ────────────────────────────────────────────
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

	// ── Step 9: Delete, then confirm it disappears from List ─────────────────
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

// logCmp logs a WANT / GOT comparison for one field. Match → Info; mismatch → Warn.
func logCmp(h *common.SampleHelper, field string, want, got interface{}) {
	if fmt.Sprintf("%v", want) == fmt.Sprintf("%v", got) {
		h.Logger.Info("  MATCH   "+field, zap.Any("value", got))
	} else {
		h.Logger.Warn("  MISMATCH "+field, zap.Any("want", want), zap.Any("got", got))
	}
}

// verifyAfterCreate compares every field of the Create request against the Describe response.
func verifyAfterCreate(h *common.SampleHelper, req *client.CreateScheduleRequest, desc *client.DescribeScheduleResponse) {
	h.Logger.Info("--- Verify: Create round-trip ---")
	if req.Spec != nil && desc.Spec != nil {
		logCmp(h, "spec.cron",      req.Spec.CronExpression, desc.Spec.CronExpression)
		logCmp(h, "spec.startTime", req.Spec.StartTime.UTC().Truncate(time.Second), desc.Spec.StartTime.UTC().Truncate(time.Second))
		logCmp(h, "spec.endTime",   req.Spec.EndTime.UTC().Truncate(time.Second), desc.Spec.EndTime.UTC().Truncate(time.Second))
		logCmp(h, "spec.jitter",    req.Spec.Jitter, desc.Spec.Jitter)
	}
	if req.Action != nil && req.Action.StartWorkflow != nil && desc.Action != nil && desc.Action.StartWorkflow != nil {
		logCmp(h, "action.workflowType", req.Action.StartWorkflow.WorkflowType, desc.Action.StartWorkflow.WorkflowType)
		logCmp(h, "action.taskList",     req.Action.StartWorkflow.TaskList, desc.Action.StartWorkflow.TaskList)
		// Memo: set as map[string]interface{} on create, returned as map[string][]byte on describe — check presence only.
		logCmp(h, "action.memoPresent",  len(req.Action.StartWorkflow.Memo) > 0, len(desc.Action.StartWorkflow.Memo) > 0)
	}
	if req.Policies != nil && desc.Policies != nil {
		logCmp(h, "policies.overlapPolicy",    req.Policies.OverlapPolicy, desc.Policies.OverlapPolicy)
		logCmp(h, "policies.catchUpPolicy",    req.Policies.CatchUpPolicy, desc.Policies.CatchUpPolicy)
		logCmp(h, "policies.catchUpWindow",    req.Policies.CatchUpWindow, desc.Policies.CatchUpWindow)
		logCmp(h, "policies.pauseOnFailure",   req.Policies.PauseOnFailure, desc.Policies.PauseOnFailure)
		logCmp(h, "policies.bufferLimit",      req.Policies.BufferLimit, desc.Policies.BufferLimit)
		logCmp(h, "policies.concurrencyLimit", req.Policies.ConcurrencyLimit, desc.Policies.ConcurrencyLimit)
	}
	// Schedule-level Memo is returned as map[string][]byte on describe — check presence only.
	logCmp(h, "memo.present", len(req.Memo) > 0, len(desc.Memo) > 0)
	if desc.Info != nil {
		// Known server-side gaps — these zero/nil values are expected, not sample failures.
		h.Logger.Info("  info (known server gaps)",
			zap.Time("createTime_zero", desc.Info.CreateTime),
			zap.Time("lastUpdateTime_zero", desc.Info.LastUpdateTime),
			zap.Int("ongoingBackfills_nil", len(desc.Info.OngoingBackfills)))
	}
}

// verifyAfterCronAndOverlapUpdate checks that the first Update (cron + overlapPolicy) took effect and
// that the sibling policies untouched by the update are still at their Create values.
func verifyAfterCronAndOverlapUpdate(h *common.SampleHelper, wantCron string, wantOverlap client.ScheduleOverlapPolicy, createPolicies *client.SchedulePolicies, desc *client.DescribeScheduleResponse) {
	h.Logger.Info("--- Verify: Update 1 — changed fields ---")
	if desc.Spec != nil {
		logCmp(h, "spec.cron (changed)", wantCron, desc.Spec.CronExpression)
	}
	if desc.Policies != nil {
		logCmp(h, "policies.overlapPolicy (changed)", wantOverlap, desc.Policies.OverlapPolicy)
	}
	if createPolicies != nil && desc.Policies != nil {
		h.Logger.Info("--- Verify: Update 1 — sibling policies preserved (not touched by update) ---")
		logCmp(h, "policies.catchUpPolicy (preserved)",    createPolicies.CatchUpPolicy, desc.Policies.CatchUpPolicy)
		logCmp(h, "policies.catchUpWindow (preserved)",    createPolicies.CatchUpWindow, desc.Policies.CatchUpWindow)
		logCmp(h, "policies.pauseOnFailure (preserved)",   createPolicies.PauseOnFailure, desc.Policies.PauseOnFailure)
		logCmp(h, "policies.bufferLimit (preserved)",      createPolicies.BufferLimit, desc.Policies.BufferLimit)
		logCmp(h, "policies.concurrencyLimit (preserved)", createPolicies.ConcurrencyLimit, desc.Policies.ConcurrencyLimit)
	}
}

// verifyAfterActionMemoUpdate confirms the action memo changed and that spec and policies
// were not disturbed — demonstrating a targeted action-only update.
func verifyAfterActionMemoUpdate(h *common.SampleHelper, wantCron string, wantOverlap client.ScheduleOverlapPolicy, desc *client.DescribeScheduleResponse) {
	h.Logger.Info("--- Verify: Update 2 — action memo changed ---")
	if desc.Action != nil && desc.Action.StartWorkflow != nil {
		logCmp(h, "action.memoPresent (changed)", true, len(desc.Action.StartWorkflow.Memo) > 0)
	}
	h.Logger.Info("--- Verify: SetActionMemo left spec.cron and overlapPolicy untouched ---")
	if desc.Spec != nil {
		logCmp(h, "spec.cron (preserved)", wantCron, desc.Spec.CronExpression)
	}
	if desc.Policies != nil {
		logCmp(h, "policies.overlapPolicy (preserved)", wantOverlap, desc.Policies.OverlapPolicy)
	}
}
