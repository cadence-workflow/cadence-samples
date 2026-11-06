package main

import (
	"context"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	RunActivitySignal = "run_activity"
	CompleteSignal    = "complete"

	ActiveActiveWorkflowName = "cadence_samples.ActiveActiveWorkflow"
	GetClusterActivityName   = "cadence_samples.GetClusterActivity"
)

func ActiveActiveWorkflow(ctx workflow.Context) ([]string, error) {
	var clusters []string
	runActivityChan := workflow.GetSignalChannel(ctx, RunActivitySignal)
	completeChan := workflow.GetSignalChannel(ctx, CompleteSignal)
	for {
		cluster, err := runActivity(ctx)
		if err != nil {
			return nil, err
		}
		clusters = append(clusters, cluster)

		var complete bool
		s := workflow.NewSelector(ctx)
		s.AddReceive(runActivityChan, func(ch workflow.Channel, _ bool) {
			ch.Receive(ctx, nil)
		})
		s.AddReceive(completeChan, func(ch workflow.Channel, _ bool) {
			ch.Receive(ctx, nil)
			complete = true
		})
		s.Select(ctx)

		if complete {
			return clusters, nil
		}
	}
}

func runActivity(ctx workflow.Context) (string, error) {
	ctx = workflow.WithActivityOptions(ctx, workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Hour,
		StartToCloseTimeout:    time.Minute,
	})
	var cluster string
	err := workflow.ExecuteActivity(ctx, GetClusterActivityName).Get(ctx, &cluster)
	return cluster, err
}

type ClusterActivities struct {
	Cluster string
}

func (a *ClusterActivities) GetCluster(ctx context.Context) (string, error) {
	activity.GetLogger(ctx).Info("GetClusterActivity started", zap.String("cluster", a.Cluster))
	return a.Cluster, nil
}
