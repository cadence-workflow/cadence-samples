package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
)

// PickFirstWorkflow races activities and uses first result.
func PickFirstWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("PickFirstWorkflow started")

	childCtx, cancelHandler := workflow.WithCancel(ctx)
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
		WaitForCancellation:    true,
	}
	childCtx = workflow.WithActivityOptions(childCtx, ao)

	selector := workflow.NewSelector(ctx)
	var firstResponse string

	f1 := workflow.ExecuteActivity(childCtx, RaceActivity, 0, time.Second*2)
	f2 := workflow.ExecuteActivity(childCtx, RaceActivity, 1, time.Second*10)
	pendingFutures := []workflow.Future{f1, f2}

	selector.AddFuture(f1, func(f workflow.Future) { f.Get(ctx, &firstResponse) })
	selector.AddFuture(f2, func(f workflow.Future) { f.Get(ctx, &firstResponse) })

	selector.Select(ctx)
	cancelHandler()

	for _, f := range pendingFutures {
		f.Get(ctx, nil)
	}

	logger.Info("PickFirstWorkflow completed")
	return nil
}

// RaceActivity runs for specified duration with heartbeats.
func RaceActivity(ctx context.Context, branchID int, duration time.Duration) (string, error) {
	elapsed := time.Duration(0)
	for elapsed < duration {
		time.Sleep(time.Second)
		elapsed += time.Second
		activity.RecordHeartbeat(ctx, "progress")
		select {
		case <-ctx.Done():
			return fmt.Sprintf("Branch %d cancelled", branchID), ctx.Err()
		default:
		}
	}
	return fmt.Sprintf("Branch %d done in %s", branchID, duration), nil
}

