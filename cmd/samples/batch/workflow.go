package batch

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/multierr"
)

type BatchWorkflowInput struct {
	Concurrency int
	TotalSize   int
}

func BatchWorkflow(ctx workflow.Context, input BatchWorkflowInput) error {
	wg := workflow.NewWaitGroup(ctx)

	buffered := workflow.NewBufferedChannel(ctx, input.Concurrency)
	futures := workflow.NewNamedChannel(ctx, "futures")

	var errs error
	wg.Add(1)
	// task result collector
	workflow.Go(ctx, func(ctx workflow.Context) {
		defer wg.Done()
		for {
			var future workflow.Future
			ok := futures.Receive(ctx, &future)
			if !ok {
				break
			}
			err := future.Get(ctx, nil)
			errs = multierr.Append(errs, err)
			buffered.Receive(ctx, nil)
		}
	})

	// submit all tasks
	for taskID := 0; taskID < input.TotalSize; taskID++ {
		taskID := taskID
		buffered.Send(ctx, nil)

		aCtx := workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
			ScheduleToStartTimeout: time.Second * 10,
			StartToCloseTimeout:    time.Second * 10,
		})
		futures.Send(ctx, workflow.ExecuteActivity(aCtx, BatchActivity, taskID))
	}
	// close the channel to signal the task result collector that no more tasks are coming
	futures.Close()

	wg.Wait(ctx)

	return errs
}

func BatchActivity(ctx context.Context, taskID int) error {
	select {
	case <-ctx.Done():
		return fmt.Errorf("batch activity %d failed: %w", taskID, ctx.Err())
	case <-time.After(time.Duration(rand.Int63n(100))*time.Millisecond + 900*time.Millisecond):
		return nil
	}
}
