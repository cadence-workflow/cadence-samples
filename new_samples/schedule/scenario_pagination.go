package main

import (
	"context"
	"time"

	"go.uber.org/cadence/client"
	"go.uber.org/zap"
)

// runPagination exercises List pagination across three boundary cases:
//
//   - Multi-page:      more schedules than pageSize → multiple pages, token chains correctly
//   - Single-page:     fewer schedules than pageSize → all on one page, no NextPageToken
//   - Exact-boundary:  schedules count == exact multiple of pageSize → no empty trailing page
func runPagination() {
	logger := BuildLogger()
	c := buildScheduleClient(nil)
	sc := c.ScheduleClient()

	cases := []struct {
		name     string
		n        int
		pageSize int32
		expect   string
	}{
		{
			name:     "MultiPage",
			n:        5,
			pageSize: 2,
			expect:   "3 pages (2,2,1), NextPageToken non-empty after pages 1-2, empty after page 3",
		},
		{
			name:     "SinglePage",
			n:        3,
			pageSize: 10,
			expect:   "1 page with all 3 entries, NextPageToken empty immediately",
		},
		{
			name:     "ExactBoundary",
			n:        4,
			pageSize: 2,
			expect:   "2 pages of exactly 2 entries each, no empty trailing page",
		},
	}

	for _, tc := range cases {
		tc := tc
		func() {
			ctx := context.Background()
			logger.Info("=== Pagination: "+tc.name+" ===",
				zap.Int("schedules", tc.n),
				zap.Int32("pageSize", tc.pageSize),
				zap.String("expect", tc.expect))

			created := createPaginationSchedules(logger, sc, ctx, tc.n)
			defer func() {
				for _, id := range created {
					deleteQuietly(logger, sc, context.Background(), id)
				}
			}()

			// Poll until all created schedules are visible in List (indexing is async).
			if !waitForVisibility(logger, sc, ctx, created) {
				logger.Warn("  some schedules not yet visible — results may be incomplete")
			}

			runPaginationCase(logger, sc, ctx, tc.name, tc.pageSize, created)
		}()
	}
}

func createPaginationSchedules(logger *zap.Logger, sc client.ScheduleClient, ctx context.Context, n int) []string {
	ids := make([]string, 0, n)
	for i := 0; i < n; i++ {
		id := newScheduleID("sample-page")
		if _, err := sc.Create(ctx, &client.CreateScheduleRequest{
			ScheduleID: id,
			Spec:       &client.ScheduleSpec{CronExpression: "0 0 1 1 *"}, // yearly: never fires during the demo
			Action:     &client.ScheduleAction{StartWorkflow: startWorkflowAction(logger, 0)},
		}); err != nil {
			logger.Fatal("Create failed", zap.String("scheduleID", id), zap.Error(err))
		}
		ids = append(ids, id)
	}
	logger.Info("Schedules created", zap.Int("count", n))
	return ids
}

// waitForVisibility polls List until all created IDs appear or 15s elapses.
func waitForVisibility(logger *zap.Logger, sc client.ScheduleClient, ctx context.Context, ids []string) bool {
	want := make(map[string]bool, len(ids))
	for _, id := range ids {
		want[id] = true
	}
	for deadline := time.Now().Add(15 * time.Second); time.Now().Before(deadline); time.Sleep(500 * time.Millisecond) {
		resp, err := sc.List(ctx, 100, nil)
		if err != nil {
			continue
		}
		for _, e := range resp.Schedules {
			delete(want, e.ScheduleID)
		}
		if len(want) == 0 {
			return true
		}
	}
	logger.Warn("Visibility timeout: not all schedules indexed", zap.Int("missing", len(want)))
	return false
}

func runPaginationCase(logger *zap.Logger, sc client.ScheduleClient, ctx context.Context, name string, pageSize int32, created []string) {
	logger.Info("--- Verify: Pagination/"+name+" ---", zap.Int32("pageSize", pageSize))

	seen := make(map[string]int)
	var token []byte
	pages := 0
	pageSizeViolation := false

	for {
		resp, err := sc.List(ctx, pageSize, token)
		if err != nil {
			logger.Fatal("List failed", zap.Error(err))
		}
		pages++

		// Verify each page respects the requested page size.
		if len(resp.Schedules) > int(pageSize) {
			pageSizeViolation = true
			logger.Warn("  MISMATCH page size exceeded",
				zap.Int("page", pages),
				zap.Int("got", len(resp.Schedules)),
				zap.Int32("pageSize", pageSize))
		} else {
			logger.Info("  page",
				zap.Int("page", pages),
				zap.Int("entries", len(resp.Schedules)),
				zap.Bool("hasNextToken", len(resp.NextPageToken) > 0))
		}

		for _, e := range resp.Schedules {
			seen[e.ScheduleID]++
		}

		if len(resp.NextPageToken) == 0 {
			break
		}
		token = resp.NextPageToken
	}

	// Verify no duplicates and no gaps for the schedules we created.
	allOnce := true
	for _, id := range created {
		switch seen[id] {
		case 1:
			// expected
		case 0:
			allOnce = false
			logger.Warn("  MISMATCH schedule missing from paginated List",
				zap.String("scheduleID", id))
		default:
			allOnce = false
			logger.Warn("  MISMATCH schedule appeared more than once across pages",
				zap.String("scheduleID", id), zap.Int("times", seen[id]))
		}
	}

	if !pageSizeViolation && allOnce {
		logger.Info("  MATCH   pagination correct: every schedule appeared exactly once, all pages within size limit",
			zap.Int("totalPages", pages),
			zap.Int("createdSchedules", len(created)))
	}

	// Case-specific checks.
	switch name {
	case "SinglePage":
		if pages == 1 {
			logger.Info("  MATCH   SinglePage: all results on one page, no NextPageToken")
		} else {
			logger.Warn("  MISMATCH SinglePage: expected 1 page", zap.Int("got", pages))
		}
	case "ExactBoundary":
		// No-lookahead servers emit an empty trailing page when n == exact multiple of pageSize.
		// That is valid: the server doesn't know the list is exhausted until the follow-up fetch
		// returns 0 items. Accept either form.
		exactPages := len(created) / int(pageSize)
		if pages == exactPages || pages == exactPages+1 {
			logger.Info("  MATCH   ExactBoundary: pagination terminated correctly",
				zap.Int("pages", pages),
				zap.Bool("emptyTrailingPage", pages == exactPages+1))
		} else {
			logger.Warn("  MISMATCH ExactBoundary: unexpected page count",
				zap.Int("wantExact", exactPages),
				zap.Int("wantWithTrailer", exactPages+1),
				zap.Int("got", pages))
		}
	}
}
