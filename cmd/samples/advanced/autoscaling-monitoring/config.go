package main

import (
	"fmt"
	"os"

	"github.com/uber-common/cadence-samples/cmd/samples/common"
	"github.com/uber-go/tally/prometheus"
	"gopkg.in/yaml.v3"
)

// AutoscalingConfiguration uses a flattened structure to avoid embedded struct issues
type AutoscalingConfiguration struct {
	// Base configuration fields (explicit, not embedded)
	DomainName      string                    `yaml:"domain"`
	ServiceName     string                    `yaml:"service"`
	HostNameAndPort string                    `yaml:"host"`
	Prometheus      *prometheus.Configuration `yaml:"prometheus"`

	// Autoscaling-specific fields
	Autoscaling AutoscalingSettings `yaml:"autoscaling"`
}

// AutoscalingSettings contains the autoscaling configuration
type AutoscalingSettings struct {
	// Worker autoscaling settings
	PollerMinCount  int `yaml:"pollerMinCount"`
	PollerMaxCount  int `yaml:"pollerMaxCount"`
	PollerInitCount int `yaml:"pollerInitCount"`

	// Load generation settings
	LoadGeneration LoadGenerationSettings `yaml:"loadGeneration"`
}

// LoadGenerationSettings contains the load generation configuration
type LoadGenerationSettings struct {
	// Workflow-level settings
	Workflows     int `yaml:"workflows"`
	WorkflowDelay int `yaml:"workflowDelay"`

	// Activity-level settings (per workflow)
	ActivitiesPerWorkflow int `yaml:"activitiesPerWorkflow"`
	BatchDelay            int `yaml:"batchDelay"`
	MinProcessingTime     int `yaml:"minProcessingTime"`
	MaxProcessingTime     int `yaml:"maxProcessingTime"`
}

// Default values as constants for easy maintenance
const (
	DefaultDomainName      = "default"
	DefaultServiceName     = "cadence-frontend"
	DefaultHostNameAndPort = "localhost:7833"
	DefaultPrometheusAddr  = "127.0.0.1:8004"

	DefaultPollerMinCount  = 2
	DefaultPollerMaxCount  = 8
	DefaultPollerInitCount = 4

	DefaultWorkflows             = 3
	DefaultWorkflowDelay         = 1000
	DefaultActivitiesPerWorkflow = 40
	DefaultBatchDelay            = 2000
	DefaultMinProcessingTime     = 1000
	DefaultMaxProcessingTime     = 6000
)

// DefaultAutoscalingConfiguration returns default configuration
func DefaultAutoscalingConfiguration() AutoscalingConfiguration {
	return AutoscalingConfiguration{
		DomainName:      DefaultDomainName,
		ServiceName:     DefaultServiceName,
		HostNameAndPort: DefaultHostNameAndPort,
		Prometheus: &prometheus.Configuration{
			ListenAddress: DefaultPrometheusAddr,
		},
		Autoscaling: AutoscalingSettings{
			PollerMinCount:  DefaultPollerMinCount,
			PollerMaxCount:  DefaultPollerMaxCount,
			PollerInitCount: DefaultPollerInitCount,
			LoadGeneration: LoadGenerationSettings{
				Workflows:             DefaultWorkflows,
				WorkflowDelay:         DefaultWorkflowDelay,
				ActivitiesPerWorkflow: DefaultActivitiesPerWorkflow,
				BatchDelay:            DefaultBatchDelay,
				MinProcessingTime:     DefaultMinProcessingTime,
				MaxProcessingTime:     DefaultMaxProcessingTime,
			},
		},
	}
}

// loadConfiguration loads the autoscaling configuration from file
func loadConfiguration(configFile string) AutoscalingConfiguration {
	// Start with defaults
	config := DefaultAutoscalingConfiguration()

	// Read config file
	configData, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Failed to read config file: %v, using defaults\n", err)
		return config
	}

	// Unmarshal into the config struct
	if err := yaml.Unmarshal(configData, &config); err != nil {
		fmt.Printf("Error parsing configuration: %v, using defaults\n", err)
		return DefaultAutoscalingConfiguration()
	}

	// Apply defaults for any missing fields
	config.applyDefaults()

	return config
}

// applyDefaults ensures all fields have sensible values
func (c *AutoscalingConfiguration) applyDefaults() {
	// Base configuration defaults
	if c.DomainName == "" {
		c.DomainName = DefaultDomainName
	}
	if c.ServiceName == "" {
		c.ServiceName = DefaultServiceName
	}
	if c.HostNameAndPort == "" {
		c.HostNameAndPort = DefaultHostNameAndPort
	}
	if c.Prometheus == nil {
		c.Prometheus = &prometheus.Configuration{
			ListenAddress: DefaultPrometheusAddr,
		}
	}

	// Autoscaling defaults
	if c.Autoscaling.PollerMinCount == 0 {
		c.Autoscaling.PollerMinCount = DefaultPollerMinCount
	}
	if c.Autoscaling.PollerMaxCount == 0 {
		c.Autoscaling.PollerMaxCount = DefaultPollerMaxCount
	}
	if c.Autoscaling.PollerInitCount == 0 {
		c.Autoscaling.PollerInitCount = DefaultPollerInitCount
	}

	// Load generation defaults
	if c.Autoscaling.LoadGeneration.Workflows == 0 {
		c.Autoscaling.LoadGeneration.Workflows = DefaultWorkflows
	}
	if c.Autoscaling.LoadGeneration.WorkflowDelay == 0 {
		c.Autoscaling.LoadGeneration.WorkflowDelay = DefaultWorkflowDelay
	}
	if c.Autoscaling.LoadGeneration.ActivitiesPerWorkflow == 0 {
		c.Autoscaling.LoadGeneration.ActivitiesPerWorkflow = DefaultActivitiesPerWorkflow
	}
	if c.Autoscaling.LoadGeneration.BatchDelay == 0 {
		c.Autoscaling.LoadGeneration.BatchDelay = DefaultBatchDelay
	}
	if c.Autoscaling.LoadGeneration.MinProcessingTime == 0 {
		c.Autoscaling.LoadGeneration.MinProcessingTime = DefaultMinProcessingTime
	}
	if c.Autoscaling.LoadGeneration.MaxProcessingTime == 0 {
		c.Autoscaling.LoadGeneration.MaxProcessingTime = DefaultMaxProcessingTime
	}
}

// ToCommonConfiguration converts to the common.Configuration type for compatibility
func (c *AutoscalingConfiguration) ToCommonConfiguration() common.Configuration {
	return common.Configuration{
		DomainName:      c.DomainName,
		ServiceName:     c.ServiceName,
		HostNameAndPort: c.HostNameAndPort,
		Prometheus:      c.Prometheus,
	}
}
