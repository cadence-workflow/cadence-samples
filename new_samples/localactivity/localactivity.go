package main

import (
	"context"
	"strings"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// Condition checkers - these run as local activities
var conditionCheckers = []func(context.Context, string) (bool, error){
	checkCondition0,
	checkCondition1,
	checkCondition2,
}

// LocalActivityWorkflow demonstrates local activities for fast condition checking.
// Local activities run in the worker process without server round-trips.
func LocalActivityWorkflow(ctx workflow.Context, data string) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("LocalActivityWorkflow started", zap.String("data", data))

	// Local activity options - short timeout since they run locally
	lao := workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: time.Second,
	}
	ctx = workflow.WithLocalActivityOptions(ctx, lao)

	// Regular activity options
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var activityFutures []workflow.Future

	// Use local activities to quickly check conditions
	for i, checker := range conditionCheckers {
		var conditionMet bool
		err := workflow.ExecuteLocalActivity(ctx, checker, data).Get(ctx, &conditionMet)
		if err != nil {
			return "", err
		}

		logger.Info("Condition checked", zap.Int("condition", i), zap.Bool("met", conditionMet))

		// Only schedule regular activity if condition is met
		if conditionMet {
			f := workflow.ExecuteActivity(ctx, ProcessActivity, i)
			activityFutures = append(activityFutures, f)
		}
	}

	// Collect results from activities that were scheduled
	var result string
	for _, f := range activityFutures {
		var activityResult string
		if err := f.Get(ctx, &activityResult); err != nil {
			return "", err
		}
		result += activityResult + " "
	}

	logger.Info("LocalActivityWorkflow completed", zap.String("result", result))
	return result, nil
}

// Local activity functions - these run in worker process, not scheduled through server

func checkCondition0(ctx context.Context, data string) (bool, error) {
	return strings.Contains(data, "_0_"), nil
}

func checkCondition1(ctx context.Context, data string) (bool, error) {
	return strings.Contains(data, "_1_"), nil
}

func checkCondition2(ctx context.Context, data string) (bool, error) {
	return strings.Contains(data, "_2_"), nil
}

// ProcessActivity is a regular activity that processes data for a matched condition.
func ProcessActivity(ctx context.Context, conditionID int) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ProcessActivity running", zap.Int("conditionID", conditionID))

	// Simulate processing
	time.Sleep(time.Second)

	return "processed_" + string(rune('0'+conditionID)), nil
}

