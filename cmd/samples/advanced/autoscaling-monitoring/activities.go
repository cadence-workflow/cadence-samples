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
func LoadGenerationActivity(ctx context.Context, taskID int) error {
	startTime := time.Now()
	logger := activity.GetLogger(ctx)
	logger.Info("Load generation activity started", zap.Int("taskID", taskID))

	// Simulate variable processing time using configuration values
	minTime := config.Autoscaling.LoadGeneration.MinProcessingTime
	maxTime := config.Autoscaling.LoadGeneration.MaxProcessingTime
	processingTime := time.Duration(rand.Intn(maxTime-minTime)+minTime) * time.Millisecond
	time.Sleep(processingTime)

	duration := time.Since(startTime)

	// Record metrics for monitoring
	RecordActivityCompleted("autoscaling-worker-1", "load_generation", duration)

	logger.Info("Load generation activity completed",
		zap.Int("taskID", taskID),
		zap.Duration("processingTime", processingTime),
		zap.Duration("totalDuration", duration))

	return nil
}
