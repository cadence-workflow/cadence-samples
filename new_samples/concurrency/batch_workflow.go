package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/cadence/workflow"
)

// BatchWorkflowInput configures the batch processing parameters.
type BatchWorkflowInput struct {
	Concurrency int // Maximum number of activities running in parallel
	TotalSize   int // Total number of activities to process
}

// BatchWorkflow demonstrates processing large batches of activities with controlled
// concurrency using workflow.NewBatchFuture. This pattern is useful for:
// - Processing thousands of records without overwhelming downstream services
// - Respecting the 1024 pending activities limit per workflow
// - Automatic error handling and retry management
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

// BatchActivity simulates a unit of work that takes 900-999ms to complete.
// In real applications, this would be your actual processing logic.
func BatchActivity(ctx context.Context, taskID int) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("batch activity %d failed: %w", taskID, ctx.Err())
	case <-time.After(time.Duration(rand.Int63n(100))*time.Millisecond + 900*time.Millisecond):
		return nil
	}
}
