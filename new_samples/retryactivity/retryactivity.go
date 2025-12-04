package main

import (
	"context"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// RetryWorkflow demonstrates retry policies for unreliable activities.
// The activity will fail and retry, resuming from heartbeated progress.
func RetryWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("RetryWorkflow started")

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute * 10,
		HeartbeatTimeout:       time.Second * 10,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:          time.Second,
			BackoffCoefficient:       2.0,
			MaximumInterval:          time.Minute,
			ExpirationInterval:       time.Minute * 5,
			MaximumAttempts:          5,
			NonRetriableErrorReasons: []string{"bad-error"},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, BatchProcessingActivity, 0, 20, time.Second).Get(ctx, nil)
	if err != nil {
		logger.Error("RetryWorkflow failed", zap.Error(err))
		return err
	}

	logger.Info("RetryWorkflow completed successfully")
	return nil
}

// BatchProcessingActivity processes tasks and intentionally fails to demonstrate retry.
// It heartbeats progress so retries can resume from where they left off.
func BatchProcessingActivity(ctx context.Context, firstTaskID, batchSize int, processDelay time.Duration) error {
	logger := activity.GetLogger(ctx)

	startFrom := firstTaskID

	// Check if we're retrying and have previous progress
	if activity.HasHeartbeatDetails(ctx) {
		var lastCompletedID int
		if err := activity.GetHeartbeatDetails(ctx, &lastCompletedID); err == nil {
			startFrom = lastCompletedID + 1
			logger.Info("Resuming from previous attempt", zap.Int("startFrom", startFrom))
		}
	}

	tasksInThisAttempt := 0
	for i := startFrom; i < firstTaskID+batchSize; i++ {
		logger.Info("Processing task", zap.Int("taskID", i))
		time.Sleep(processDelay)

		// Record progress
		activity.RecordHeartbeat(ctx, i)
		tasksInThisAttempt++

		// Simulate failure after processing 1/3 of tasks (but not on last task)
		if tasksInThisAttempt >= batchSize/3 && i < firstTaskID+batchSize-1 {
			logger.Info("Simulating failure - will retry from heartbeated progress")
			return cadence.NewCustomError("some-retryable-error")
		}
	}

	logger.Info("BatchProcessingActivity completed all tasks")
	return nil
}

