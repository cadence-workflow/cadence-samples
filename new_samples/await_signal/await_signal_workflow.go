package main

import (
	"context"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
)

var signalToSignalTimeout = time.Second * 30
var fromFirstSignalTimeout = time.Second * 60

type AwaitSignals struct {
	FirstSignalTime time.Time
	Signal1Received bool
	Signal2Received bool
	Signal3Received bool
}

// Listen to signals Signal1, Signal2, and Signal3
func (a *AwaitSignals) Listen(ctx workflow.Context) {
	logger := workflow.GetLogger(ctx)
	for {
		selector := workflow.NewSelector(ctx)
		selector.AddReceive(workflow.GetSignalChannel(ctx, "Signal1"), func(c workflow.Channel, more bool) {
			c.Receive(ctx, nil)
			a.Signal1Received = true
			logger.Info("Signal1 Received")
		})
		selector.AddReceive(workflow.GetSignalChannel(ctx, "Signal2"), func(c workflow.Channel, more bool) {
			c.Receive(ctx, nil)
			a.Signal2Received = true
			logger.Info("Signal2 Received")
		})
		selector.AddReceive(workflow.GetSignalChannel(ctx, "Signal3"), func(c workflow.Channel, more bool) {
			c.Receive(ctx, nil)
			a.Signal3Received = true
			logger.Info("Signal3 Received")
		})
		selector.Select(ctx)
		if a.FirstSignalTime.IsZero() {
			a.FirstSignalTime = workflow.Now(ctx)
		}
	}
}

// GetNextTimeout returns the maximum time allowed to wait for the next signal.
func (a *AwaitSignals) GetNextTimeout(ctx workflow.Context) (time.Duration, error) {
	if a.FirstSignalTime.IsZero() {
		panic("FirstSignalTime is not yet set")
	}
	total := workflow.Now(ctx).Sub(a.FirstSignalTime)
	totalLeft := fromFirstSignalTimeout - total
	if totalLeft <= 0 {
		return 0, cadence.NewCustomError("FromFirstSignalTimeout")
	}
	if signalToSignalTimeout < totalLeft {
		return signalToSignalTimeout, nil
	}
	return totalLeft, nil
}

// cadence.AwaitWithTimeout is not exported, so we need to implement it ourselves.
func awaitWithTimeout(ctx workflow.Context, timeout time.Duration, condition func() bool) (bool, error) {
	ctx, cancel := workflow.WithCancel(ctx)
	defer cancel()
	timer := workflow.NewTimer(ctx, timeout)
	err := workflow.Await(ctx, func() bool {
		return timer.IsReady() || condition()
	})
	if err != nil {
		return false, err
	}
	return condition(), nil
}

func AwaitSignalWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("AwaitSignalWorkflow started")

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute * 60,
		StartToCloseTimeout:    time.Minute * 60,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	awaitSignals := &AwaitSignals{}
	workflow.Go(ctx, awaitSignals.Listen)

	err := workflow.Await(ctx, func() bool {
		return awaitSignals.Signal1Received
	})

	if err != nil {
		return err
	}
	err = workflow.ExecuteActivity(ctx, Signal1Activity).Get(ctx, nil)
	if err != nil {
		return err
	}

	// Wait for Signal2
	nextTimeout, err := awaitSignals.GetNextTimeout(ctx)

	if err != nil {
		return err
	}

	ok, err := awaitWithTimeout(ctx, nextTimeout, func() bool {
		return awaitSignals.Signal2Received
	})
	if err != nil {
		return err
	}
	if !ok {
		return cadence.NewCustomError("Signal2 not received")
	}
	err = workflow.ExecuteActivity(ctx, Signal2Activity).Get(ctx, nil)
	if err != nil {
		return err
	}

	// Wait for Signal3
	nextTimeout, err = awaitSignals.GetNextTimeout(ctx)
	if err != nil {
		return err
	}
	ok, err = awaitWithTimeout(ctx, nextTimeout, func() bool {
		return awaitSignals.Signal3Received
	})
	if err != nil {
		return err
	}
	if !ok {
		return cadence.NewCustomError("Signal3 not received")
	}
	err = workflow.ExecuteActivity(ctx, Signal3Activity).Get(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}

func Signal1Activity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Signal1Activity started")
	return nil
}

func Signal2Activity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Signal2Activity started")
	return nil
}

func Signal3Activity(ctx context.Context) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Signal3Activity started")
	return nil
}
