package main

import (
	"context"
	"math/rand"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/zap"
)

const (
	loadGenerationActivityName = "loadGenerationActivity"
)

// LoadGenerationActivity simulates work that can be scaled
// It includes random delays to simulate real-world processing time
func LoadGenerationActivity(ctx context.Context, taskID int, minProcessingTime, maxProcessingTime int) error {
	startTime := time.Now()
	logger := activity.GetLogger(ctx)
	logger.Info("Load generation activity started", zap.Int("taskID", taskID))

	// Simulate variable processing time using configuration values
	processingTime := time.Duration(rand.Intn(maxProcessingTime - minProcessingTime) + minProcessingTime) * time.Millisecond
	time.Sleep(processingTime)

	duration := time.Since(startTime)

	logger.Info("Load generation activity completed",
		zap.Int("taskID", taskID),
		zap.Duration("processingTime", processingTime),
		zap.Duration("totalDuration", duration))

	return nil
}
