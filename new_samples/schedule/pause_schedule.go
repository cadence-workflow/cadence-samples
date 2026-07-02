package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func runPause() {
	logger := BuildLogger()
	c := buildScheduleClient()

	const reason = "paused via sample"
	if err := c.Pause(context.Background(), ScheduleID, reason); err != nil {
		logger.Fatal("Pause failed", zap.Error(err))
	}
	fmt.Printf("Paused schedule %q (reason: %q)\n", ScheduleID, reason)
}
