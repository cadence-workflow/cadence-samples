package batch

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/cadence/x"
)

type BatchWorkflowInput struct {
	Concurrency int
	TotalSize   int
}

func BatchWorkflow(ctx workflow.Context, input BatchWorkflowInput) error {
	factories := make([]func(workflow.Context) workflow.Future, input.TotalSize)
	for taskID := 0; taskID < input.TotalSize; taskID++ {
		taskID := taskID
		factories[taskID] = func(ctx workflow.Context) workflow.Future {
			aCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				ScheduleToStartTimeout: time.Second * 10,
				StartToCloseTimeout:    time.Second * 10,
			})
			return workflow.ExecuteActivity(aCtx, BatchActivity, taskID)
		}
	}

	batch, err := x.NewBatchFuture(ctx, input.Concurrency, factories)
	if err != nil {
		return fmt.Errorf("failed to create batch future: %w", err)
	}

	return batch.Get(ctx, nil)
}

func BatchActivity(ctx context.Context, taskID int) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("batch activity %d failed: %w", taskID, ctx.Err())
	case <-time.After(time.Duration(rand.Int63n(100))*time.Millisecond + 900*time.Millisecond):
		return nil
	}
}
