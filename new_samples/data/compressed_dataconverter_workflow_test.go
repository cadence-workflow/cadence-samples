package main

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/worker"
)

func Test_CompressionDataConverterWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(CompressionDataConverterWorkflow)
	env.RegisterActivity(CompressionDataConverterActivity)

	dataConverter := NewCompressedJSONDataConverter()
	workerOptions := worker.Options{
		DataConverter: dataConverter,
	}
	env.SetWorkerOptions(workerOptions)

	var activityResult LargePayload
	env.SetOnActivityCompletedListener(func(activityInfo *activity.Info, result encoded.Value, err error) {
		result.Get(&activityResult)
	})

	// Workflow generates its own payload internally, no input needed
	env.ExecuteWorkflow(CompressionDataConverterWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, "Comprehensive Product Catalog (Processed)", activityResult.Name)
	require.Equal(t, 100, activityResult.Stats.TotalItems)
}
