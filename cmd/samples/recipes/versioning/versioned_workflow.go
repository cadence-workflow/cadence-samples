package main

import (
	"context"
	"github.com/uber-common/cadence-samples/cmd/samples/common"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"time"
)

/**
 * This sample workflow continuously counting signals and do continue as new
 */

const (
	// ApplicationName is the task list for this sample
	ApplicationName = "versioning"

	// TestChangeID is a constant used to identify the version change in the workflow.
	TestChangeID = "test-change"

	// FooActivityName and BarActivityName are the names of the activities used in the workflows.
	FooActivityName = "FooActivity"
	BarActivityName = "BarActivity"

	// VersionedWorkflowName is the name of the versioned workflow.
	VersionedWorkflowName = "VersionedWorkflow"

	// VersionedWorkflowID is the ID of the versioned workflow.
	VersionedWorkflowID = "versioned_workflow"

	// StopSignalName is the name of the signal used to stop the workflow to finish it successfully
	StopSignalName = "StopSignal"
)

const (
	V1 int32 = iota + 1
	V2
	V3
	V4
)

var activityOptions = workflow.ActivityOptions{
	ScheduleToStartTimeout: time.Minute,
	StartToCloseTimeout:    time.Minute,
	HeartbeatTimeout:       time.Second * 20,
}

// VersionedWorkflowV1 is the first version of the workflow, supports only DefaultVersion.
// All workflows started by this version will have the change ID set to DefaultVersion.
func VersionedWorkflowV1(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	err := workflow.ExecuteActivity(ctx, FooActivityName).Get(ctx, nil)
	if err != nil {
		return err
	}

	return waitForSignal(ctx, V1)
}

// VersionedWorkflowV2 is the second version of the workflow, supports DefaultVersion and 1
// All workflows started by this version will have the change ID set to DefaultVersion.
func VersionedWorkflowV2(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var err error
	var version workflow.Version

	version = workflow.GetVersion(ctx, TestChangeID, workflow.DefaultVersion, 1, workflow.ExecuteWithMinVersion())
	if version == workflow.DefaultVersion {
		err = workflow.ExecuteActivity(ctx, FooActivityName).Get(ctx, nil)
	} else {
		err = workflow.ExecuteActivity(ctx, BarActivityName).Get(ctx, nil)
	}
	if err != nil {
		return err
	}

	return waitForSignal(ctx, V2)
}

// VersionedWorkflowV3 is the third version of the workflow, supports DefaultVersion and 1
// All workflows started by this version will have the change ID set to 1.
func VersionedWorkflowV3(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var err error
	var version workflow.Version

	version = workflow.GetVersion(ctx, TestChangeID, workflow.DefaultVersion, 1)
	if version == workflow.DefaultVersion {
		err = workflow.ExecuteActivity(ctx, FooActivityName).Get(ctx, nil)
	} else {
		err = workflow.ExecuteActivity(ctx, BarActivityName).Get(ctx, nil)
	}
	if err != nil {
		return err
	}

	return waitForSignal(ctx, V3)
}

// VersionedWorkflowV4 is the fourth version of the workflow, supports only version 1
// All workflows started by this version will have the change ID set to 1.
func VersionedWorkflowV4(ctx workflow.Context) error {
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	workflow.GetVersion(ctx, TestChangeID, 1, 1)
	err := workflow.ExecuteActivity(ctx, BarActivityName).Get(ctx, nil)
	if err != nil {
		return err
	}

	return waitForSignal(ctx, V4)
}

func waitForSignal(ctx workflow.Context, version int32) error {
	workflow.GetLogger(ctx).Info("Waiting for signal", zap.Int32("Worker Version", version))

	signalCh := workflow.GetSignalChannel(ctx, StopSignalName)

	for {
		var signal string
		if signalCh.ReceiveAsync(&signal) {
			break
		}

		workflow.GetLogger(ctx).Info("No signal received yet, continuing to wait...", zap.Int32("Worker Version", version))
		workflow.Sleep(ctx, time.Second*5)
	}

	workflow.GetLogger(ctx).Info("Got the signal, finishing the workflow", zap.Int32("Worker Version", version))
	return nil
}

// SetupHelperForVersionedWorkflowV1 registers VersionedWorkflowV1 and FooActivity
func SetupHelperForVersionedWorkflowV1(h *common.SampleHelper) {
	h.RegisterWorkflowWithAlias(VersionedWorkflowV1, VersionedWorkflowName)
	h.RegisterActivityWithAlias(FooActivity, FooActivityName)
}

// SetupHelperForVersionedWorkflowV2 registers VersionedWorkflowV2, FooActivity, and BarActivity
func SetupHelperForVersionedWorkflowV2(h *common.SampleHelper) {
	h.RegisterWorkflowWithAlias(VersionedWorkflowV2, VersionedWorkflowName)
	h.RegisterActivityWithAlias(FooActivity, FooActivityName)
	h.RegisterActivityWithAlias(BarActivity, BarActivityName)
}

// SetupHelperForVersionedWorkflowV3 registers VersionedWorkflowV3, FooActivity, and BarActivity
func SetupHelperForVersionedWorkflowV3(h *common.SampleHelper) {
	h.RegisterWorkflowWithAlias(VersionedWorkflowV3, VersionedWorkflowName)
	h.RegisterActivityWithAlias(FooActivity, FooActivityName)
	h.RegisterActivityWithAlias(BarActivity, BarActivityName)
}

// SetupHelperForVersionedWorkflowV4 registers VersionedWorkflowV4 and BarActivity
func SetupHelperForVersionedWorkflowV4(h *common.SampleHelper) {
	h.RegisterWorkflowWithAlias(VersionedWorkflowV4, VersionedWorkflowName)
	h.RegisterActivityWithAlias(BarActivity, BarActivityName)
}

// FooActivity returns "foo" as a result of the activity execution.
func FooActivity(ctx context.Context) (string, error) {
	activity.GetLogger(ctx).Info("Executing FooActivity")
	return "foo", nil
}

// BarActivity returns "bar" as a result of the activity execution.
func BarActivity(ctx context.Context) (string, error) {
	activity.GetLogger(ctx).Info("Executing BarActivity")
	return "bar", nil
}
