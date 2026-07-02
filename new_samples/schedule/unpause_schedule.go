package main

import (
	"context"
	"fmt"

	"go.uber.org/cadence/client"
	"go.uber.org/zap"
)

func runUnpause() {
	logger := BuildLogger()
	c := buildScheduleClient()

	const reason = "resuming via sample"
	if err := c.Unpause(context.Background(), ScheduleID, reason, client.ScheduleCatchUpPolicySkip); err != nil {
		logger.Fatal("Unpause failed", zap.Error(err))
	}
	fmt.Printf("Unpaused schedule %q (reason: %q)\n", ScheduleID, reason)
}
