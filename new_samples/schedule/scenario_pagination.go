package main

import (
	"context"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/zap"
)

// runPagination demonstrates List pagination.
//
// It creates several schedules, then pages through ALL schedules in the domain using the
// returned NextPageToken with a small page size, and verifies each created schedule appears
// exactly once across pages (no duplicates, no gaps, final token empty).
func runPagination() {
	logger := BuildLogger()
	c := buildScheduleClient(nil)
	sc := c.ScheduleClient()
	ctx := context.Background()

	const n = 5
	created := make([]string, 0, n)
	for i := 0; i < n; i++ {
		id := newScheduleID("sample-page")
		if _, err := sc.Create(ctx, &client.CreateScheduleRequest{
			ScheduleID: id,
			Spec:       &client.ScheduleSpec{CronExpression: "0 0 1 1 *"}, // yearly: never fires during the demo
			Action:     &client.ScheduleAction{StartWorkflow: startWorkflowAction(logger, 0)},
		}); err != nil {
			logger.Fatal("Create failed", zap.String("scheduleID", id), zap.Error(err))
		}
		created = append(created, id)
	}
	defer func() {
		for _, id := range created {
			deleteQuietly(logger, sc, context.Background(), id)
		}
	}()
	logger.Info("Created schedules for pagination demo", zap.Int("count", n))

	// Give the index a moment to catch up (List is visibility-backed and async).
	time.Sleep(3 * time.Second)

	const pageSize = 2
	seen := make(map[string]int)
	var token []byte
	pages := 0
	for {
		resp, err := sc.List(ctx, pageSize, token)
		if err != nil {
			logger.Fatal("List failed", zap.Error(err))
		}
		pages++
		for _, e := range resp.Schedules {
			seen[e.ScheduleID]++
		}
		logger.Info("Listed page",
			zap.Int("page", pages),
			zap.Int("entries", len(resp.Schedules)),
			zap.Bool("hasNextToken", len(resp.NextPageToken) > 0))
		if len(resp.NextPageToken) == 0 {
			break
		}
		token = resp.NextPageToken
	}

	allOnce := true
	for _, id := range created {
		switch seen[id] {
		case 1:
			// good
		case 0:
			allOnce = false
			logger.Warn("schedule missing from paginated List (indexing may lag)", zap.String("scheduleID", id))
		default:
			allOnce = false
			logger.Warn("schedule appeared more than once across pages",
				zap.String("scheduleID", id), zap.Int("times", seen[id]))
		}
	}
	if allOnce {
		logger.Info("Pagination verified: every created schedule appeared exactly once",
			zap.Int("createdSchedules", n), zap.Int("pages", pages))
	}
}
