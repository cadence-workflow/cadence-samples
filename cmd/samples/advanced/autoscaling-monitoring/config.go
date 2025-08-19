package main

import (
	"github.com/uber-common/cadence-samples/cmd/samples/common"
	"github.com/uber-go/tally/prometheus"
)

// AutoscalingConfiguration extends the base Configuration with autoscaling-specific settings
type AutoscalingConfiguration struct {
	common.Configuration
	Autoscaling AutoscalingSettings       `yaml:"autoscaling"`
	Prometheus  *prometheus.Configuration `yaml:"prometheus"`
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
	Iterations        int `yaml:"iterations"`
	BatchDelay        int `yaml:"batchDelay"`
	MinProcessingTime int `yaml:"minProcessingTime"`
	MaxProcessingTime int `yaml:"maxProcessingTime"`
}

// DefaultAutoscalingConfiguration returns default autoscaling settings
func DefaultAutoscalingConfiguration() AutoscalingConfiguration {
	return AutoscalingConfiguration{
		Configuration: common.Configuration{
			DomainName:      "default",
			ServiceName:     "cadence-frontend",
			HostNameAndPort: "localhost:7833",
		},
		Prometheus: &prometheus.Configuration{
			ListenAddress: "127.0.0.1:8004",
		},
		Autoscaling: AutoscalingSettings{
			PollerMinCount:  2,
			PollerMaxCount:  8,
			PollerInitCount: 4,
			LoadGeneration: LoadGenerationSettings{
				Iterations:        50,
				BatchDelay:        2,
				MinProcessingTime: 1000,
				MaxProcessingTime: 6000,
			},
		},
	}
}
