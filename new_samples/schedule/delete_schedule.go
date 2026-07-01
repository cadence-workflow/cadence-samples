package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func runDelete() {
	logger := BuildLogger()
	c := buildScheduleClient()

	if err := c.Delete(context.Background(), ScheduleID); err != nil {
		logger.Fatal("Delete failed", zap.Error(err))
	}
	fmt.Printf("Deleted schedule %q. Deletion is async — wait a few seconds before creating again.\n", ScheduleID)
}
