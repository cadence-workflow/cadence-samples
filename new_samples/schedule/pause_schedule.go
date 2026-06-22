package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func runPause() {
	logger := BuildLogger()
	c := buildScheduleClient()
	sc := c.ScheduleClient()

	const reason = "paused via sample"
	if err := sc.Pause(context.Background(), ScheduleID, reason); err != nil {
		logger.Fatal("Pause failed", zap.Error(err))
	}
	fmt.Printf("Paused schedule %q (reason: %q)\n", ScheduleID, reason)
}
