package main

import (
	"testing"

	"go.uber.org/cadence/worker"
	"go.uber.org/zap/zaptest"

	"github.com/stretchr/testify/require"
)

// This replay test is the recommended way to make sure changing workflow code is backward compatible without non-deterministic errors.
// "greetings.json" can be downloaded from cadence CLI:
//
//	cadence --do default wf show -w greetings_5d5f8e5c-4807-444d-9dc5-80abea22a324 --output_filename ~/tmp/greetings.json
//
// Or from Cadence Web UI. And you may need to change workflowType in the first event.
func TestReplayWorkflowHistoryFromFile(t *testing.T) {
	replayer := worker.NewWorkflowReplayer()

	replayer.RegisterWorkflow(sampleGreetingsWorkflow)

	err := replayer.ReplayWorkflowHistoryFromJSONFile(zaptest.NewLogger(t), "greetings.json")
	require.NoError(t, err)
}
