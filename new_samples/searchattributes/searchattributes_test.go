package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/testsuite"
)

func Test_SearchAttributesWorkflow(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}
	env := testSuite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(searchAttributesWorkflow)
	env.RegisterActivity(listExecutions)

	env.SetSearchAttributesOnStart(getSearchAttributesForStart())

	env.OnUpsertSearchAttributes(map[string]interface{}{
		"CustomIntField":      2,
		"CustomKeywordField":  "Update1",
		"CustomBoolField":     true,
		"CustomDoubleField":   3.14,
		"CustomDatetimeField": time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local),
		"CustomStringField":   "String field is for full-text search. It is tokenized and cannot be used in ORDER BY.",
	}).Return(nil).Once()

	env.OnUpsertSearchAttributes(map[string]interface{}{
		"CustomKeywordField": "Update2",
	}).Return(nil).Once()

	env.OnActivity(listExecutions, mock.Anything, mock.Anything).
		Return([]*shared.WorkflowExecutionInfo{{}}, nil).Once()

	env.ExecuteWorkflow(searchAttributesWorkflow)

	require.True(t, env.IsWorkflowCompleted())
	require.NoError(t, env.GetWorkflowError())
}
