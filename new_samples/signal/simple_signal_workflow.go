package main

import (
	"context"
	"go.uber.org/cadence/workflow"
	"go.uber.org/cadence/activity"
	"strconv"
	"time"
	"go.uber.org/zap"
)

const (
	CompleteSignalChan = "complete"
)

func SimpleSignalWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute * 60,
		StartToCloseTimeout:    time.Minute * 60,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("SimpleSignalWorkflow started")

	var complete bool
	completeChan := workflow.GetSignalChannel(ctx, CompleteSignalChan)
	for {
		s := workflow.NewSelector(ctx)
		s.AddReceive(completeChan, func(ch workflow.Channel, ok bool) {
			if ok {
				ch.Receive(ctx, &complete)
			}
			logger.Info("Signal input: " + strconv.FormatBool(complete))
		})
		s.Select(ctx)

		var result string
		err := workflow.ExecuteActivity(ctx, SimpleSignalActivity, complete).Get(ctx, &result)
		if err != nil {
			return err
		}
		logger.Info("Activity result: " + result)
		if complete {
			return nil
		}
	}
}

func SimpleSignalActivity(ctx context.Context, complete bool) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("SimpleSignalActivity started, a new signal has been received", zap.Bool("complete", complete))
	if complete {
		return "Workflow will complete now", nil
	}
	return "Workflow will continue to run", nil
}
