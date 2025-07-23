package main

import (
	"context"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type MyPayload struct {
	Msg   string
	Count int
}

const DataConverterWorkflowName = "dataConverterWorkflow"

func dataConverterWorkflow(ctx workflow.Context, input MyPayload) (MyPayload, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Workflow started", zap.Any("input", input))

	activityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result MyPayload
	err := workflow.ExecuteActivity(ctx, dataConverterActivity, input).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed", zap.Error(err))
		return MyPayload{}, err
	}
	logger.Info("Workflow completed", zap.Any("result", result))
	return result, nil
}

func dataConverterActivity(ctx context.Context, input MyPayload) (MyPayload, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Activity received input", zap.Any("input", input))
	input.Msg = input.Msg + " processed"
	input.Count++
	logger.Info("Activity returning", zap.Any("output", input))
	return input, nil
}
