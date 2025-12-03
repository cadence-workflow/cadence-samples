package main

import (
	"github.com/google/uuid"
	"go.uber.org/cadence/workflow"
)

// SideEffectWorkflow demonstrates the SideEffect API for non-deterministic operations.
// SideEffect ensures the result is recorded and replayed deterministically.
func SideEffectWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("SideEffectWorkflow started")

	value := ""

	// Register query handler to inspect the value
	err := workflow.SetQueryHandler(ctx, "value", func(input []byte) (string, error) {
		return value, nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed: " + err.Error())
		return err
	}

	// SideEffect records the result of non-deterministic operations
	// On replay, it returns the recorded value instead of re-executing
	workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
		return uuid.New().String()
	}).Get(&value)

	logger.Info("SideEffect value: " + value)
	logger.Info("SideEffectWorkflow completed")
	return nil
}

