package main

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/testsuite"
)

// runSaga executes SagaWorkflow and returns the forward and compensated steps, in order.
func runSaga(t *testing.T, input SagaInput) (env *testsuite.TestWorkflowEnvironment, forward, compensated []int) {
	var suite testsuite.WorkflowTestSuite
	env = suite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(SagaWorkflow)
	env.RegisterActivity(StepActivity)
	env.RegisterActivity(CompensateStepActivity)

	env.SetOnActivityStartedListener(func(info *activity.Info, _ context.Context, args encoded.Values) {
		var step int
		var shouldFail bool
		if strings.HasSuffix(info.ActivityType.Name, "CompensateStepActivity") {
			require.NoError(t, args.Get(&step))
			compensated = append(compensated, step)
			return
		}
		require.NoError(t, args.Get(&step, &shouldFail))
		forward = append(forward, step)
	})

	env.ExecuteWorkflow(SagaWorkflow, input)
	require.True(t, env.IsWorkflowCompleted())
	return env, forward, compensated
}

// requireStepFailed checks that err is the original step error.
func requireStepFailed(t *testing.T, err error) {
	var customErr *cadence.CustomError
	require.True(t, errors.As(err, &customErr), "want *cadence.CustomError, got %T: %v", err, err)
	assert.Equal(t, "step-failed", customErr.Reason())
}

func TestSagaWorkflow_Success(t *testing.T) {
	env, forward, compensated := runSaga(t, SagaInput{FailAtStep: 0})

	assert.NoError(t, env.GetWorkflowError())
	assert.Equal(t, []int{1, 2, 3}, forward)
	assert.Empty(t, compensated)
}

func TestSagaWorkflow_LastStepFails_CompensatesInReverseOrder(t *testing.T) {
	env, forward, compensated := runSaga(t, SagaInput{FailAtStep: 3})

	requireStepFailed(t, env.GetWorkflowError())
	assert.Equal(t, []int{1, 2, 3}, forward)
	assert.Equal(t, []int{2, 1}, compensated, "only completed steps are undone, newest first")
}

func TestSagaWorkflow_FirstStepFails_NothingToCompensate(t *testing.T) {
	env, forward, compensated := runSaga(t, SagaInput{FailAtStep: 1})

	requireStepFailed(t, env.GetWorkflowError())
	assert.Equal(t, []int{1}, forward)
	assert.Empty(t, compensated)
}

func TestSagaWorkflow_InvalidFailAtStep(t *testing.T) {
	for _, failAtStep := range []int{-1, NumberOfSteps + 1} {
		env, forward, compensated := runSaga(t, SagaInput{FailAtStep: failAtStep})

		err := env.GetWorkflowError()
		require.Error(t, err, "FailAtStep=%d", failAtStep)
		assert.Contains(t, err.Error(), "FailAtStep must be between 0 and 3")
		assert.Empty(t, forward, "no step runs on invalid input")
		assert.Empty(t, compensated)
	}
}

func TestSagaWorkflow_FailedCompensationDoesNotStopTheOthers(t *testing.T) {
	var suite testsuite.WorkflowTestSuite
	env := suite.NewTestWorkflowEnvironment()
	env.RegisterWorkflow(SagaWorkflow)
	env.RegisterActivity(StepActivity)
	env.RegisterActivity(CompensateStepActivity)

	env.OnActivity(CompensateStepActivity, mock.Anything, 2).Return(errors.New("undo step 2 failed")).Once()
	env.OnActivity(CompensateStepActivity, mock.Anything, 1).Return(nil).Once()

	env.ExecuteWorkflow(SagaWorkflow, SagaInput{FailAtStep: 3})

	require.True(t, env.IsWorkflowCompleted())
	err := env.GetWorkflowError()
	requireStepFailed(t, err)
	assert.NotContains(t, err.Error(), "undo step 2 failed", "compensation errors are logged, not returned")
	env.AssertExpectations(t)
}
