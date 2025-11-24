package main

import (
	"context"
	"fmt"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"time"
)

type sampleInput struct {
	Message string `json:"message"`
}

// This is the workflow function
// Given an input, HelloWorldWorkflow returns "Hello <input>!" 
func HelloWorldWorkflow(ctx workflow.Context, input sampleInput) (string, error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorldWorkflow started")

	var greetingMsg string
	err := workflow.ExecuteActivity(ctx, HelloWorldActivity, input).Get(ctx, &greetingMsg)
	if err != nil {
		logger.Error("HelloWorldActivity failed", zap.Error(err))
		return "", err
	}

	logger.Info("Workflow result", zap.String("greeting", greetingMsg))
	return greetingMsg, nil
}

// This is the activity function
// Given an input, HelloWorldActivity returns "Hello <input>!"
func HelloWorldActivity(ctx context.Context, input sampleInput) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("HelloWorldActivity started")
	return fmt.Sprintf("Hello, %s!", input.Message), nil
}
