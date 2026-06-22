package main

import (
	"context"
	"fmt"

	"go.uber.org/cadence/client"
	"go.uber.org/zap"
)

func runUpdate() {
	logger := BuildLogger()
	c := buildScheduleClient()
	sc := c.ScheduleClient()

	const newCron = "*/2 * * * *"
	if err := sc.Update(context.Background(), ScheduleID, func(u *client.ScheduleUpdate) error {
		if u.Spec == nil {
			u.Spec = &client.ScheduleSpec{}
		}
		u.Spec.CronExpression = newCron
		return nil
	}); err != nil {
		logger.Fatal("Update failed", zap.Error(err))
	}
	fmt.Printf("Updated schedule %q cron=%q\n", ScheduleID, newCron)
}
