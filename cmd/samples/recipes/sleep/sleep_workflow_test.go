package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/testsuite"
)

func Test_Sleep(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}

	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(sleepWorkflow)
	env.RegisterActivity(mainSleepActivity)

	var activityMessage string
	env.SetOnActivityCompletedListener(func(activityInfo *activity.Info, result encoded.Value, err error) {
		result.Get(&activityMessage)
	})

	sleepDuration := 5 * time.Second
	env.ExecuteWorkflow(sleepWorkflow, sleepDuration)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
	require.Equal(t, "Main sleep activity completed", activityMessage)
}
