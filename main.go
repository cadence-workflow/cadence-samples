package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

func main() {
	// Replace with your actual token
	token := "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"

	// GitHub org and repo
	// org := "cadence-workflow"
	// repo := "cadence-java-samples"
	org := "vishwa-test-2"
	repo := "sample"

	// Issue details
	title := "Sample Issue Title"
	body := "This is a test issue created via the GitHub API."

	// Authenticate using the token
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(ctx, ts)
	client := github.NewClient(tc)

	// Create issue request
	issueRequest := &github.IssueRequest{
		Title: github.String(title),
		Body:  github.String(body),
	}

	// Create issue
	issue, _, err := client.Issues.Create(ctx, org, repo, issueRequest)
	if err != nil {
		log.Fatalf("Error creating issue: %v", err)
	}

	fmt.Printf("Issue created: %s\n", issue.GetHTMLURL())
}
