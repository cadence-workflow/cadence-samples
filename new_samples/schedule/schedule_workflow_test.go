package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/testsuite"
)

func Test_ScheduledWorkflow_Completes(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(scheduledWorkflow)
	env.RegisterActivity(scheduledActivity)

	env.ExecuteWorkflow(scheduledWorkflow, 0)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())
}

func Test_ScheduledWorkflow_ExecutesActivity(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(scheduledWorkflow)
	env.RegisterActivity(scheduledActivity)

	activityCalled := false
	env.SetOnActivityCompletedListener(func(activityInfo *activity.Info, result encoded.Value, err error) {
		activityCalled = true
	})

	env.ExecuteWorkflow(scheduledWorkflow, 0)

	assert.True(t, env.IsWorkflowCompleted())
	assert.NoError(t, env.GetWorkflowError())
	assert.True(t, activityCalled, "scheduledActivity must be called by the workflow")
}
