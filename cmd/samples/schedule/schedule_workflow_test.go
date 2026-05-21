package main

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/testsuite"
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
	s.env.RegisterWorkflow(scheduledWorkflow)
	s.env.RegisterActivity(scheduledActivity)
}

func (s *UnitTestSuite) TearDownTest() {
	s.env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_ScheduledWorkflow_Completes() {
	s.env.ExecuteWorkflow(scheduledWorkflow)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_ScheduledWorkflow_ExecutesActivity() {
	activityCalled := false
	s.env.SetOnActivityCompletedListener(func(activityInfo *activity.Info, result encoded.Value, err error) {
		activityCalled = true
	})

	s.env.ExecuteWorkflow(scheduledWorkflow)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	s.True(activityCalled, "scheduledActivity must be called by the workflow")
}
