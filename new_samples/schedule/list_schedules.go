package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

func runList() {
	logger := BuildLogger()
	c := buildScheduleClient()
	sc := c.ScheduleClient()

	count := 0
	var token []byte
	for {
		resp, err := sc.List(context.Background(), 100, token)
		if err != nil {
			logger.Fatal("List failed", zap.Error(err))
		}
		for _, entry := range resp.Schedules {
			paused := ""
			if entry.State != nil && entry.State.Paused {
				paused = "  [paused]"
			}
			fmt.Printf("  %s%s\n", entry.ScheduleID, paused)
			count++
		}
		if len(resp.NextPageToken) == 0 {
			break
		}
		token = resp.NextPageToken
	}
	if count == 0 {
		fmt.Println("(no schedules found)")
	} else {
		fmt.Printf("\n%d schedule(s) in domain %q\n", count, Domain)
	}
}
