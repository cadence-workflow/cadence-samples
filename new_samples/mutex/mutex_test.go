package main

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/testsuite"
	"go.uber.org/cadence/workflow"
)

type UnitTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestWorkflowEnvironment
}

func TestUnitTestSuite(t *testing.T) {
	suite.Run(t, new(UnitTestSuite))
}

func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterWorkflow(mutexWorkflow)
	s.env.RegisterWorkflow(sampleWorkflowWithMutex)
	s.env.RegisterActivity(signalWithStartMutexWorkflowActivity)
}

func (s *UnitTestSuite) TearDownTest() {
	s.env.AssertExpectations(s.T())
}

// mockMutexLock stubs the signalWithStartMutexWorkflowActivity and immediately
// grants the lock by sending the AcquireLock signal.
func mockMutexLock(env *testsuite.TestWorkflowEnvironment, resourceID string, mockError error) {
	mockExecution := &workflow.Execution{ID: "mockID", RunID: "mockRunID"}
	env.OnActivity(signalWithStartMutexWorkflowActivity,
		mock.Anything, mock.Anything, resourceID, mock.Anything, mock.Anything).
		Return(mockExecution, mockError)
	env.RegisterDelayedCallback(func() {
		env.SignalWorkflow(AcquireLockSignalName, "mockReleaseLockChannelName")
	}, 0)
	if mockError == nil {
		env.OnSignalExternalWorkflow(mock.Anything, mock.Anything, mockExecution.RunID,
			mock.Anything, mock.Anything).Return(nil)
	}
}

func (s *UnitTestSuite) Test_Workflow_Success() {
	mockMutexLock(s.env, "mockResourceID", nil)
	s.env.ExecuteWorkflow(sampleWorkflowWithMutex, "mockResourceID")
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_Workflow_Error() {
	mockMutexLock(s.env, "mockResourceID", errors.New("bad-error"))
	s.env.ExecuteWorkflow(sampleWorkflowWithMutex, "mockResourceID")
	s.True(s.env.IsWorkflowCompleted())
	s.EqualError(s.env.GetWorkflowError(), "bad-error")
}

func (s *UnitTestSuite) Test_MutexWorkflow_Success() {
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(RequestLockSignalName, "mockSenderWorkflowID")
	}, 0)
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow("unlock-event-mockSenderWorkflowID", "releaseLock")
	}, 0)
	s.env.OnSignalExternalWorkflow(mock.Anything, "mockSenderWorkflowID", "",
		AcquireLockSignalName, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(mutexWorkflow, "mockNamespace", "mockResourceID", 10*time.Minute)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_MutexWorkflow_TimeoutSuccess() {
	s.env.RegisterDelayedCallback(func() {
		s.env.SignalWorkflow(RequestLockSignalName, "mockSenderWorkflowID")
	}, 0)
	s.env.OnSignalExternalWorkflow(mock.Anything, "mockSenderWorkflowID", "",
		AcquireLockSignalName, mock.Anything).Return(nil)

	s.env.ExecuteWorkflow(mutexWorkflow, "mockNamespace", "mockResourceID", 10*time.Minute)
	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}
