package main

import (
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// GreetingsWorkflow executes activities sequentially.
func GreetingsWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("GreetingsWorkflow started")

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var greeting string
	if err := workflow.ExecuteActivity(ctx, GetGreetingActivity).Get(ctx, &greeting); err != nil {
		return err
	}

	var name string
	if err := workflow.ExecuteActivity(ctx, GetNameActivity).Get(ctx, &name); err != nil {
		return err
	}

	var result string
	if err := workflow.ExecuteActivity(ctx, SayGreetingActivity, greeting, name).Get(ctx, &result); err != nil {
		return err
	}

	logger.Info("GreetingsWorkflow completed", zap.String("result", result))
	return nil
}

func GetGreetingActivity() (string, error) { return "Hello", nil }
func GetNameActivity() (string, error)     { return "Cadence", nil }
func SayGreetingActivity(greeting, name string) (string, error) {
	return fmt.Sprintf("%s %s!", greeting, name), nil
}

