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
    workflows: 5
    workflowDelay: 3
    activitiesPerWorkflow: 100
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
	assert.Equal(t, 5, config.Autoscaling.LoadGeneration.Workflows)
	assert.Equal(t, 3, config.Autoscaling.LoadGeneration.WorkflowDelay)
	assert.Equal(t, 100, config.Autoscaling.LoadGeneration.ActivitiesPerWorkflow)
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

func TestLoadConfiguration_PartialConfiguration(t *testing.T) {
	testCases := []struct {
		name               string
		configContent      string
		expectedDomain     string
		expectedService    string
		expectedHost       string
		expectedPollerMin  int
		expectedPollerMax  int
		expectedPollerInit int
		description        string
	}{
		{
			name: "Only domain specified",
			configContent: `
domain: "custom-domain"
`,
			expectedDomain:     "custom-domain",
			expectedService:    DefaultServiceName,
			expectedHost:       DefaultHostNameAndPort,
			expectedPollerMin:  DefaultPollerMinCount,
			expectedPollerMax:  DefaultPollerMaxCount,
			expectedPollerInit: DefaultPollerInitCount,
			description:        "Only domain field provided",
		},
		{
			name: "Only autoscaling settings specified",
			configContent: `
autoscaling:
  pollerMinCount: 5
  pollerMaxCount: 15
  pollerInitCount: 8
`,
			expectedDomain:     DefaultDomainName,
			expectedService:    DefaultServiceName,
			expectedHost:       DefaultHostNameAndPort,
			expectedPollerMin:  5,
			expectedPollerMax:  15,
			expectedPollerInit: 8,
			description:        "Only autoscaling fields provided",
		},
		{
			name: "Mixed configuration",
			configContent: `
domain: "mixed-domain"
service: "mixed-service"
autoscaling:
  pollerMinCount: 1
  loadGeneration:
    workflows: 5
    activitiesPerWorkflow: 75
    batchDelay: 3
`,
			expectedDomain:     "mixed-domain",
			expectedService:    "mixed-service",
			expectedHost:       DefaultHostNameAndPort,
			expectedPollerMin:  1,
			expectedPollerMax:  DefaultPollerMaxCount,
			expectedPollerInit: DefaultPollerInitCount,
			description:        "Mix of provided and default values",
		},
		{
			name:               "Empty file",
			configContent:      "",
			expectedDomain:     DefaultDomainName,
			expectedService:    DefaultServiceName,
			expectedHost:       DefaultHostNameAndPort,
			expectedPollerMin:  DefaultPollerMinCount,
			expectedPollerMax:  DefaultPollerMaxCount,
			expectedPollerInit: DefaultPollerInitCount,
			description:        "Empty YAML file",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tc.configContent)
			require.NoError(t, err)
			tmpFile.Close()

			// Load configuration
			config := loadConfiguration(tmpFile.Name())

			// Validate expected values
			assert.Equal(t, tc.expectedDomain, config.DomainName)
			assert.Equal(t, tc.expectedService, config.ServiceName)
			assert.Equal(t, tc.expectedHost, config.HostNameAndPort)
			assert.Equal(t, tc.expectedPollerMin, config.Autoscaling.PollerMinCount)
			assert.Equal(t, tc.expectedPollerMax, config.Autoscaling.PollerMaxCount)
			assert.Equal(t, tc.expectedPollerInit, config.Autoscaling.PollerInitCount)

			// Validate that other fields have appropriate defaults
			if tc.configContent == "" || tc.configContent == `
domain: "custom-domain"
` {
				// Should have all defaults for load generation
				assert.Equal(t, DefaultWorkflows, config.Autoscaling.LoadGeneration.Workflows)
				assert.Equal(t, DefaultWorkflowDelay, config.Autoscaling.LoadGeneration.WorkflowDelay)
				assert.Equal(t, DefaultActivitiesPerWorkflow, config.Autoscaling.LoadGeneration.ActivitiesPerWorkflow)
				assert.Equal(t, DefaultBatchDelay, config.Autoscaling.LoadGeneration.BatchDelay)
				assert.Equal(t, DefaultMinProcessingTime, config.Autoscaling.LoadGeneration.MinProcessingTime)
				assert.Equal(t, DefaultMaxProcessingTime, config.Autoscaling.LoadGeneration.MaxProcessingTime)
			}
		})
	}
}

func TestLoadConfiguration_MalformedYAML(t *testing.T) {
	testCases := []struct {
		name          string
		configContent string
		description   string
	}{
		{
			name: "Invalid YAML syntax",
			configContent: `
domain: "test-domain"
service: "test-service"
host: "test-host:7833"
autoscaling:
  pollerMinCount: 3
  pollerMaxCount: 10
  pollerInitCount: 5
  loadGeneration:
    workflows: 5
    activitiesPerWorkflow: 100
    batchDelay: 5
    minProcessingTime: 2000
    maxProcessingTime: 8000
invalid: yaml: syntax: here
`,
			description: "YAML with syntax error",
		},
		{
			name: "Invalid field types",
			configContent: `
domain: "test-domain"
service: "test-service"
host: "test-host:7833"
autoscaling:
  pollerMinCount: "not-a-number"
  pollerMaxCount: 10
  pollerInitCount: 5
  loadGeneration:
    workflows: 5
    activitiesPerWorkflow: 100
    batchDelay: 5
    minProcessingTime: 2000
    maxProcessingTime: 8000
`,
			description: "YAML with invalid field types",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temporary file
			tmpFile, err := os.CreateTemp("", "test-config-*.yaml")
			require.NoError(t, err)
			defer os.Remove(tmpFile.Name())

			_, err = tmpFile.WriteString(tc.configContent)
			require.NoError(t, err)
			tmpFile.Close()

			// Load configuration - should not panic and should return defaults
			config := loadConfiguration(tmpFile.Name())

			// Validate that default configuration is returned
			assert.Equal(t, DefaultDomainName, config.DomainName)
			assert.Equal(t, DefaultServiceName, config.ServiceName)
			assert.Equal(t, DefaultHostNameAndPort, config.HostNameAndPort)
			assert.Equal(t, DefaultPollerMinCount, config.Autoscaling.PollerMinCount)
			assert.Equal(t, DefaultPollerMaxCount, config.Autoscaling.PollerMaxCount)
			assert.Equal(t, DefaultPollerInitCount, config.Autoscaling.PollerInitCount)
		})
	}
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
