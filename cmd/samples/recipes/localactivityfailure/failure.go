package main

import (
	"context"
	"fmt"
	"go.uber.org/cadence"
	"go.uber.org/zap"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/cadence/workflow"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var globalMu sync.Mutex

// a data-racy, completely problematic means to capture the completion of an activity
// locally and show where sometimes it's being retried
var sideEffectsCapture int

type Result struct {
	Result *int
}

func testLocalActivity(ctx workflow.Context) error {
	log := workflow.GetLogger(ctx)
	// zero out the global variable to start with to ensure test is sane
	// this workflow is only a demonstration and global variables like this are *A BAD IDEA*
	// it's being expicitly used to demonstrate how side-effects are occurring
	sideEffectsCapture = 0
	ao := workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: time.Second,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    time.Nanosecond,
			BackoffCoefficient: 1,
			MaximumAttempts:    100,
		},
	}
	ctx = workflow.WithLocalActivityOptions(ctx, ao)

	mu := sync.Mutex{}
	sum := 0
	wg := workflow.NewWaitGroup(ctx)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		workflow.Go(ctx, func(ctx workflow.Context) {
			res := Result{}
			// try doing an activity which crashes
			err := workflow.ExecuteLocalActivity(ctx, SumActivity).Get(ctx, &res)
			if err == nil {
				if res.Result != nil {
					mu.Lock()
					sum += *res.Result
					mu.Unlock()
				}
			} else {
				log.Error("err - something went wrong ", zap.Error(err))
			}
			wg.Done()
		})
	}
	wg.Wait(ctx)
	log.Info(">>> Result - should be 100: ", zap.Int("sum", sum), zap.Int("side-effects", sideEffectsCapture))
	return nil
}

func testNormalActivity(ctx workflow.Context) error {
	log := workflow.GetLogger(ctx)
	sideEffectsCapture = 0
	ao := workflow.ActivityOptions{
		StartToCloseTimeout:    time.Minute,
		ScheduleToCloseTimeout: time.Second,
		ScheduleToStartTimeout: time.Second,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    time.Nanosecond,
			BackoffCoefficient: 1,
			MaximumAttempts:    100,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	mu := sync.Mutex{}
	sum := 0
	wg := workflow.NewWaitGroup(ctx)

	for i := 0; i < 20; i++ {
		wg.Add(1)
		workflow.Go(ctx, func(ctx workflow.Context) {
			res := Result{}
			// try doing an activity which crashes
			err := workflow.ExecuteActivity(ctx, SumActivity).Get(ctx, &res)
			if err == nil {
				if res.Result != nil {
					mu.Lock()
					sum += *res.Result
					mu.Unlock()
				}
			} else {
				log.Error("err - something went wrong ", zap.Error(err))
			}
			wg.Done()
		})
	}
	wg.Wait(ctx)
	log.Info(">>> Result - should be 100: ", zap.Int("sum", sum), zap.Int("side-effects", sideEffectsCapture))
	return nil
}

// an activity which crashes a lot and times out a lot
// and is expected to give the right result ultimately through
// the use of an aggressive retry policy
func SumActivity(ctx context.Context) (*Result, error) {

	if rand.Intn(10) > 5 {
		fmt.Println(">> returning an error")
		return nil, fmt.Errorf("an error")
	}

	if rand.Intn(10) > 5 {
		// timeout
		time.Sleep(time.Second)
	}

	//if rand.Intn(10) > 5 {
	//  // at the time of writing panics aren't handled by localactivities
	//  panic("an error")
	//}

	output := 5
	globalMu.Lock()
	sideEffectsCapture += 5
	globalMu.Unlock()
	return &Result{Result: &output}, nil
}
