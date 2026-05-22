package main

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"strings"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/encoded"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// compressedJSONDataConverter implements encoded.DataConverter with gzip compression.
// It serializes data to JSON, then compresses using gzip to reduce storage size.
type compressedJSONDataConverter struct{}

// NewCompressedJSONDataConverter creates a new compressed JSON data converter.
func NewCompressedJSONDataConverter() encoded.DataConverter {
	return &compressedJSONDataConverter{}
}

func (dc *compressedJSONDataConverter) ToData(value ...interface{}) ([]byte, error) {
	var jsonBuf bytes.Buffer
	enc := json.NewEncoder(&jsonBuf)
	for i, obj := range value {
		err := enc.Encode(obj)
		if err != nil {
			return nil, fmt.Errorf("unable to encode argument: %d, %v, with error: %v", i, reflect.TypeOf(obj), err)
		}
	}

	var compressedBuf bytes.Buffer
	gzipWriter := gzip.NewWriter(&compressedBuf)

	_, err := gzipWriter.Write(jsonBuf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("unable to compress data: %v", err)
	}

	err = gzipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("unable to close gzip writer: %v", err)
	}

	return compressedBuf.Bytes(), nil
}

func (dc *compressedJSONDataConverter) FromData(input []byte, valuePtr ...interface{}) error {
	// Handle empty input (e.g., when workflow is started without --input from CLI)
	if len(input) == 0 {
		return nil
	}

	gzipReader, err := gzip.NewReader(bytes.NewBuffer(input))
	if err != nil {
		return fmt.Errorf("unable to create gzip reader: %v", err)
	}
	defer gzipReader.Close()

	decompressedData, err := io.ReadAll(gzipReader)
	if err != nil {
		return fmt.Errorf("unable to decompress data: %v", err)
	}

	dec := json.NewDecoder(bytes.NewBuffer(decompressedData))
	for i, obj := range valuePtr {
		err := dec.Decode(obj)
		if err != nil {
			return fmt.Errorf("unable to decode argument: %d, %v, with error: %v", i, reflect.TypeOf(obj), err)
		}
	}
	return nil
}

// LargePayload represents a complex data structure with nested objects and arrays
// to demonstrate compression benefits.
type LargePayload struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Metadata    map[string]interface{} `json:"metadata"`
	Items       []Item                 `json:"items"`
	Config      Config                 `json:"config"`
	History     []HistoryEntry         `json:"history"`
	Tags        []string               `json:"tags"`
	Stats       Statistics             `json:"statistics"`
}

type Item struct {
	ItemID      string            `json:"item_id"`
	Title       string            `json:"title"`
	Description string            `json:"description"`
	Price       float64           `json:"price"`
	Categories  []string          `json:"categories"`
	Attributes  map[string]string `json:"attributes"`
	Reviews     []Review          `json:"reviews"`
	Inventory   Inventory         `json:"inventory"`
}

type Review struct {
	ReviewID   string  `json:"review_id"`
	UserID     string  `json:"user_id"`
	Rating     int     `json:"rating"`
	Comment    string  `json:"comment"`
	Helpful    int     `json:"helpful_votes"`
	NotHelpful int     `json:"not_helpful_votes"`
	Date       string  `json:"date"`
	Verified   bool    `json:"verified_purchase"`
	Score      float64 `json:"score"`
}

type Inventory struct {
	Quantity    int    `json:"quantity"`
	Location    string `json:"location"`
	LastUpdated string `json:"last_updated"`
	Status      string `json:"status"`
}

type Config struct {
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Settings    map[string]string `json:"settings"`
	Features    []string          `json:"features"`
	Limits      Limits            `json:"limits"`
}

type Limits struct {
	MaxItems    int `json:"max_items"`
	MaxRequests int `json:"max_requests_per_minute"`
	MaxFileSize int `json:"max_file_size_mb"`
	MaxUsers    int `json:"max_concurrent_users"`
	TimeoutSecs int `json:"timeout_seconds"`
}

type HistoryEntry struct {
	EventID   string            `json:"event_id"`
	Timestamp string            `json:"timestamp"`
	EventType string            `json:"event_type"`
	UserID    string            `json:"user_id"`
	Details   map[string]string `json:"details"`
	Severity  string            `json:"severity"`
}

type Statistics struct {
	TotalItems     int     `json:"total_items"`
	TotalUsers     int     `json:"total_users"`
	AverageRating  float64 `json:"average_rating"`
	TotalRevenue   float64 `json:"total_revenue"`
	ActiveOrders   int     `json:"active_orders"`
	CompletionRate float64 `json:"completion_rate"`
}

