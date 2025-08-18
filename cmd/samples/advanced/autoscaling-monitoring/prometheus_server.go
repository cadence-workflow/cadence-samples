package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/m3db/prometheus_client_golang/prometheus"
	"github.com/m3db/prometheus_client_golang/prometheus/promhttp"
	"github.com/uber-common/cadence-samples/cmd/samples/common"
	"go.uber.org/zap"
)

const (
	defaultPrometheusPort = ":8004"
)

var (
	// Custom metrics for autoscaling monitoring
	autoscalingWorkflowsStarted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "autoscaling_workflows_started_total",
			Help: "Total number of autoscaling workflows started",
		},
		[]string{"worker_id"},
	)

	autoscalingActivitiesCompleted = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "autoscaling_activities_completed_total",
			Help: "Total number of autoscaling activities completed",
		},
		[]string{"worker_id", "activity_type"},
	)

	autoscalingActivityDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "autoscaling_activity_duration_seconds",
			Help:    "Duration of autoscaling activities in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"worker_id", "activity_type"},
	)

	autoscalingWorkerLoad = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "autoscaling_worker_load",
			Help: "Current load on the autoscaling worker",
		},
		[]string{"worker_id", "load_type"},
	)
)

func init() {
	// Register custom metrics
	prometheus.MustRegister(autoscalingWorkflowsStarted)
	prometheus.MustRegister(autoscalingActivitiesCompleted)
	prometheus.MustRegister(autoscalingActivityDuration)
	prometheus.MustRegister(autoscalingWorkerLoad)
}

// startPrometheusHTTPServer starts an HTTP server to expose Prometheus metrics
func startPrometheusHTTPServer(h *common.SampleHelper) {
	port := defaultPrometheusPort
	if h.Config.Prometheus != nil && h.Config.Prometheus.ListenAddress != "" {
		port = h.Config.Prometheus.ListenAddress
	}

	h.Logger.Info("Starting Prometheus HTTP server", zap.String("port", port))

	// Set up HTTP handlers for Prometheus metrics
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Autoscaling Monitoring Sample - Prometheus Metrics Server\n")
		fmt.Fprintf(w, "Metrics available at: /metrics\n")
		fmt.Fprintf(w, "Health check at: /health\n")
		fmt.Fprintf(w, "\nDashboard: http://localhost:3000/d/dehkspwgabvuoc/cadence-client\n")
	})

	// Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK\n")
	})

	// Start server in goroutine
	go func() {
		if err := http.ListenAndServe(port, nil); err != nil {
			h.Logger.Error("Failed to start Prometheus server", zap.Error(err))
			os.Exit(1)
		}
	}()

	h.Logger.Info("Prometheus server started successfully", zap.String("port", port))
}

// RecordWorkflowStarted records when a workflow is started
func RecordWorkflowStarted(workerID string) {
	autoscalingWorkflowsStarted.WithLabelValues(workerID).Inc()
}

// RecordActivityCompleted records when an activity is completed
func RecordActivityCompleted(workerID, activityType string, duration time.Duration) {
	autoscalingActivitiesCompleted.WithLabelValues(workerID, activityType).Inc()
	autoscalingActivityDuration.WithLabelValues(workerID, activityType).Observe(duration.Seconds())
}
