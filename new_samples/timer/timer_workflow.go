package main

import (
	"context"
	"math/rand"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// TimerWorkflow demonstrates using timers for timeouts and delayed notifications.
// It starts a long-running process and sends a notification if it takes too long.
func TimerWorkflow(ctx workflow.Context, processingTimeThreshold time.Duration) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("TimerWorkflow started", zap.Duration("threshold", processingTimeThreshold))

	// Create a cancellable context for the timer
	childCtx, cancelHandler := workflow.WithCancel(ctx)
	selector := workflow.NewSelector(ctx)

	// Track if processing is done
	var processingDone bool

	// Start the order processing activity
	f := workflow.ExecuteActivity(ctx, OrderProcessingActivity)
	selector.AddFuture(f, func(f workflow.Future) {
		processingDone = true
		// Cancel the timer since processing completed
		cancelHandler()
		logger.Info("Processing completed, timer cancelled")
	})

	// Start a timer that fires if processing takes too long
	timerFuture := workflow.NewTimer(childCtx, processingTimeThreshold)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		if !processingDone {
			// Processing not done when timer fires, send notification
			logger.Info("Timer fired - processing taking too long, sending notification")
			workflow.ExecuteActivity(ctx, SendEmailActivity).Get(ctx, nil)
		}
	})

	// Wait for either timer or processing to complete first
	selector.Select(ctx)

	// If timer fired first, still wait for processing to complete
	if !processingDone {
		selector.Select(ctx)
	}

	logger.Info("TimerWorkflow completed")
	return nil
}

// OrderProcessingActivity simulates a long-running order processing operation.
// Processing time is random between 0-10 seconds.
func OrderProcessingActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("OrderProcessingActivity started")

	// Simulate random processing time
	timeNeededToProcess := time.Second * time.Duration(rand.Intn(10))
	time.Sleep(timeNeededToProcess)

	logger.Info("OrderProcessingActivity completed", zap.Duration("duration", timeNeededToProcess))
	return nil
}

// SendEmailActivity sends a notification email when processing takes too long.
func SendEmailActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("SendEmailActivity: Sending notification - processing is taking longer than expected")
	return nil
}

