package main

import (
	"time"

	"go.uber.org/cadence/workflow"
)

// QueryWorkflow demonstrates query handlers for inspecting workflow state.
// Query the workflow with: cadence workflow query --wid <id> --qt state
func QueryWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("QueryWorkflow started")

	queryResult := "started"

	// Register query handler for "state" query type
	err := workflow.SetQueryHandler(ctx, "state", func(input []byte) (string, error) {
		return queryResult, nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed: " + err.Error())
		return err
	}

	// Update state and wait on timer
	queryResult = "waiting on timer"
	logger.Info("State changed to: waiting on timer")

	// Wait for 2 minutes (query the workflow while it's waiting!)
	workflow.NewTimer(ctx, time.Minute*2).Get(ctx, nil)
	logger.Info("Timer fired")

	queryResult = "done"
	logger.Info("QueryWorkflow completed")
	return nil
}

