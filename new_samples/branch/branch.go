package main

import (
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"
)

const totalBranches = 3

// BranchWorkflow executes multiple activities in parallel and waits for all.
func BranchWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("BranchWorkflow started")

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var futures []workflow.Future
	for i := 1; i <= totalBranches; i++ {
		input := fmt.Sprintf("branch %d of %d", i, totalBranches)
		future := workflow.ExecuteActivity(ctx, BranchActivity, input)
		futures = append(futures, future)
	}

	for _, f := range futures {
		if err := f.Get(ctx, nil); err != nil {
			return err
		}
	}

	logger.Info("BranchWorkflow completed")
	return nil
}

// BranchActivity processes a single branch.
func BranchActivity(input string) (string, error) {
	fmt.Printf("BranchActivity: %s\n", input)
	return "Result_" + input, nil
}

