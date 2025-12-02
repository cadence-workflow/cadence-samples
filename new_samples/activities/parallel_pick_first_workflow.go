package workflows

import (
	"context"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"time"
)

type parallelBranchInput struct {
	Message string `json:"message"`
}

// ParallelBranchPickFirstWorkflow is a sample workflow simulating two parallel activities running
// at the same time and picking the first successful result.
func ParallelBranchPickFirstWorkflow(ctx workflow.Context) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ParallelBranchPickFirstWorkflow started")

	selector := workflow.NewSelector(ctx)
	var firstResp string

	// Use a cancel handler to cancel all rest of other activities.
	childCtx, cancelHandler := workflow.WithCancel(ctx)
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
		WaitForCancellation:    true, // wait for cancellation to complete
	}
	childCtx = workflow.WithActivityOptions(childCtx, ao)

	// Set WaitForCancellation to true to demonstrate the cancellation to the other activities. In real world case,
	// you might not care about them and could set WaitForCancellation to false (which is default value).

	// Run two activities in parallel
	f1 := workflow.ExecuteActivity(childCtx, ParallelActivity, parallelBranchInput{Message: "first activity"}, time.Second*10)
	f2 := workflow.ExecuteActivity(childCtx, ParallelActivity, parallelBranchInput{Message: "second activity"}, time.Second*2)
	pendingFutures := []workflow.Future{f1, f2}
	selector.AddFuture(f1, func(f workflow.Future) {
		f.Get(ctx, &firstResp)
	}).AddFuture(f2, func(f workflow.Future) {
		f.Get(ctx, &firstResp)
	})

	// wait for any of the future to complete
	selector.Select(ctx)

	// now if at least one future is complete, cancel all other pending futures
	cancelHandler()

	// - If you want to wait for pending activities to finish after issuing cancellation
	// then wait for the future to complete.
	// - if you don't want to wait for completion of pending activities cancellation then you can choose to
	// set WaitForCancellation to false through WithWaitForCancellation(false)
	for _, f := range pendingFutures {
		err := f.Get(ctx, &firstResp)
		if err != nil {
			return "", err
		}
	}

	logger.Info("ParallelBranchPickFirstWorkflow completed")
	return firstResp, nil
}

func ParallelActivity(ctx context.Context, input parallelBranchInput) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ParallelActivity started")
	return "Hello " + input.Message, nil
}
