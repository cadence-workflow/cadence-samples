package main

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	// "code.uber.internal/devexp/utils/jirawithretry"
	// "github.com/andygrunwald/go-jira"
	jira "github.com/andygrunwald/go-jira"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	// ApplicationName is the task list for this sample
	ApplicationName = "jiraToGithubGroup"
	jiraURL         = "https://t3.uberinternal.com"
	jiraUsername    = "svc-cadence-jira@uber.com"
)

// type Activity struct {
// 	jiraClient jirawithretry.IssueClient
// }

func jiraToGithubWorkflow(ctx workflow.Context) (result string, err error) {
	// step 1, get JIRA tasks from cadence project
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute * 4,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx1 := workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("Jira to Github workflow started")

	var issues []jira.Issue
	err = workflow.ExecuteActivity(ctx1, getJiraTasksActivity).Get(ctx1, &issues)
	if err != nil {
		return "", err
	}

	// step 2, create issues in github
	ao = workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute * 8,
	}
	ctx2 := workflow.WithActivityOptions(ctx, ao)

	err = workflow.ExecuteActivity(ctx2, createGithubIssuesActivity, issues).Get(ctx2, nil)
	if err != nil {
		return "", err
	}

	logger.Info("Workflow completed with Github issues created.")
	return "COMPLETED", nil
}

// func New() (*Activity, error) {
// 	client, err := jirawithretry.NewIssueClient(&http.Client{}, jiraURL)
// 	if err != nil {
// 		return nil, err
// 	}
// 	envVars, err := getEnvVars()
// 	if err != nil {
// 		return nil, err
// 	}
// 	client.Authentication().SetBasicAuth(jiraUsername, envVars["JIRA_API_TOKEN"])
// 	return &Activity{jiraClient: client}, nil
// }

func getJiraTasksActivity(ctx context.Context) ([]jira.Issue, error) {
	jiraClient, err := jira.NewClient(nil, jiraURL)
	if err != nil {
		return nil, err
	}

	jql := "project = Cadence AND created >= -365d AND issuetype = Task AND labels = opensourceable ORDER BY created DESC" //AND labels = opensourceable AND description IS NOT EMPTY
	options := &jira.SearchOptions{
		MaxResults: 10,
	}
	issues, _, err := jiraClient.Issue.Search(jql, options)
	if err != nil {
		return nil, err
	}

	activity.GetLogger(ctx).Info("JIRA tasks with label opensourceable fetched", zap.Int("Count", len(issues)))
	return issues, err
}

func createGithubIssuesActivity(ctx context.Context, issues []jira.Issue) error {
	// Replace with your actual token
	token := "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"

	// GitHub org and repo
	// org := "cadence-workflow"
	// repo := "cadence-java-samples"
	client := github.NewClient(nil)

	org := "vishwa-test-2"
	repo := "sample"

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client2 := github.NewClient(tc)
	for _, issue := range issues {
		title := issue.Fields.Summary
		// key := issue.Key
		body := issue.Fields.Description

		// Check if the issue already exists in GitHub
		searchOpts := &github.SearchOptions{
			TextMatch: true,
		}
		query := fmt.Sprintf("repo:vishwa-test-2/sample in:title %s state:open", title)
		result, _, err := client.Search.Issues(ctx, query, searchOpts)
		if err != nil {
			return err
		}

		if len(result.Issues) == 0 {
			// Create issue request
			issueRequest := &github.IssueRequest{
				Title: github.String(title),
				Body:  github.String(body),
			}

			// Create issue
			issue, _, err := client2.Issues.Create(ctx, org, repo, issueRequest)
			if err != nil {
				log.Fatalf("Error creating issue: %v", err)
			}

			// req := &github.IssueRequest{
			// 	Title: github.String(fmt.Sprintf("[JIRA %s] %s", key, title)),
			// 	Body:  github.String(body),
			// }
			// _, _, err := client.Issues.Create(ctx, "vishwa-test-2", "sample", req)
			// if err != nil {
			// 	return err
			// }
			activity.GetLogger(ctx).Info("Created an issue in GitHub", zap.String("title", issue.GetHTMLURL()))
		} else {
			activity.GetLogger(ctx).Info("GitHub issue already exists", zap.String("title", title))
		}
	}
	return nil
}

// type IssueRequest struct {
// 	Title string `json:"title"`
// 	Body  string `json:"body"`
// }

// func createGithubIssuesActivity(ctx context.Context, issues []jira.Issue) error {
// 	for _, issue := range issues {
// 		title := issue.Fields.Summary
// 		key := issue.Key
// 		body := issue.Fields.Description

// 		// Check if the issue already exists in GitHub
// 		searchArgs := []string{
// 			"gh",
// 			"issue",
// 			"list",
// 			"--repo",
// 			"cadence-workflow/cadence-java-samples",
// 			"--search",
// 			fmt.Sprintf("[JIRA issue] %s in:title", title),
// 		}

// 		searchOutput, err := exec.Command(searchArgs[0], searchArgs[1:]...).Output()
// 		if err != nil {
// 			return err
// 		}

// 		if len(searchOutput) == 0 {
// Create a new issue in GitHub

// issueData, err := json.Marshal(IssueRequest{
// 	Title: title,
// 	Body:  body,
// })
// if err != nil {
// 	fmt.Println("Error marshaling JSON:", err)
// }

// url := fmt.Sprintf("https://api.github.com/repos/%s/%s/issues", "vishwa2-uber", "cadence-java-samples")
// req, err := http.NewRequest("POST", url, bytes.NewBuffer(issueData))
// if err != nil {
// 	fmt.Println("Error creating request:", err)
// }

// req.Header.Set("Content-Type", "application/json")

// client := &http.Client{}
// resp, err := client.Do(req)
// if err != nil {
// 	fmt.Println("Error sending request:", err)
// }
// defer resp.Body.Close()

// if resp.StatusCode != http.StatusCreated {
// 	fmt.Println("Failed to create issue. Status code:", resp.StatusCode)
// 	buf := new(bytes.Buffer)
// 	buf.ReadFrom(resp.Body)
// 	newStr := buf.String()
// 	fmt.Println(newStr)
// 	os.Exit(1)
// 	// return
// }

// fmt.Println("Issue created successfully!")

// 			createArgs := []string{
// 				"gh",
// 				"issue",
// 				"create",
// 				"--repo",
// 				"cadence-workflow/cadence-java-samples",
// 				"--title",
// 				fmt.Sprintf("[JIRA %s] %s", key, title),
// 				"--body",
// 				body,
// 			}

// 			_, err := exec.Command(createArgs[0], createArgs[1:]...).Output()
// 			if err != nil {
// 				return err
// 			}
// 			activity.GetLogger(ctx).Info("Created an issue in GitHub", zap.String("title", title))
// 		} else {
// 			activity.GetLogger(ctx).Info("GitHub issue already exists", zap.String("title", title))
// 		}
// 	}
// 	return nil
// }

func getEnvVars() (map[string]string, error) {
	jiraArgs := []string{"usso", "-ussh", "t3", "-print"}
	output, err := exec.Command(jiraArgs[0], jiraArgs[1:]...).Output()
	if err != nil {
		return nil, err
	}

	githubArgs := []string{"usso", "-ussh", "git", "-print"}
	githubTokenOutput, err := exec.Command(githubArgs[0], githubArgs[1:]...).Output()
	if err != nil {
		return nil, err
	}

	envVars := map[string]string{
		"JIRA_API_TOKEN": string(output),
		"GITHUB_TOKEN":   strings.TrimSuffix(string(githubTokenOutput), "\n"),
	}

	return envVars, nil
}
