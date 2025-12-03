// THIS IS A GENERATED FILE
// PLEASE DO NOT EDIT

package main

import (
	"github.com/uber-common/cadence-samples/new_samples/common"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// StartWorker creates and starts a Cadence worker.
func StartWorker() {
	// Create Cadence client - all gRPC/YARPC boilerplate is handled by the helper
	cadenceClient := common.MustNewCadenceClient(
		common.DefaultTaskList,
		common.DefaultHostPort,
	)

	w := worker.New(
		cadenceClient,
		common.DefaultDomain,
		common.DefaultTaskList,
		worker.Options{},
	)

	// Register workflows
	w.RegisterWorkflowWithOptions(SimpleSignalWorkflow, workflow.RegisterOptions{Name: "cadence_samples.SimpleSignalWorkflow"})

	// Register activities
	w.RegisterActivityWithOptions(SimpleSignalActivity, activity.RegisterOptions{Name: "cadence_samples.SimpleSignalActivity"})

	if err := w.Start(); err != nil {
		panic("Failed to start worker: " + err.Error())
	}

	logger, _ := zap.NewDevelopment()
	logger.Info("Started Worker.", zap.String("taskList", common.DefaultTaskList))
}
