package main

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/zap"
)

func runBackfill() {
	logger := BuildLogger()
	c := buildScheduleClient()
	sc := c.ScheduleClient()

	end := time.Now()
	start := end.Add(-2 * time.Hour)

	if err := sc.Backfill(context.Background(), ScheduleID, &client.BackfillRequest{
		StartTime:     start,
		EndTime:       end,
		OverlapPolicy: client.ScheduleOverlapPolicyBuffer,
	}); err != nil {
		logger.Fatal("Backfill failed", zap.Error(err))
	}
	fmt.Printf("Backfilled schedule %q (%s -> %s UTC, 2h window)\n",
		ScheduleID,
		start.UTC().Format("15:04"),
		end.UTC().Format("15:04"))
}
