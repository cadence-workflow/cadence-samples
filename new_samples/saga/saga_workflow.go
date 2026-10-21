package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// NumberOfSteps is how many forward steps SagaWorkflow runs.
const NumberOfSteps = 3

// SagaInput controls which path the workflow takes.
type SagaInput struct {
	// FailAtStep is the step that fails (1 to NumberOfSteps), or 0 for the success path.
	FailAtStep int
}

// SagaWorkflow runs the steps in order. If a step fails, it compensates the
// completed steps in reverse order and returns the error.
func SagaWorkflow(ctx workflow.Context, input SagaInput) error {
	if input.FailAtStep < 0 || input.FailAtStep > NumberOfSteps {
		return fmt.Errorf("FailAtStep must be between 0 and %d, got %d", NumberOfSteps, input.FailAtStep)
	}

	logger := workflow.GetLogger(ctx)
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	})

	saga := &Saga{}

	for step := 1; step <= NumberOfSteps; step++ {
		shouldFail := step == input.FailAtStep

		err := workflow.ExecuteActivity(ctx, StepActivity, step, shouldFail).Get(ctx, nil)
		if err != nil {
			logger.Error("Step failed, compensating completed steps", zap.Int("step", step), zap.Error(err))
			if compensationErr := saga.Compensate(ctx); compensationErr != nil {
				logger.Error("Compensation failed", zap.Error(compensationErr))
			}
			return err
		}

		// Register only after the step succeeds: a step that never ran has nothing to undo.
		completedStep := step
		saga.AddCompensation(func(ctx workflow.Context) error {
			return workflow.ExecuteActivity(ctx, CompensateStepActivity, completedStep).Get(ctx, nil)
		})
	}

	logger.Info("All steps completed, no compensation needed")
	return nil
}

// StepActivity runs one forward step, or fails it when shouldFail is true.
func StepActivity(ctx context.Context, step int, shouldFail bool) error {
	logger := activity.GetLogger(ctx)
	if shouldFail {
		logger.Info("Step failing on purpose", zap.Int("step", step))
		return cadence.NewCustomError("step-failed")
	}
	logger.Info("Step done", zap.Int("step", step))
	return nil
}

// CompensateStepActivity undoes a completed step.
func CompensateStepActivity(ctx context.Context, step int) error {
	activity.GetLogger(ctx).Info("Step compensated", zap.Int("step", step))
	return nil
}
