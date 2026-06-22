package main

import (
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/cadence/encoded"
	"go.uber.org/zap"
)

func buildScheduleClient() client.Client {
	return client.NewClient(BuildCadenceClient(), Domain, nil)
}

func encodeWorkflowInput(logger *zap.Logger, sleepSeconds int) []byte {
	data, err := encoded.GetDefaultDataConverter().ToData(sleepSeconds)
	if err != nil {
		logger.Fatal("Failed to encode workflow input", zap.Error(err))
	}
	return data
}

func startWorkflowAction(logger *zap.Logger) *client.ScheduleStartWorkflowAction {
	return &client.ScheduleStartWorkflowAction{
		WorkflowType:                    scheduledWorkflowName,
		TaskList:                        TaskListName,
		Input:                           encodeWorkflowInput(logger, 0),
		ExecutionStartToCloseTimeout:    10 * time.Minute,
		DecisionTaskStartToCloseTimeout: 10 * time.Second,
	}
}
