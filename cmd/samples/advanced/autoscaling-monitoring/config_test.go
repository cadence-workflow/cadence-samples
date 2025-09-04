package main

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test the improved configuration loader for regressions
func TestLoadConfiguration_SuccessfulLoading(t *testing.T) {
	// Create a temporary configuration file with all fields populated
	configContent := `
domain: "test-domain"
service: "test-service"
host: "test-host:7833"
prometheus:
  listenAddress: "127.0.0.1:9000"
autoscaling:
  pollerMinCount: 3
  pollerMaxCount: 10
  pollerInitCount: 5
  loadGeneration:
    workflows: 10
    workflowDelay: 1000
    activitiesPerWorkflow: 30
    batchDelay: 5
    minProcessingTime: 2000
    maxProcessingTime: 8000
`

	// Create temporary file
	tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	_, err = tmpFile.WriteString(configContent)
	require.NoError(t, err)
	tmpFile.Close()

	// Load configuration
	config := loadConfiguration(tmpFile.Name())

	// Validate all fields are populated correctly
	assert.Equal(t, "test-domain", config.DomainName)
	assert.Equal(t, "test-service", config.ServiceName)
	assert.Equal(t, "test-host:7833", config.HostNameAndPort)
	require.NotNil(t, config.Prometheus)
	assert.Equal(t, "127.0.0.1:9000", config.Prometheus.ListenAddress)
	assert.Equal(t, 3, config.Autoscaling.PollerMinCount)
	assert.Equal(t, 10, config.Autoscaling.PollerMaxCount)
	assert.Equal(t, 5, config.Autoscaling.PollerInitCount)
	assert.Equal(t, 10, config.Autoscaling.LoadGeneration.Workflows)
	assert.Equal(t, 1000, config.Autoscaling.LoadGeneration.WorkflowDelay)
	assert.Equal(t, 30, config.Autoscaling.LoadGeneration.ActivitiesPerWorkflow)
	assert.Equal(t, 5, config.Autoscaling.LoadGeneration.BatchDelay)
	assert.Equal(t, 2000, config.Autoscaling.LoadGeneration.MinProcessingTime)
	assert.Equal(t, 8000, config.Autoscaling.LoadGeneration.MaxProcessingTime)
}

func TestLoadConfiguration_MissingFileFallback(t *testing.T) {
	// Use a non-existent file path
	config := loadConfiguration("/non/existent/path/config.yaml")

	// Validate that default configuration is returned
	assert.Equal(t, DefaultDomainName, config.DomainName)
	assert.Equal(t, DefaultServiceName, config.ServiceName)
	assert.Equal(t, DefaultHostNameAndPort, config.HostNameAndPort)
	assert.Equal(t, DefaultPollerMinCount, config.Autoscaling.PollerMinCount)
	assert.Equal(t, DefaultPollerMaxCount, config.Autoscaling.PollerMaxCount)
	assert.Equal(t, DefaultPollerInitCount, config.Autoscaling.PollerInitCount)
	assert.Equal(t, DefaultWorkflows, config.Autoscaling.LoadGeneration.Workflows)
	assert.Equal(t, DefaultWorkflowDelay, config.Autoscaling.LoadGeneration.WorkflowDelay)
	assert.Equal(t, DefaultActivitiesPerWorkflow, config.Autoscaling.LoadGeneration.ActivitiesPerWorkflow)
	assert.Equal(t, DefaultBatchDelay, config.Autoscaling.LoadGeneration.BatchDelay)
	assert.Equal(t, DefaultMinProcessingTime, config.Autoscaling.LoadGeneration.MinProcessingTime)
	assert.Equal(t, DefaultMaxProcessingTime, config.Autoscaling.LoadGeneration.MaxProcessingTime)
}

func TestDefaultAutoscalingConfiguration(t *testing.T) {
	config := DefaultAutoscalingConfiguration()

	// Validate all default values
	assert.Equal(t, DefaultDomainName, config.DomainName)
	assert.Equal(t, DefaultServiceName, config.ServiceName)
	assert.Equal(t, DefaultHostNameAndPort, config.HostNameAndPort)
	require.NotNil(t, config.Prometheus)
	assert.Equal(t, DefaultPrometheusAddr, config.Prometheus.ListenAddress)
	assert.Equal(t, DefaultPollerMinCount, config.Autoscaling.PollerMinCount)
	assert.Equal(t, DefaultPollerMaxCount, config.Autoscaling.PollerMaxCount)
	assert.Equal(t, DefaultPollerInitCount, config.Autoscaling.PollerInitCount)
	assert.Equal(t, DefaultWorkflows, config.Autoscaling.LoadGeneration.Workflows)
	assert.Equal(t, DefaultWorkflowDelay, config.Autoscaling.LoadGeneration.WorkflowDelay)
	assert.Equal(t, DefaultActivitiesPerWorkflow, config.Autoscaling.LoadGeneration.ActivitiesPerWorkflow)
	assert.Equal(t, DefaultBatchDelay, config.Autoscaling.LoadGeneration.BatchDelay)
	assert.Equal(t, DefaultMinProcessingTime, config.Autoscaling.LoadGeneration.MinProcessingTime)
	assert.Equal(t, DefaultMaxProcessingTime, config.Autoscaling.LoadGeneration.MaxProcessingTime)
}

func TestApplyDefaults(t *testing.T) {
	// Test with empty configuration
	config := AutoscalingConfiguration{}
	config.applyDefaults()

	// Validate that all defaults are applied
	assert.Equal(t, DefaultDomainName, config.DomainName)
	assert.Equal(t, DefaultServiceName, config.ServiceName)
	assert.Equal(t, DefaultHostNameAndPort, config.HostNameAndPort)
	require.NotNil(t, config.Prometheus)
	assert.Equal(t, DefaultPrometheusAddr, config.Prometheus.ListenAddress)
	assert.Equal(t, DefaultPollerMinCount, config.Autoscaling.PollerMinCount)
	assert.Equal(t, DefaultPollerMaxCount, config.Autoscaling.PollerMaxCount)
	assert.Equal(t, DefaultPollerInitCount, config.Autoscaling.PollerInitCount)
	assert.Equal(t, DefaultWorkflows, config.Autoscaling.LoadGeneration.Workflows)
	assert.Equal(t, DefaultWorkflowDelay, config.Autoscaling.LoadGeneration.WorkflowDelay)
	assert.Equal(t, DefaultActivitiesPerWorkflow, config.Autoscaling.LoadGeneration.ActivitiesPerWorkflow)
	assert.Equal(t, DefaultBatchDelay, config.Autoscaling.LoadGeneration.BatchDelay)
	assert.Equal(t, DefaultMinProcessingTime, config.Autoscaling.LoadGeneration.MinProcessingTime)
	assert.Equal(t, DefaultMaxProcessingTime, config.Autoscaling.LoadGeneration.MaxProcessingTime)
}
