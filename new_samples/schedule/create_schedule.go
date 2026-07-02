package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/client"
	"go.uber.org/zap"
)

func runCreate() {
	logger := BuildLogger()
	c := buildScheduleClient()

	action := startWorkflowAction(logger)
	action.WorkflowIDPrefix = ScheduleID + "-"

	_, err := c.Create(context.Background(), &client.CreateScheduleRequest{
		ScheduleID: ScheduleID,
		Spec:       &client.ScheduleSpec{CronExpression: "* * * * *"},
		Action:     &client.ScheduleAction{StartWorkflow: action},
		Policies: &client.SchedulePolicies{
			OverlapPolicy:  client.ScheduleOverlapPolicySkipNew,
			CatchUpPolicy:  client.ScheduleCatchUpPolicySkip,
			PauseOnFailure: true,
		},
		Memo: map[string]interface{}{
			"owner": "platform-team",
			"env":   "dev",
		},
	})
	if err != nil {
		if be, ok := err.(*shared.BadRequestError); ok && strings.Contains(be.Message, "already") {
			fmt.Printf("Schedule %q already exists. Run delete first or use a different schedule ID.\n", ScheduleID)
			return
		}
		logger.Fatal("Create failed", zap.Error(err))
	}
	fmt.Printf("Created schedule %q (fires every minute)\n", ScheduleID)

	desc, err := c.Describe(context.Background(), ScheduleID)
	if err != nil {
		logger.Fatal("Describe failed", zap.Error(err))
	}
	printSchedule(ScheduleID, desc)
}

func printSchedule(id string, desc *client.DescribeScheduleResponse) {
	fmt.Printf("\nSchedule : %q\n", id)
	if desc.Spec != nil {
		fmt.Printf("  cron           : %q\n", desc.Spec.CronExpression)
	}
	if desc.State != nil {
		fmt.Printf("  paused         : %v\n", desc.State.Paused)
		if desc.State.Paused && desc.State.PauseInfo != nil {
			fmt.Printf("  pause reason   : %q\n", desc.State.PauseInfo.Reason)
		}
	}
	if desc.Policies != nil {
		fmt.Printf("  overlap policy : %v\n", desc.Policies.OverlapPolicy)
		fmt.Printf("  pause_on_fail  : %v\n", desc.Policies.PauseOnFailure)
	}
	if desc.Info != nil {
		if !desc.Info.NextRunTime.IsZero() {
			fmt.Printf("  next run       : %s UTC\n", desc.Info.NextRunTime.UTC().Format(time.RFC3339))
		}
		if !desc.Info.LastRunTime.IsZero() {
			fmt.Printf("  last run       : %s UTC\n", desc.Info.LastRunTime.UTC().Format(time.RFC3339))
		}
		fmt.Printf("  total runs     : %d\n", desc.Info.TotalRuns)
	}
}
