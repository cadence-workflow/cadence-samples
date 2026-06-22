package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func runDelete() {
	logger := BuildLogger()
	c := buildScheduleClient()
	sc := c.ScheduleClient()

	if err := sc.Delete(context.Background(), ScheduleID); err != nil {
		logger.Fatal("Delete failed", zap.Error(err))
	}
	fmt.Printf("Deleted schedule %q. Deletion is async — wait a few seconds before creating again.\n", ScheduleID)
}