// CreateLargePayload creates a sample large payload with realistic data
// to demonstrate compression benefits.
func CreateLargePayload() LargePayload {
	largeDescription := strings.Repeat("This is a comprehensive product catalog containing thousands of items with detailed descriptions, specifications, and user reviews. Each item includes pricing information, inventory status, and customer feedback. The catalog is designed to provide complete information for customers making purchasing decisions. ", 50)

	items := make([]Item, 100)
	for i := 0; i < 100; i++ {
		reviews := make([]Review, 25)
		for j := 0; j < 25; j++ {
			reviews[j] = Review{
				ReviewID:   fmt.Sprintf("review_%d_%d", i, j),
				UserID:     fmt.Sprintf("user_%d", j),
				Rating:     1 + (j % 5),
				Comment:    strings.Repeat("This is a detailed customer review with comprehensive feedback about the product quality, delivery experience, and overall satisfaction. The customer provides specific details about their experience. ", 3),
				Helpful:    j * 2,
				NotHelpful: j,
				Date:       "2024-01-15T10:30:00Z",
				Verified:   j%2 == 0,
				Score:      float64(1+(j%5)) + float64(j%10)/10.0,
			}
		}

		attributes := make(map[string]string)
		for k := 0; k < 20; k++ {
			attributes[fmt.Sprintf("attr_%d", k)] = strings.Repeat("This is a detailed attribute description with comprehensive information about the product specification. ", 2)
		}

		items[i] = Item{
			ItemID:      fmt.Sprintf("item_%d", i),
			Title:       fmt.Sprintf("High-Quality Product %d with Advanced Features", i),
			Description: strings.Repeat("This is a premium product with exceptional quality and advanced features designed for professional use. It includes comprehensive documentation and support. ", 10),
			Price:       float64(100+i*10) + float64(i%100)/100.0,
			Categories:  []string{"Electronics", "Professional", "Premium", "Advanced"},
			Attributes:  attributes,
			Reviews:     reviews,
			Inventory: Inventory{
				Quantity:    100 + i,
				Location:    fmt.Sprintf("Warehouse %d", i%5),
				LastUpdated: "2024-01-15T10:30:00Z",
				Status:      "In Stock",
			},
		}
	}

	history := make([]HistoryEntry, 50)
	for i := 0; i < 50; i++ {
		details := make(map[string]string)
		for j := 0; j < 10; j++ {
			details[fmt.Sprintf("detail_%d", j)] = strings.Repeat("This is a detailed event description with comprehensive information about the system event and its impact. ", 2)
		}

		history[i] = HistoryEntry{
			EventID:   fmt.Sprintf("event_%d", i),
			Timestamp: "2024-01-15T10:30:00Z",
			EventType: "system_update",
			UserID:    fmt.Sprintf("admin_%d", i%5),
			Details:   details,
			Severity:  "medium",
		}
	}

	metadata := make(map[string]interface{})
	for i := 0; i < 30; i++ {
		metadata[fmt.Sprintf("meta_key_%d", i)] = strings.Repeat("This is comprehensive metadata information with detailed descriptions and specifications. ", 5)
	}

	return LargePayload{
		ID:          "large_payload_001",
		Name:        "Comprehensive Product Catalog",
		Description: largeDescription,
		Metadata:    metadata,
		Items:       items,
		Config: Config{
			Version:     "2.1.0",
			Environment: "production",
			Settings: map[string]string{
				"cache_enabled":     "true",
				"compression_level": "high",
				"timeout":           "30s",
				"max_connections":   "1000",
				"retry_attempts":    "3",
			},
			Features: []string{"advanced_search", "real_time_updates", "analytics", "reporting", "integration"},
			Limits: Limits{
				MaxItems:    10000,
				MaxRequests: 1000,
				MaxFileSize: 100,
				MaxUsers:    5000,
				TimeoutSecs: 30,
			},
		},
		History: history,
		Tags:    []string{"catalog", "products", "inventory", "analytics", "reporting", "integration", "api", "dashboard"},
		Stats: Statistics{
			TotalItems:     10000,
			TotalUsers:     5000,
			AverageRating:  4.2,
			TotalRevenue:   1250000.50,
			ActiveOrders:   250,
			CompletionRate: 98.5,
		},
	}
}

// GetPayloadSizeInfo returns information about the payload size before and after compression.
func GetPayloadSizeInfo(payload LargePayload, converter encoded.DataConverter) (int, int, float64, error) {
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to marshal payload: %v", err)
	}
	originalSize := len(jsonData)

	compressedData, err := converter.ToData(payload)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to compress payload: %v", err)
	}
	compressedSize := len(compressedData)

	compressionRatio := float64(compressedSize) / float64(originalSize)
	compressionPercentage := (1.0 - compressionRatio) * 100

	return originalSize, compressedSize, compressionPercentage, nil
}

// CompressionDataConverterWorkflow demonstrates processing large payloads with compression.
// The DataConverter automatically compresses/decompresses all workflow data.
// Note: The workflow generates its own payload internally so it can be started from CLI
// without requiring the CLI to use the custom DataConverter. The compression demonstration
// happens when data is passed between workflow and activity.
func CompressionDataConverterWorkflow(ctx workflow.Context) (LargePayload, error) {
	logger := workflow.GetLogger(ctx)

	// Generate the large payload internally - this allows the workflow to be started
	// from CLI without needing a custom DataConverter on the client side.
	// The compression benefit is demonstrated when passing data to/from activities.
	input := CreateLargePayload()

	logger.Info("Large payload workflow started", zap.String("payload_id", input.ID))
	logger.Info("Processing large payload with compression", zap.Int("items_count", len(input.Items)))

	activityOptions := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	var result LargePayload
	err := workflow.ExecuteActivity(ctx, CompressionDataConverterActivity, input).Get(ctx, &result)
	if err != nil {
		logger.Error("Large payload activity failed", zap.Error(err))
		return LargePayload{}, err
	}

	logger.Info("Large payload workflow completed", zap.String("result_id", result.ID))
	logger.Info("Note: All large payload data was automatically compressed/decompressed using gzip compression")
	return result, nil
}

// CompressionDataConverterActivity processes the large payload.
// In production, this might involve data transformation, validation, etc.
func CompressionDataConverterActivity(ctx context.Context, input LargePayload) (LargePayload, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Large payload activity received input", zap.String("payload_id", input.ID), zap.Int("items_count", len(input.Items)))

	input.Name = input.Name + " (Processed)"
	input.Stats.TotalItems = len(input.Items)

	logger.Info("Large payload activity completed", zap.String("result_id", input.ID))
	return input, nil
}
