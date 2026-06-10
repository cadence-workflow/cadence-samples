package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/zap"
)

// runLifecycle walks the full schedule lifecycle with rich, fully-populated requests and
// verifies the round-trip via Describe. It covers:
//   - 1. full-field create → describe round-trip
//   - 2. Create is not idempotent
//   - 3. Update is describe-then-update (mutate current state; untouched fields are preserved)
//   - 4. Pause (reason) / Unpause
//   - 5. List entry fields
//   - 6. Delete (absent from List afterward)
//
// Backfill is covered separately by scenario_backfill.go.
func runLifecycle() {
	logger := BuildLogger()
	c := buildScheduleClient(nil)
	sc := c.ScheduleClient()
	ctx := context.Background()
	scheduleID := newScheduleID("sample-lifecycle")
	defer deleteQuietly(logger, sc, context.Background(), scheduleID)

	// ── Step 1: Create with every settable field ─────────────────────────────
	logger.Info("=== Create (full field) ===", zap.String("scheduleID", scheduleID))
	now := time.Now()
	action := startWorkflowAction(logger, 0)
	action.WorkflowIDPrefix = "lifecycle-run"
	action.Memo = map[string]interface{}{"actionNote": "round-trips via describe"}

	createReq := &client.CreateScheduleRequest{
		ScheduleID: scheduleID,
		Spec: &client.ScheduleSpec{
			CronExpression: "0 * * * *",
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
		Memo: map[string]interface{}{"owner": "schedule-sample", "purpose": "lifecycle-demo"},
	}

	_, err := sc.Create(ctx, createReq)
	if err != nil {
		logger.Fatal("Create failed", zap.Error(err))
	}

	desc := mustDescribe(logger, sc, ctx, scheduleID)
	verifyAfterCreate(logger, createReq, desc)

	// ── Step 2: Create is not idempotent ─────────────────────────────────────
	logger.Info("=== Create again with same ID (expected to fail — not idempotent) ===")
	if _, err = sc.Create(ctx, &client.CreateScheduleRequest{
		ScheduleID: scheduleID,
		Spec:       &client.ScheduleSpec{CronExpression: "0 * * * *"},
		Action:     &client.ScheduleAction{StartWorkflow: startWorkflowAction(logger, 0)},
	}); err != nil {
		logger.Info("Duplicate create rejected as expected", zap.String("error", err.Error()))
	} else {
		logger.Warn("Duplicate create unexpectedly SUCCEEDED — Create should not be idempotent")
	}

	// ── Step 3: Update — change cron + one policy sub-field ──────────────────
	logger.Info("=== Update (describe-then-update: change cron + one policy sub-field) ===")
	wantCron := "0 */2 * * *"
	wantOverlap := client.ScheduleOverlapPolicyConcurrent
	if err = sc.Update(ctx, scheduleID, func(u *client.ScheduleUpdate) error {
		u.Spec.CronExpression = wantCron
		u.Policies.OverlapPolicy = wantOverlap
		return nil
	}); err != nil {
		logger.Fatal("Update failed", zap.Error(err))
	}
	desc = mustDescribe(logger, sc, ctx, scheduleID)
	verifyAfterCronAndOverlapUpdate(logger, wantCron, wantOverlap, createReq.Policies, desc)

	// ── Step 4: Update — action memo only ───────────────────────────────────
	logger.Info("=== Update (action memo only, via SetActionMemo) ===")
	if err = sc.Update(ctx, scheduleID, func(u *client.ScheduleUpdate) error {
		return u.SetActionMemo(map[string]interface{}{"actionNote": "updated via describe-then-update"})
	}); err != nil {
		logger.Fatal("Update (action memo) failed", zap.Error(err))
	}
	desc = mustDescribe(logger, sc, ctx, scheduleID)
	verifyAfterActionMemoUpdate(logger, wantCron, wantOverlap, desc)

	// ── Step 5: Pause ────────────────────────────────────────────────────────
	const pauseReason = "lifecycle demo pause"
	logger.Info("=== Pause ===", zap.String("reason", pauseReason))
	if err = sc.Pause(ctx, scheduleID, pauseReason); err != nil {
		logger.Fatal("Pause failed", zap.Error(err))
	}
	desc = mustDescribe(logger, sc, ctx, scheduleID)
	verifyAfterPause(logger, pauseReason, desc)

	// ── Step 6: Unpause ──────────────────────────────────────────────────────
	logger.Info("=== Unpause ===")
	if err = sc.Unpause(ctx, scheduleID, "resuming after demo", client.ScheduleCatchUpPolicySkip); err != nil {
		logger.Fatal("Unpause failed", zap.Error(err))
	}
	desc = mustDescribe(logger, sc, ctx, scheduleID)
	verifyAfterUnpause(logger, desc)

	// ── Step 7: List entry fields ────────────────────────────────────────────
	logger.Info("=== List (entry fields) ===")
	entry := findInList(logger, sc, ctx, scheduleID)
	verifyListEntry(logger, scheduleID, entry)

	// ── Step 8: Delete, then confirm it disappears from List ─────────────────
	logger.Info("=== Delete ===")
	if err = sc.Delete(ctx, scheduleID); err != nil {
		logger.Fatal("Delete failed", zap.Error(err))
	}
	gone := false
	for deadline := time.Now().Add(15 * time.Second); time.Now().Before(deadline); time.Sleep(time.Second) {
		if findInList(logger, sc, ctx, scheduleID) == nil {
			gone = true
			break
		}
	}
	if gone {
		logger.Info("Confirmed: deleted schedule no longer appears in List")
	} else {
		logger.Warn("Deleted schedule still in List after 15s (deletion is async / indexing may lag)")
	}
}

func mustDescribe(logger *zap.Logger, sc client.ScheduleClient, ctx context.Context, id string) *client.DescribeScheduleResponse {
	desc, err := sc.Describe(ctx, id)
	if err != nil {
		logger.Fatal("Describe failed", zap.String("scheduleID", id), zap.Error(err))
	}
	return desc
}

func findInList(logger *zap.Logger, sc client.ScheduleClient, ctx context.Context, id string) *client.ScheduleListEntry {
	resp, err := sc.List(ctx, 100, nil)
	if err != nil {
		logger.Fatal("List failed", zap.Error(err))
	}
	for _, e := range resp.Schedules {
		if e.ScheduleID == id {
			return e
		}
	}
	return nil
}

func logCmp(logger *zap.Logger, field string, want, got interface{}) {
	if fmt.Sprintf("%v", want) == fmt.Sprintf("%v", got) {
		logger.Info("  MATCH   "+field, zap.Any("value", got))
	} else {
		logger.Warn("  MISMATCH "+field, zap.Any("want", want), zap.Any("got", got))
	}
}

func verifyAfterCreate(logger *zap.Logger, req *client.CreateScheduleRequest, desc *client.DescribeScheduleResponse) {
	logger.Info("--- Verify: Create round-trip ---")
	if req.Spec != nil && desc.Spec != nil {
		logCmp(logger, "spec.cron", req.Spec.CronExpression, desc.Spec.CronExpression)
		logCmp(logger, "spec.startTime", req.Spec.StartTime.UTC().Truncate(time.Second), desc.Spec.StartTime.UTC().Truncate(time.Second))
		logCmp(logger, "spec.endTime", req.Spec.EndTime.UTC().Truncate(time.Second), desc.Spec.EndTime.UTC().Truncate(time.Second))
		logCmp(logger, "spec.jitter", req.Spec.Jitter, desc.Spec.Jitter)
	}
	if req.Action != nil && req.Action.StartWorkflow != nil && desc.Action != nil && desc.Action.StartWorkflow != nil {
		logCmp(logger, "action.workflowType", req.Action.StartWorkflow.WorkflowType, desc.Action.StartWorkflow.WorkflowType)
		logCmp(logger, "action.taskList", req.Action.StartWorkflow.TaskList, desc.Action.StartWorkflow.TaskList)
		logCmp(logger, "action.memoPresent", len(req.Action.StartWorkflow.Memo) > 0, len(desc.Action.StartWorkflow.Memo) > 0)
	}
	if req.Policies != nil && desc.Policies != nil {
		logCmp(logger, "policies.overlapPolicy", req.Policies.OverlapPolicy, desc.Policies.OverlapPolicy)
		logCmp(logger, "policies.catchUpPolicy", req.Policies.CatchUpPolicy, desc.Policies.CatchUpPolicy)
		logCmp(logger, "policies.catchUpWindow", req.Policies.CatchUpWindow, desc.Policies.CatchUpWindow)
		logCmp(logger, "policies.pauseOnFailure", req.Policies.PauseOnFailure, desc.Policies.PauseOnFailure)
		logCmp(logger, "policies.bufferLimit", req.Policies.BufferLimit, desc.Policies.BufferLimit)
		logCmp(logger, "policies.concurrencyLimit", req.Policies.ConcurrencyLimit, desc.Policies.ConcurrencyLimit)
	}
	logCmp(logger, "memo.present", len(req.Memo) > 0, len(desc.Memo) > 0)
	if desc.Info != nil {
		logger.Info("  info (known server gaps)",
			zap.Time("createTime_zero", desc.Info.CreateTime),
			zap.Time("lastUpdateTime_zero", desc.Info.LastUpdateTime),
			zap.Int("ongoingBackfills_nil", len(desc.Info.OngoingBackfills)))
	}
}

func verifyAfterCronAndOverlapUpdate(logger *zap.Logger, wantCron string, wantOverlap client.ScheduleOverlapPolicy, createPolicies *client.SchedulePolicies, desc *client.DescribeScheduleResponse) {
	logger.Info("--- Verify: Update 1 — changed fields ---")
	if desc.Spec != nil {
		logCmp(logger, "spec.cron (changed)", wantCron, desc.Spec.CronExpression)
	}
	if desc.Policies != nil {
		logCmp(logger, "policies.overlapPolicy (changed)", wantOverlap, desc.Policies.OverlapPolicy)
	}
	if createPolicies != nil && desc.Policies != nil {
		logger.Info("--- Verify: Update 1 — sibling policies preserved ---")
		logCmp(logger, "policies.catchUpPolicy (preserved)", createPolicies.CatchUpPolicy, desc.Policies.CatchUpPolicy)
		logCmp(logger, "policies.catchUpWindow (preserved)", createPolicies.CatchUpWindow, desc.Policies.CatchUpWindow)
		logCmp(logger, "policies.pauseOnFailure (preserved)", createPolicies.PauseOnFailure, desc.Policies.PauseOnFailure)
		logCmp(logger, "policies.bufferLimit (preserved)", createPolicies.BufferLimit, desc.Policies.BufferLimit)
		logCmp(logger, "policies.concurrencyLimit (preserved)", createPolicies.ConcurrencyLimit, desc.Policies.ConcurrencyLimit)
	}
}

func verifyAfterActionMemoUpdate(logger *zap.Logger, wantCron string, wantOverlap client.ScheduleOverlapPolicy, desc *client.DescribeScheduleResponse) {
	logger.Info("--- Verify: Update 2 — action memo changed ---")
	if desc.Action != nil && desc.Action.StartWorkflow != nil {
		logCmp(logger, "action.memoPresent (changed)", true, len(desc.Action.StartWorkflow.Memo) > 0)
	}
	logger.Info("--- Verify: SetActionMemo left spec.cron and overlapPolicy untouched ---")
	if desc.Spec != nil {
		logCmp(logger, "spec.cron (preserved)", wantCron, desc.Spec.CronExpression)
	}
	if desc.Policies != nil {
		logCmp(logger, "policies.overlapPolicy (preserved)", wantOverlap, desc.Policies.OverlapPolicy)
	}
}

func verifyAfterPause(logger *zap.Logger, wantReason string, desc *client.DescribeScheduleResponse) {
	logger.Info("--- Verify: Pause ---")
	if desc.State == nil {
		logger.Warn("  MISMATCH state is nil")
		return
	}
	logCmp(logger, "state.paused", true, desc.State.Paused)
	if desc.State.PauseInfo != nil {
		logCmp(logger, "state.pauseInfo.reason", wantReason, desc.State.PauseInfo.Reason)
	} else {
		logger.Warn("  MISMATCH state.pauseInfo is nil, want reason=" + wantReason)
	}
}

func verifyAfterUnpause(logger *zap.Logger, desc *client.DescribeScheduleResponse) {
	logger.Info("--- Verify: Unpause ---")
	if desc.State == nil {
		logger.Warn("  MISMATCH state is nil")
		return
	}
	logCmp(logger, "state.paused", false, desc.State.Paused)
}

func verifyListEntry(logger *zap.Logger, wantScheduleID string, entry *client.ScheduleListEntry) {
	logger.Info("--- Verify: List entry fields ---")
	if entry == nil {
		logger.Warn("  MISMATCH schedule not found in List (indexing may lag)")
		return
	}
	logCmp(logger, "list.scheduleID", wantScheduleID, entry.ScheduleID)
	if entry.State != nil {
		logCmp(logger, "list.paused", false, entry.State.Paused)
	} else {
		logger.Warn("  MISMATCH list.state is nil")
	}
}
