package batch

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/cadence/testsuite"
)

func Test_BatchWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()

	env.RegisterWorkflow(BatchWorkflow)
	env.RegisterActivity(BatchActivity)

	env.ExecuteWorkflow(BatchWorkflow, BatchWorkflowInput{
		Concurrency: 2,
		TotalSize:   10,
	})

	assert.True(t, env.IsWorkflowCompleted())
	assert.Nil(t, env.GetWorkflowError())
}
