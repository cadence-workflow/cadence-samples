package main

import (
	"context"
	"math/rand"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// TimerWorkflow demonstrates using timers for timeout notifications.
func TimerWorkflow(ctx workflow.Context, processingTimeThreshold time.Duration) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("TimerWorkflow started")

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	childCtx, cancelHandler := workflow.WithCancel(ctx)
	selector := workflow.NewSelector(ctx)

	var processingDone bool
	f := workflow.ExecuteActivity(ctx, OrderProcessingActivity)
	selector.AddFuture(f, func(f workflow.Future) {
		processingDone = true
		cancelHandler()
	})

	timerFuture := workflow.NewTimer(childCtx, processingTimeThreshold)
	selector.AddFuture(timerFuture, func(f workflow.Future) {
		if !processingDone {
			workflow.ExecuteActivity(ctx, SendEmailActivity).Get(ctx, nil)
		}
	})

	selector.Select(ctx)
	if !processingDone {
		selector.Select(ctx)
	}

	logger.Info("TimerWorkflow completed")
	return nil
}

// OrderProcessingActivity simulates order processing (random 0-10s).
func OrderProcessingActivity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	duration := time.Second * time.Duration(rand.Intn(10))
	logger.Info("OrderProcessingActivity started", zap.Duration("duration", duration))
	time.Sleep(duration)
	return nil
}

// SendEmailActivity sends notification when processing takes too long.
func SendEmailActivity(ctx context.Context) error {
	activity.GetLogger(ctx).Info("SendEmailActivity: Sending delay notification")
	return nil
}

