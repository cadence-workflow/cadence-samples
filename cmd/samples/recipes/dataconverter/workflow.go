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

// LargeDataConverterWorkflowName is the workflow name for large payload processing
const LargeDataConverterWorkflowName = "largeDataConverterWorkflow"

// largeDataConverterWorkflow demonstrates processing large payloads with compression
func largeDataConverterWorkflow(ctx workflow.Context, input LargePayload) (LargePayload, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Large payload workflow started", zap.String("payload_id", input.ID))
	logger.Info("Processing large payload with compression", zap.Int("items_count", len(input.Items)))

	activityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result LargePayload
	err := workflow.ExecuteActivity(ctx, largeDataConverterActivity, input).Get(ctx, &result)
	if err != nil {
		logger.Error("Large payload activity failed", zap.Error(err))
		return LargePayload{}, err
	}

	logger.Info("Large payload workflow completed", zap.String("result_id", result.ID))
	logger.Info("Note: All large payload data was automatically compressed/decompressed using gzip compression")
	return result, nil
}

func largeDataConverterActivity(ctx context.Context, input LargePayload) (LargePayload, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Large payload activity received input", zap.String("payload_id", input.ID), zap.Int("items_count", len(input.Items)))

	// Process the large payload (in a real scenario, this might involve data transformation, validation, etc.)
	input.Name = input.Name + " (Processed)"
	input.Stats.TotalItems = len(input.Items)

	logger.Info("Large payload activity completed", zap.String("result_id", input.ID))
	return input, nil
}
