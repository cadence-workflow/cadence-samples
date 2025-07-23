package main

import (
	"fmt"
	"testing"
	"time"

	jira "github.com/andygrunwald/go-jira"
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
func (s *UnitTestSuite) Test_getJiraTasksActivity_Success() {
	mockEnvVars := map[string]string{
		"JIRA_API_TOKEN": "mockToken",
	}

	s.env.OnActivity(getEnvVars).Return(mockEnvVars, nil).Once()

	mockJiraClient := new(mockJiraClient)
	mockIssues := []jira.Issue{
		{
			Key: "TEST-1",
			Fields: &jira.IssueFields{
				Summary:     "Test issue 1",
				Description: "Description for test issue 1",
			},
		},
		{
			Key: "TEST-2",
			Fields: &jira.IssueFields{
				Summary:     "Test issue 2",
				Description: "Description for test issue 2",
			},
		},
	}

	s.env.OnActivity(createGithubIssuesActivity, mock.Anything, mock.Anything).Return(nil).Once()
	mockJiraClient.On("Search", mock.Anything, mock.Anything).Return(mockIssues, &jira.Response{}, nil)

	// Replace the actual jiraClient with the mock client
	jiraClient = mockJiraClient

	// Execute the activity
	s.env.ExecuteWorkflow(func(ctx workflow.Context) error {
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
		})
		var issues []jira.Issue
		err := workflow.ExecuteActivity(ctx, getJiraTasksActivity).Get(ctx, &issues)
		s.NoError(err)
		s.Equal(2, len(issues))
		s.Equal("TEST-1", issues[0].Key)
		s.Equal("TEST-2", issues[1].Key)
		return nil
	})

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
}

func (s *UnitTestSuite) Test_getJiraTasksActivity_AuthFailure() {
	mockEnvVars := map[string]string{
		"JIRA_API_TOKEN": "mockToken",
	}

	s.env.OnActivity(getEnvVars).Return(mockEnvVars, nil).Once()

	mockJiraClient := new(mockJiraClient)
	mockJiraClient.On("Search", mock.Anything, mock.Anything).Return(nil, nil, fmt.Errorf("authentication failed"))

	// Replace the actual jiraClient with the mock client
	jiraClient = mockJiraClient

	// Execute the activity
	s.env.ExecuteWorkflow(func(ctx workflow.Context) error {
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
		})
		var issues []jira.Issue
		err := workflow.ExecuteActivity(ctx, getJiraTasksActivity).Get(ctx, &issues)
		s.Error(err)
		s.Contains(err.Error(), "authentication failed")
		s.Nil(issues)
		return nil
	})

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}

type mockJiraClient struct {
	mock.Mock
}

func (m *mockJiraClient) Search(jql string, options *jira.SearchOptions) ([]jira.Issue, *jira.Response, error) {
	args := m.Called(jql, options)
	return args.Get(0).([]jira.Issue), args.Get(1).(*jira.Response), args.Error(2)
}

var jiraClient *mockJiraClient

func (s *UnitTestSuite) Test_GetJiraTasksActivity() {
	mockEnvVars := map[string]string{
		"JIRA_API_TOKEN": "mockToken",
	}

	s.env.OnActivity(getEnvVars).Return(mockEnvVars, nil).Once()

	mockJiraClient := new(mockJiraClient)
	mockIssues := []jira.Issue{
		{
			Key: "TEST-1",
			Fields: &jira.IssueFields{
				Summary:     "Test issue 1",
				Description: "Description for test issue 1",
			},
		},
		{
			Key: "TEST-2",
			Fields: &jira.IssueFields{
				Summary:     "Test issue 2",
				Description: "Description for test issue 2",
			},
		},
	}

	mockJiraClient.On("Search", mock.Anything, mock.Anything).Return(mockIssues, &jira.Response{}, nil)

	// Replace the actual jiraClient with the mock client
	jiraClient = mockJiraClient

	// Execute the activity
	// ctx := context.Background()

	// issues, err := getJiraTasksActivity(ctx)
	// var issues []jira.Issue
	s.env.ExecuteWorkflow(jiraToGithubWorkflow)
	// s.Error(err)

	// Assertions
	// s.NoError(err)
	// s.Equal(2, len(issues))
	// s.Equal("TEST-1", issues[0].Key)
	// s.Equal("TEST-2", issues[1].Key)
}

