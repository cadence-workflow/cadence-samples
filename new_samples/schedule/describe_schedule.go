package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func runDescribe() {
	logger := BuildLogger()
	c := buildScheduleClient()
	sc := c.ScheduleClient()

	desc, err := sc.Describe(context.Background(), ScheduleID)
	if err != nil {
		logger.Fatal("Describe failed", zap.String("scheduleID", ScheduleID), zap.Error(err))
	}
	fmt.Printf("Described schedule %q\n", ScheduleID)
	printSchedule(ScheduleID, desc)
}
