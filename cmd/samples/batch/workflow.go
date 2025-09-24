package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/cadence/workflow"
)

// ApplicationName is the task list for this sample
const ApplicationName = "batchGroup"

const batchWorkflowName = "batchWorkflow"

type BatchWorkflowInput struct {
	Concurrency int
	TotalSize   int
}

func BatchWorkflow(ctx workflow.Context, input BatchWorkflowInput) error {
	// Create activity factories for each task (not yet executed)
	factories := make([]func(workflow.Context) workflow.Future, input.TotalSize)
	for taskID := 0; taskID < input.TotalSize; taskID++ {
		taskID := taskID // Capture loop variable for closure
		factories[taskID] = func(ctx workflow.Context) workflow.Future {
			// Configure activity timeouts
			aCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
				ScheduleToStartTimeout: time.Minute * 1,
				StartToCloseTimeout:    time.Minute * 1,
			})
			return workflow.ExecuteActivity(aCtx, BatchActivity, taskID)
		}
	}

	// Execute all activities with controlled concurrency
	batch, err := workflow.NewBatchFuture(ctx, input.Concurrency, factories)
	if err != nil {
		return fmt.Errorf("failed to create batch future: %w", err)
	}

	// Wait for all activities to complete
	return batch.Get(ctx, nil)
}

func BatchActivity(ctx context.Context, taskID int) error {
	select {
	case <-ctx.Done():
		// Return error if workflow/activity is cancelled
		return fmt.Errorf("batch activity %d failed: %w", taskID, ctx.Err())
	case <-time.After(time.Duration(rand.Int63n(100))*time.Millisecond + 900*time.Millisecond):
		// Wait for random duration (900-999ms) to simulate work, then return success
		return nil
	}
}
