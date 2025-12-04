package main

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// ParentWorkflow demonstrates invoking a child workflow from a parent.
// The parent waits for the child to complete before finishing.
func ParentWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("ParentWorkflow started")

	execution := workflow.GetInfo(ctx).WorkflowExecution
	// Parent can specify its own ID for child execution
	childID := fmt.Sprintf("child_workflow:%v", execution.RunID)

	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:                   childID,
		ExecutionStartToCloseTimeout: time.Minute,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	// Execute child workflow and wait for result
	var result string
	err := workflow.ExecuteChildWorkflow(ctx, ChildWorkflow, 0, 5).Get(ctx, &result)
	if err != nil {
		logger.Error("Child workflow failed", zap.Error(err))
		return err
	}

	logger.Info("ParentWorkflow completed", zap.String("result", result))
	return nil
}

// ChildWorkflow demonstrates ContinueAsNew pattern.
// It runs multiple times, restarting itself with ContinueAsNew until runCount reaches 0.
func ChildWorkflow(ctx workflow.Context, totalCount, runCount int) (string, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ChildWorkflow started", zap.Int("totalCount", totalCount), zap.Int("runCount", runCount))

	if runCount <= 0 {
		logger.Error("Invalid run count", zap.Int("runCount", runCount))
		return "", errors.New("invalid run count")
	}

	totalCount++
	runCount--

	if runCount == 0 {
		result := fmt.Sprintf("Child workflow completed after %d runs", totalCount)
		logger.Info("ChildWorkflow completed", zap.String("result", result))
		return result, nil
	}

	// ContinueAsNew: start a new run with fresh history
	logger.Info("ChildWorkflow continuing as new", zap.Int("remainingRuns", runCount))
	return "", workflow.NewContinueAsNewError(ctx, ChildWorkflow, totalCount, runCount)
}

