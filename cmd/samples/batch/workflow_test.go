package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/cadence/testsuite"
)

func Test_BatchWorkflow(t *testing.T) {
	// Create test environment for workflow testing
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	// Register the workflow and activity functions
	env.RegisterWorkflow(BatchWorkflow)
	env.RegisterActivity(BatchActivity)

	// Execute workflow with 3 concurrent workers processing 10 tasks
	env.ExecuteWorkflow(BatchWorkflow, BatchWorkflowInput{
		Concurrency: 3,
		TotalSize:   10,
	})

	// Assert workflow completed successfully without errors
	assert.True(t, env.IsWorkflowCompleted())
	assert.Nil(t, env.GetWorkflowError())
}
