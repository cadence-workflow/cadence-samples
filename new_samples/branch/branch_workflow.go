package main

import (
	"errors"
	"fmt"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const totalBranches = 3

// BranchWorkflow demonstrates executing multiple activities in parallel using Futures.
// All branches run concurrently and we wait for all to complete.
func BranchWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("BranchWorkflow started")

	// Start all activities in parallel
	var futures []workflow.Future
	for i := 1; i <= totalBranches; i++ {
		activityInput := fmt.Sprintf("branch %d of %d", i, totalBranches)
		future := workflow.ExecuteActivity(ctx, BranchActivity, activityInput)
		futures = append(futures, future)
	}

	// Wait for all futures to complete
	for i, future := range futures {
		var result string
		if err := future.Get(ctx, &result); err != nil {
			logger.Error("Branch failed", zap.Int("branch", i+1), zap.Error(err))
			return err
		}
		logger.Info("Branch completed", zap.Int("branch", i+1), zap.String("result", result))
	}

	logger.Info("BranchWorkflow completed - all branches finished")
	return nil
}

// ParallelWorkflow demonstrates using workflow.Go() to run coroutines in parallel.
// Each coroutine can run multiple sequential activities.
func ParallelWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("ParallelWorkflow started")

	waitChannel := workflow.NewChannel(ctx)

	// First coroutine: runs two activities sequentially
	workflow.Go(ctx, func(ctx workflow.Context) {
		err := workflow.ExecuteActivity(ctx, BranchActivity, "branch1.1").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
			waitChannel.Send(ctx, err.Error())
			return
		}
		err = workflow.ExecuteActivity(ctx, BranchActivity, "branch1.2").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
			waitChannel.Send(ctx, err.Error())
			return
		}
		waitChannel.Send(ctx, "")
	})

	// Second coroutine: runs one activity
	workflow.Go(ctx, func(ctx workflow.Context) {
		err := workflow.ExecuteActivity(ctx, BranchActivity, "branch2").Get(ctx, nil)
		if err != nil {
			logger.Error("Activity failed", zap.Error(err))
			waitChannel.Send(ctx, err.Error())
			return
		}
		waitChannel.Send(ctx, "")
	})

	// Wait for both coroutines to complete
	var errMsg string
	for i := 0; i < 2; i++ {
		waitChannel.Receive(ctx, &errMsg)
		if errMsg != "" {
			err := errors.New(errMsg)
			logger.Error("Coroutine failed", zap.Error(err))
			return err
		}
	}

	logger.Info("ParallelWorkflow completed")
	return nil
}

// BranchActivity is a simple activity that logs and returns a result.
func BranchActivity(input string) (string, error) {
	fmt.Printf("BranchActivity running with input: %s\n", input)
	return "Result_" + input, nil
}

