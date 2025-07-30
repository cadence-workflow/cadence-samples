package main

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.uber.org/cadence/encoded"
)

// LargePayload represents a complex data structure with nested objects and arrays
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

// Item represents a single item in the payload
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

// Review represents a product review
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

// Inventory represents inventory information
type Inventory struct {
	Quantity    int    `json:"quantity"`
	Location    string `json:"location"`
	LastUpdated string `json:"last_updated"`
	Status      string `json:"status"`
}

// Config represents configuration settings
type Config struct {
	Version     string            `json:"version"`
	Environment string            `json:"environment"`
	Settings    map[string]string `json:"settings"`
	Features    []string          `json:"features"`
	Limits      Limits            `json:"limits"`
}

// Limits represents system limits
type Limits struct {
	MaxItems    int `json:"max_items"`
	MaxRequests int `json:"max_requests_per_minute"`
	MaxFileSize int `json:"max_file_size_mb"`
	MaxUsers    int `json:"max_concurrent_users"`
	TimeoutSecs int `json:"timeout_seconds"`
}

// HistoryEntry represents a historical event
type HistoryEntry struct {
	EventID   string            `json:"event_id"`
	Timestamp string            `json:"timestamp"`
	EventType string            `json:"event_type"`
	UserID    string            `json:"user_id"`
	Details   map[string]string `json:"details"`
	Severity  string            `json:"severity"`
}

// Statistics represents statistical data
type Statistics struct {
	TotalItems     int     `json:"total_items"`
	TotalUsers     int     `json:"total_users"`
	AverageRating  float64 `json:"average_rating"`
	TotalRevenue   float64 `json:"total_revenue"`
	ActiveOrders   int     `json:"active_orders"`
	CompletionRate float64 `json:"completion_rate"`
}

// CreateLargePayload creates a sample large payload with realistic data
func CreateLargePayload() LargePayload {
	// Create a large description with repeated text to demonstrate compression
	largeDescription := strings.Repeat("This is a comprehensive product catalog containing thousands of items with detailed descriptions, specifications, and user reviews. Each item includes pricing information, inventory status, and customer feedback. The catalog is designed to provide complete information for customers making purchasing decisions. ", 50)

	// Create sample items
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

	// Create history entries
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

	// Create metadata
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

// GetPayloadSizeInfo returns information about the payload size before and after compression
func GetPayloadSizeInfo(payload LargePayload, converter encoded.DataConverter) (int, int, float64, error) {
	// Serialize to JSON to get original size
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to marshal payload: %v", err)
	}
	originalSize := len(jsonData)

	// Compress using our converter
	compressedData, err := converter.ToData(payload)
	if err != nil {
		return 0, 0, 0, fmt.Errorf("failed to compress payload: %v", err)
	}
	compressedSize := len(compressedData)

	// Calculate compression ratio
	compressionRatio := float64(compressedSize) / float64(originalSize)
	compressionPercentage := (1.0 - compressionRatio) * 100

	return originalSize, compressedSize, compressionPercentage, nil
}
