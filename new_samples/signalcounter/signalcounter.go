package main

import (
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

/**
 * This sample workflow continuously counting signals and do continue as new
 */

// ApplicationName is the task list for this sample
const ApplicationName = "signal_counter"

// A workflow execution cannot receive infinite number of signals due to history limit
// By default 10000 is MaximumSignalsPerExecution which can be configured by DynamicConfig of Cadence cluster.
// But it's recommended to do continueAsNew after receiving certain number of signals(in production, use a number <1000)
const maxSignalsPerExecution = 3

// SignalCounterWorkflow Workflow Decider.
func SignalCounterWorkflow(ctx workflow.Context, counter int) error {

	var drainedAllSignals bool
	signalsPerExecution := 0

	logger := workflow.GetLogger(ctx)
	logger.Info("Started SignalCounterWorkflow")

	for {
		s := workflow.NewSelector(ctx)
		s.AddReceive(workflow.GetSignalChannel(ctx, "channelA"), func(c workflow.Channel, ok bool) {
			if ok {
				var i int
				c.Receive(ctx, &i)
				counter += i
				signalsPerExecution += 1
				logger.Info("Received signal on channelA.", zap.Int("Counter", i))
			}
		})
		s.AddReceive(workflow.GetSignalChannel(ctx, "channelB"), func(c workflow.Channel, ok bool) {
			if ok {
				var i int
				c.Receive(ctx, &i)
				counter += i
				signalsPerExecution += 1
				logger.Info("Received signal on channelB.", zap.Int("Counter", i))
			}
		})

		if signalsPerExecution >= maxSignalsPerExecution {
			s.AddDefault(func() {
				// this indicate that we have drained all signals within the decision task, and it's safe to do a continueAsNew
				drainedAllSignals = true
				logger.Info("Reached maxSignalsPerExecution limit")
			})
		}

		s.Select(ctx)

		if drainedAllSignals {
			logger.Info("Returning ContinueAsNewError")
			return workflow.NewContinueAsNewError(ctx, SignalCounterWorkflow, counter)
		}
	}
}
