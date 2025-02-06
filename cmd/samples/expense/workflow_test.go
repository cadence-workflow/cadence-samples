package main

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
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
	s.env.RegisterWorkflow(sampleExpenseWorkflow)
	s.env.RegisterActivity(createExpenseActivity)
	s.env.RegisterActivity(waitForDecisionActivity)
	s.env.RegisterActivity(paymentActivity)
}

func (s *UnitTestSuite) TearDownTest() {
	s.env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_WorkflowWithMockActivities() {
	s.env.OnActivity(createExpenseActivity, mock.Anything, mock.Anything).Return(nil).Once()
	s.env.OnActivity(waitForDecisionActivity, mock.Anything, mock.Anything).Return("APPROVED", nil).Once()
	s.env.OnActivity(paymentActivity, mock.Anything, mock.Anything).Return(nil).Once()

	s.env.ExecuteWorkflow(sampleExpenseWorkflow, "test-expense-id")

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	var workflowResult string
	err := s.env.GetWorkflowResult(&workflowResult)
	s.NoError(err)
	s.Equal("COMPLETED", workflowResult)
}

func (s *UnitTestSuite) Test_TimeoutWithMockActivities() {
	s.env.OnActivity(createExpenseActivity, mock.Anything, mock.Anything).Return(nil).Once()
	s.env.SetWorkflowTimeout(time.Microsecond * 500)
	s.env.SetTestTimeout(time.Minute * 10)

	s.env.ExecuteWorkflow(sampleExpenseWorkflow, "test-expense-id")

	var workflowResult string
	err := s.env.GetWorkflowResult(&workflowResult)
	s.Equal("TimeoutType: SCHEDULE_TO_CLOSE", err.Error())
	s.Empty(workflowResult)
}

func (s *UnitTestSuite) Test_WorkflowStatusRejected() {
	s.env.OnActivity(createExpenseActivity, mock.Anything, mock.Anything).Return(nil).Once()
	s.env.OnActivity(waitForDecisionActivity, mock.Anything, mock.Anything).Return("REJECTED", nil).Once()

	s.env.ExecuteWorkflow(sampleExpenseWorkflow, "test-expense-id")

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	var workflowResult string
	err := s.env.GetWorkflowResult(&workflowResult)
	s.NoError(err)
	s.Empty(workflowResult)
}

func (s *UnitTestSuite) Test_WorkflowStatusCancelled() {
	s.env.OnActivity(createExpenseActivity, mock.Anything, mock.Anything).Return(nil).Once()
	s.env.OnActivity(waitForDecisionActivity, mock.Anything, mock.Anything).Return("CANCELLED", nil).Once()

	s.env.ExecuteWorkflow(sampleExpenseWorkflow, "test-expense-id")

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	var workflowResult string
	err := s.env.GetWorkflowResult(&workflowResult)
	s.NoError(err)
	s.Empty(workflowResult)
}

func (s *UnitTestSuite) Test_WorkflowStatusApprovedWithPaymentError() {
	s.env.OnActivity(createExpenseActivity, mock.Anything, mock.Anything).Return(nil).Once()
	s.env.OnActivity(waitForDecisionActivity, mock.Anything, mock.Anything).Return("APPROVED", nil).Once()
	s.env.OnActivity(paymentActivity, mock.Anything, mock.Anything).Return(errors.New("payment error")).Once()

	s.env.ExecuteWorkflow(sampleExpenseWorkflow, "test-expense-id")

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
	var workflowResult string
	err := s.env.GetWorkflowResult(&workflowResult)
	s.Equal("payment error", err.Error())
	s.Empty(workflowResult)
}

func (s *UnitTestSuite) Test_CreateActivityFailed() {
	s.env.OnActivity(createExpenseActivity, mock.Anything, mock.Anything).Return(errors.New("expense id is empty")).Once()

	s.env.ExecuteWorkflow(sampleExpenseWorkflow, "")

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
	var workflowResult string

	err := s.env.GetWorkflowResult(&workflowResult)
	s.Equal("expense id is empty", err.Error())
	s.Empty(workflowResult)
}

func (s *UnitTestSuite) Test_WaitForDecisionActivityFailed() {
	s.env.OnActivity(createExpenseActivity, mock.Anything, mock.Anything).Return(nil).Once()
	s.env.OnActivity(waitForDecisionActivity, mock.Anything, mock.Anything).Return("", errors.New("failed to get decision")).Once()

	s.env.ExecuteWorkflow(sampleExpenseWorkflow, "test-expense-id")

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
	var workflowResult string

	err := s.env.GetWorkflowResult(&workflowResult)
	s.Equal("failed to get decision", err.Error())
	s.Empty(workflowResult)
}

func (s *UnitTestSuite) Test_PaymentActivityFailed() {
	s.env.OnActivity(createExpenseActivity, mock.Anything, mock.Anything).Return(nil).Once()
	s.env.OnActivity(waitForDecisionActivity, mock.Anything, mock.Anything).Return("APPROVED", nil).Once()
	s.env.OnActivity(paymentActivity, mock.Anything, mock.Anything).Return(errors.New("payment failed")).Once()

	s.env.ExecuteWorkflow(sampleExpenseWorkflow, "test-expense-id")

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
	var workflowResult string

	err := s.env.GetWorkflowResult(&workflowResult)
	s.Equal("payment failed", err.Error())
	s.Empty(workflowResult)
}

func (s *UnitTestSuite) Test_WorkflowWithMockServer() {
	// setup mock expense server
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/text")
		switch r.URL.Path {
		case "/create":
		case "/registerCallback":
			taskToken := []byte(r.PostFormValue("task_token"))
			// simulate the expense is approved one hour later.
			s.env.RegisterDelayedCallback(func() {
				s.env.CompleteActivity(taskToken, "APPROVED", nil)
			}, time.Hour)
		case "/action":
		}
		io.WriteString(w, "SUCCEED")
	}
	server := httptest.NewServer(http.HandlerFunc(handler))
	defer server.Close()

	// pointing server to test mock
	expenseServerHostPort = server.URL

	s.env.ExecuteWorkflow(sampleExpenseWorkflow, "test-expense-id")

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	var workflowResult string
	err := s.env.GetWorkflowResult(&workflowResult)
	s.NoError(err)
	s.Equal("COMPLETED", workflowResult)
}
