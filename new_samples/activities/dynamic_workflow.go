package workflows

import (
	"context"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"time"
)

const DynamicGreetingActivityName = "cadence_samples.DynamicGreetingActivity"

type dynamicWorkflowInput struct {
	Message string `json:"message"`
}

func DynamicWorkflow(ctx workflow.Context, input dynamicWorkflowInput) (string, error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("DynamicWorkflow started")

	var greetingMsg string
	err := workflow.ExecuteActivity(ctx, DynamicGreetingActivityName, input.Message).Get(ctx, &greetingMsg)
	if err != nil {
		logger.Error("DynamicGreetingActivity failed", zap.Error(err))
		return "", err
	}

	logger.Info("Workflow result", zap.String("greeting", greetingMsg))
	return greetingMsg, nil
}

func DynamicGreetingActivity(ctx context.Context, message string) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("DynamicGreetingActivity started.")
	return "Hello, " + message, nil
}