func (s *UnitTestSuite) Test_getJiraTasksActivity_Failure() {
	// envVars := map[string]string{
	// 	"JIRA_API_TOKEN": "dummy_token",
	// }

	// s.env.OnActivity(getEnvVars).Return(envVars, nil).Once()
	s.env.OnActivity(getJiraTasksActivity, mock.Anything).Return(nil, fmt.Errorf("failed to fetch JIRA tasks")).Once()

	s.env.ExecuteWorkflow(func(ctx workflow.Context) error {
		ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Minute,
			StartToCloseTimeout:    time.Minute,
		})
		var issues []jira.Issue
		err := workflow.ExecuteActivity(ctx, getJiraTasksActivity).Get(ctx, &issues)
		s.Error(err)
		s.Contains(err.Error(), "failed to fetch JIRA tasks")
		s.Nil(issues)
		return nil
	})

	s.True(s.env.IsWorkflowCompleted())
	s.Error(s.env.GetWorkflowError())
}
func (s *UnitTestSuite) SetupTest() {
	s.env = s.NewTestWorkflowEnvironment()
	s.env.RegisterWorkflow(jiraToGithubWorkflow)
	s.env.RegisterActivity(createGithubIssuesActivity)
	s.env.RegisterActivity(getJiraTasksActivity)
}

func (s *UnitTestSuite) TearDownTest() {
	s.env.AssertExpectations(s.T())
}

func (s *UnitTestSuite) Test_WorkflowWithMockActivities() {
	mockIssues := []jira.Issue{
		{
			Key: "TEST-1",
			Fields: &jira.IssueFields{
				Summary:     "Test issue 1",
				Description: "Description for test issue 1",
			},
		},
		{
			Key: "TEST-2",
			Fields: &jira.IssueFields{
				Summary:     "Test issue 2",
				Description: "Description for test issue 2",
			},
		},
	}

	s.env.OnActivity(getJiraTasksActivity, mock.Anything).Return(mockIssues, nil).Once()
	// s.env.OnActivity(createGithubIssuesActivity, mock.Anything, mock.Anything).Return(nil).Once()

	s.env.ExecuteWorkflow(jiraToGithubWorkflow)

	s.True(s.env.IsWorkflowCompleted())
	s.NoError(s.env.GetWorkflowError())
	var workflowResult string
	err := s.env.GetWorkflowResult(&workflowResult)
	s.NoError(err)
	s.Equal("COMPLETED", workflowResult)
}

func (s *UnitTestSuite) Test_WorkflowWithTimeout() {
	mockIssues := []jira.Issue{
		{
			Key: "TEST-1",
			Fields: &jira.IssueFields{
				Summary:     "Test issue 1",
				Description: "Description for test issue 1",
			},
		},
		{
			Key: "TEST-2",
			Fields: &jira.IssueFields{
				Summary:     "Test issue 2",
				Description: "Description for test issue 2",
			},
		},
	}
	s.env.OnActivity(getJiraTasksActivity, mock.Anything).Return(mockIssues, nil).Once()

	s.env.SetWorkflowTimeout(time.Millisecond * 2)
	s.env.SetTestTimeout(time.Minute * 10)

	s.env.ExecuteWorkflow(jiraToGithubWorkflow)

	var workflowResult string
	err := s.env.GetWorkflowResult(&workflowResult)
	s.Equal("TimeoutType: SCHEDULE_TO_CLOSE", err.Error())
	s.Empty(workflowResult)
}
