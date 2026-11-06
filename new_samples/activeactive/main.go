package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"slices"
	"time"

	"go.uber.org/cadence/client"
)

// Regions are the cluster attributes, in scope "region", registered on the domain.
var Regions = []string{"us-west", "us-central", "us-east"}

const ClusterAttributeScope = "region"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "worker":
		runWorker(os.Args[2:])
	case "start":
		startWorkflow(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: go run . <command> [flags]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  worker --cluster <cluster>          Run a worker that polls cluster0 or cluster1")
	fmt.Println("  start --region <region> [--wid ID]  Start an ActiveActiveWorkflow in a region:", Regions)
}

func runWorker(args []string) {
	fs := flag.NewFlagSet("worker", flag.ExitOnError)
	cluster := fs.String("cluster", "", "cluster to poll: cluster0 or cluster1")
	fs.Parse(args)

	if _, ok := Clusters[*cluster]; !ok {
		fmt.Printf("Unknown cluster %q\n", *cluster)
		os.Exit(1)
	}
	if err := RunWorker(*cluster); err != nil {
		fmt.Printf("Worker failed: %v\n", err)
		os.Exit(1)
	}
}

func startWorkflow(args []string) {
	fs := flag.NewFlagSet("start", flag.ExitOnError)
	region := fs.String("region", "", fmt.Sprint("region of the workflow: ", Regions))
	workflowID := fs.String("wid", "", "workflow ID (default: a random ID)")
	fs.Parse(args)

	if !slices.Contains(Regions, *region) {
		fmt.Printf("Unknown region %q, expected one of %v\n", *region, Regions)
		os.Exit(1)
	}

	// Either cluster can take the request; the workflow runs where its region is active.
	cadenceClient := client.NewClient(BuildCadenceClient(Clusters["cluster0"]), Domain, &client.Options{})
	options := client.StartWorkflowOptions{
		ID:                           *workflowID,
		TaskList:                     TaskListName,
		ExecutionStartToCloseTimeout: time.Hour,
		ActiveClusterSelectionPolicy: &client.ActiveClusterSelectionPolicy{
			ClusterAttribute: &client.ClusterAttribute{
				Scope: ClusterAttributeScope,
				Name:  *region,
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	execution, err := cadenceClient.StartWorkflow(ctx, options, ActiveActiveWorkflowName)
	if err != nil {
		fmt.Printf("Failed to start workflow: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Started workflow %s (region %s)\n", execution.ID, *region)
}
