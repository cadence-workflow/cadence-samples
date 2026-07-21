package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	// AcquireLockSignalName is the signal channel used to hand back the release-lock channel name.
	AcquireLockSignalName = "acquire-lock-event"
	// RequestLockSignalName is the signal channel used by callers to request the lock.
	RequestLockSignalName = "request-lock-event"
)

// UnlockFunc releases the distributed lock when called.
type UnlockFunc func() error

// Mutex is a distributed mutex backed by a Cadence workflow.
type Mutex struct {
	currentWorkflowID string
	lockNamespace     string
}

// NewMutex initializes a Mutex for the calling workflow.
func NewMutex(currentWorkflowID string, lockNamespace string) *Mutex {
	return &Mutex{
		currentWorkflowID: currentWorkflowID,
		lockNamespace:     lockNamespace,
	}
}

// Lock acquires the distributed lock for resourceID.
// It blocks until the lock is granted, then returns an UnlockFunc to release it.
func (s *Mutex) Lock(ctx workflow.Context, resourceID string, unlockTimeout time.Duration) (UnlockFunc, error) {
	activityCtx := workflow.WithLocalActivityOptions(ctx, workflow.LocalActivityOptions{
		ScheduleToCloseTimeout: time.Minute,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			ExpirationInterval: time.Minute * 10,
			MaximumAttempts:    5,
		},
	})

	var releaseLockChannelName string
	var execution workflow.Execution
	err := workflow.ExecuteLocalActivity(activityCtx,
		signalWithStartMutexWorkflowActivity,
		s.lockNamespace, resourceID, s.currentWorkflowID, unlockTimeout,
	).Get(ctx, &execution)
	if err != nil {
		return nil, err
	}

	// Wait for the mutex workflow to grant the lock and send us the release channel name.
	workflow.GetSignalChannel(ctx, AcquireLockSignalName).Receive(ctx, &releaseLockChannelName)

	unlockFunc := func() error {
		return workflow.SignalExternalWorkflow(ctx, execution.ID, execution.RunID,
			releaseLockChannelName, "releaseLock").Get(ctx, nil)
	}
	return unlockFunc, nil
}

// mutexWorkflow serializes access to a resource by processing lock requests one at a time.
func mutexWorkflow(
	ctx workflow.Context,
	namespace string,
	resourceID string,
	unlockTimeout time.Duration,
) error {
	currentWorkflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
	if currentWorkflowID == "default-test-workflow-id" {
		// Unit-testing hack: give the test environment time to register delayed callbacks.
		// See https://github.com/uber-go/cadence-client/issues/663
		workflow.Sleep(ctx, 10*time.Millisecond)
	}
	logger := workflow.GetLogger(ctx).With(zap.String("currentWorkflowID", currentWorkflowID))
	logger.Info("mutex workflow started")

	var ack string
	requestLockCh := workflow.GetSignalChannel(ctx, RequestLockSignalName)
	for {
		var senderWorkflowID string
		if !requestLockCh.ReceiveAsync(&senderWorkflowID) {
			logger.Info("no more pending lock requests")
			break
		}

		var releaseLockChannelName string
		_ = workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
			return generateUnlockChannelName(senderWorkflowID)
		}).Get(&releaseLockChannelName)

		logger := logger.With(zap.String("releaseLockChannelName", releaseLockChannelName))
		logger.Info("granting lock to workflow", zap.String("senderWorkflowID", senderWorkflowID))

		err := workflow.SignalExternalWorkflow(ctx, senderWorkflowID, "",
			AcquireLockSignalName, releaseLockChannelName).Get(ctx, nil)
		if err != nil {
			// If the requesting workflow is already gone, skip and release immediately.
			logger.Info("failed to signal requesting workflow, skipping", zap.Error(err))
			continue
		}

		// Wait for the lock holder to release, or for the timeout to expire.
		selector := workflow.NewSelector(ctx)
		selector.AddFuture(workflow.NewTimer(ctx, unlockTimeout), func(f workflow.Future) {
			logger.Info("unlock timeout exceeded, releasing lock")
		})
		selector.AddReceive(workflow.GetSignalChannel(ctx, releaseLockChannelName), func(c workflow.Channel, more bool) {
			c.Receive(ctx, &ack)
			logger.Info("lock released by holder")
		})
		selector.Select(ctx)
	}
	return nil
}

// signalWithStartMutexWorkflowActivity is a local activity that uses SignalWithStart to
// either start a new mutex workflow for the resource or signal an existing one to queue this request.
func signalWithStartMutexWorkflowActivity(
	ctx context.Context,
	namespace string,
	resourceID string,
	senderWorkflowID string,
	unlockTimeout time.Duration,
) (*workflow.Execution, error) {
	workflowID := fmt.Sprintf("mutex:%s:%s", namespace, resourceID)
	workflowOptions := client.StartWorkflowOptions{
		ID:                              workflowID,
		TaskList:                        TaskListName,
		ExecutionStartToCloseTimeout:    time.Hour,
		DecisionTaskStartToCloseTimeout: time.Hour,
		RetryPolicy: &cadence.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			ExpirationInterval: time.Minute * 10,
			MaximumAttempts:    5,
		},
		WorkflowIDReusePolicy: client.WorkflowIDReusePolicyAllowDuplicate,
	}

	cadenceClient := client.NewClient(BuildCadenceClient(), Domain, nil)
	return cadenceClient.SignalWithStartWorkflow(
		ctx, workflowID, RequestLockSignalName, senderWorkflowID,
		workflowOptions, mutexWorkflow, namespace, resourceID, unlockTimeout,
	)
}

// generateUnlockChannelName creates a unique signal channel name for releasing the lock.
func generateUnlockChannelName(senderWorkflowID string) string {
	return fmt.Sprintf("unlock-event-%s", senderWorkflowID)
}

// sampleWorkflowWithMutex demonstrates two workflows competing for the same resource lock.
func sampleWorkflowWithMutex(ctx workflow.Context, resourceID string) error {
	currentWorkflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
	logger := workflow.GetLogger(ctx).
		With(zap.String("currentWorkflowID", currentWorkflowID)).
		With(zap.String("resourceID", resourceID))
	logger.Info("workflow started, acquiring mutex")

	mutex := NewMutex(currentWorkflowID, "TestUseCase")
	unlockFunc, err := mutex.Lock(ctx, resourceID, 10*time.Minute)
	if err != nil {
		return err
	}
	defer unlockFunc() //nolint:errcheck
	logger.Info("mutex acquired, starting critical section")

	workflow.Sleep(ctx, 10*time.Second)

	logger.Info("critical section finished")
	return nil
}
