package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// PickFirstWorkflow demonstrates race condition handling.
// It executes activities in parallel and uses the result of the first to complete.
func PickFirstWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("PickFirstWorkflow started")

	// Create cancellable context for all activities
	childCtx, cancelHandler := workflow.WithCancel(ctx)
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
		WaitForCancellation:    true, // Wait for cancellation to complete
	}
	childCtx = workflow.WithActivityOptions(childCtx, ao)

	selector := workflow.NewSelector(ctx)
	var firstResponse string

	// Start two activities with different durations
	// Activity 0: takes 2 seconds (will win)
	// Activity 1: takes 10 seconds (will be cancelled)
	f1 := workflow.ExecuteActivity(childCtx, RaceActivity, 0, time.Second*2)
	f2 := workflow.ExecuteActivity(childCtx, RaceActivity, 1, time.Second*10)
	pendingFutures := []workflow.Future{f1, f2}

	selector.AddFuture(f1, func(f workflow.Future) {
		f.Get(ctx, &firstResponse)
	}).AddFuture(f2, func(f workflow.Future) {
		f.Get(ctx, &firstResponse)
	})

	// Wait for first to complete
	selector.Select(ctx)
	logger.Info("First activity completed", zap.String("result", firstResponse))

	// Cancel all other pending activities
	cancelHandler()

	// Wait for all activities to acknowledge cancellation
	for _, f := range pendingFutures {
		f.Get(ctx, nil)
	}

	logger.Info("PickFirstWorkflow completed")
	return nil
}

// RaceActivity simulates an activity that takes a specified duration.
// It heartbeats every second and checks for cancellation.
func RaceActivity(ctx context.Context, branchID int, duration time.Duration) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("RaceActivity started", zap.Int("branch", branchID), zap.Duration("duration", duration))

	elapsed := time.Duration(0)
	for elapsed < duration {
		time.Sleep(time.Second)
		elapsed += time.Second

		// Heartbeat to check for cancellation
		activity.RecordHeartbeat(ctx, fmt.Sprintf("branch %d: %v elapsed", branchID, elapsed))

		select {
		case <-ctx.Done():
			// Activity was cancelled
			msg := fmt.Sprintf("Branch %d cancelled after %v", branchID, elapsed)
			logger.Info(msg)
			return msg, ctx.Err()
		default:
			// Continue working
		}
	}

	msg := fmt.Sprintf("Branch %d completed in %v", branchID, duration)
	logger.Info(msg)
	return msg, nil
}

