package main

import (
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// ConsistentQueryWorkflow demonstrates query handlers with signal handling.
func ConsistentQueryWorkflow(ctx workflow.Context) error {
	queryResult := 0
	logger := workflow.GetLogger(ctx)
	logger.Info("ConsistentQueryWorkflow started")

	// Setup query handler for "state" query type
	err := workflow.SetQueryHandler(ctx, "state", func(input []byte) (int, error) {
		return queryResult, nil
	})
	if err != nil {
		logger.Info("SetQueryHandler failed: " + err.Error())
		return err
	}

	signalChan := workflow.GetSignalChannel(ctx, "increase")

	s := workflow.NewSelector(ctx)
	s.AddReceive(signalChan, func(c workflow.Channel, more bool) {
		c.Receive(ctx, nil)
		queryResult += 1
		workflow.GetLogger(ctx).Info("Received signal!", zap.String("signal", "increase"))
	})

	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			s.Select(ctx)
		}
	})

	// Wait for timer before completing
	workflow.NewTimer(ctx, time.Minute*2).Get(ctx, nil)
	logger.Info("Timer fired")

	logger.Info("ConsistentQueryWorkflow completed")
	return nil
}

