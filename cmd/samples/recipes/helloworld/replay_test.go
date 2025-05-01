package main

import (
	"testing"

	"go.uber.org/cadence/worker"
	"go.uber.org/zap/zaptest"

	"github.com/stretchr/testify/require"
)

// This replay test is the recommended way to make sure changing workflow code is backward compatible without non-deterministic errors.
// "helloworld.json" can be downloaded from cadence CLI:
//
//	cadence --do default wf show -w helloworld_d002cd3a-aeee-4a11-aa30-1c62385b4d87 --output_filename ~/tmp/helloworld.json
//
// Or from Cadence Web UI. And you may need to change workflowType in the first event.
func TestReplayWorkflowHistoryFromFile(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()

	replayer.RegisterWorkflow(helloWorldWorkflow)

	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "helloworld.json")
	require.NoError(t, err)
}
