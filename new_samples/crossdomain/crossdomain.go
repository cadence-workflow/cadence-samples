package main

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// Configuration for cross-domain execution
const (
	ChildDomain   = "child-domain"     // Must be registered separately
	ChildTaskList = "child-task-list"
)

// CrossDomainData is passed between workflows
type CrossDomainData struct {
	Value string
}

// CrossDomainWorkflow demonstrates executing child workflows in different domains.
func CrossDomainWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("CrossDomainWorkflow started")

	// Execute child workflow in a different domain
	childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
		Domain:                       ChildDomain,
		WorkflowID:                   "child-wf-" + uuid.New().String(),
		TaskList:                     ChildTaskList,
		ExecutionStartToCloseTimeout: time.Minute,
	})

	err := workflow.ExecuteChildWorkflow(childCtx, ChildDomainWorkflow, CrossDomainData{Value: "test"}).Get(ctx, nil)
	if err != nil {
		logger.Error("Child workflow failed", zap.Error(err))
		return err
	}

	logger.Info("CrossDomainWorkflow completed")
	return nil
}

// ChildDomainWorkflow runs in the child domain.
func ChildDomainWorkflow(ctx workflow.Context, data CrossDomainData) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("ChildDomainWorkflow started", zap.String("value", data.Value))

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, ChildDomainActivity).Get(ctx, nil)
	if err != nil {
		return err
	}

	logger.Info("ChildDomainWorkflow completed")
	return nil
}

// ChildDomainActivity runs in the child domain.
func ChildDomainActivity(ctx context.Context) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("ChildDomainActivity running")
	return "Hello from child domain!", nil
}

