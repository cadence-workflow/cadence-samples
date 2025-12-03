package main

import (
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// GreetingsWorkflow demonstrates sequential activity execution.
// It executes 3 activities in sequence, passing results from one to the next.
func GreetingsWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("GreetingsWorkflow started")

	// Step 1: Get greeting
	var greeting string
	err := workflow.ExecuteActivity(ctx, GetGreetingActivity).Get(ctx, &greeting)
	if err != nil {
		logger.Error("GetGreetingActivity failed", zap.Error(err))
		return err
	}
	logger.Info("Got greeting", zap.String("greeting", greeting))

	// Step 2: Get name
	var name string
	err = workflow.ExecuteActivity(ctx, GetNameActivity).Get(ctx, &name)
	if err != nil {
		logger.Error("GetNameActivity failed", zap.Error(err))
		return err
	}
	logger.Info("Got name", zap.String("name", name))

	// Step 3: Combine greeting and name
	var result string
	err = workflow.ExecuteActivity(ctx, SayGreetingActivity, greeting, name).Get(ctx, &result)
	if err != nil {
		logger.Error("SayGreetingActivity failed", zap.Error(err))
		return err
	}

	logger.Info("Workflow completed", zap.String("result", result))
	return nil
}

// GetGreetingActivity returns a greeting word.
func GetGreetingActivity() (string, error) {
	return "Hello", nil
}

// GetNameActivity returns a name.
func GetNameActivity() (string, error) {
	return "Cadence", nil
}

// SayGreetingActivity combines greeting and name into a full greeting message.
func SayGreetingActivity(greeting string, name string) (string, error) {
	result := fmt.Sprintf("%s %s!", greeting, name)
	return result, nil
}

