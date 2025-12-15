package main

import (
	"context"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// VersionedWorkflow demonstrates workflow versioning for safe code changes.
// Use GetVersion to branch between old and new code paths during deployment.
func VersionedWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("VersionedWorkflow started")

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	// GetVersion allows branching between old and new code paths
	// - First param: unique change ID
	// - Second param: minimum supported version (DefaultVersion = -1)
	// - Third param: maximum supported version
	version := workflow.GetVersion(ctx, "activity-change", workflow.DefaultVersion, 1)

	var err error
	if version == workflow.DefaultVersion {
		// Old code path - for workflows started before the change
		err = workflow.ExecuteActivity(ctx, OldActivity).Get(ctx, nil)
	} else {
		// New code path - for workflows started after the change
		err = workflow.ExecuteActivity(ctx, NewActivity).Get(ctx, nil)
	}

	if err != nil {
		return err
	}

	logger.Info("VersionedWorkflow completed", zap.Int("version", int(version)))
	return nil
}

// OldActivity represents the original activity before the change.
func OldActivity(ctx context.Context) (string, error) {
	activity.GetLogger(ctx).Info("Executing OldActivity (version DefaultVersion)")
	return "old result", nil
}

// NewActivity represents the new activity after the change.
func NewActivity(ctx context.Context) (string, error) {
	activity.GetLogger(ctx).Info("Executing NewActivity (version 1)")
	return "new result", nil
}

