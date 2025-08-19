package main

import (
	"net/http"
	"os"

	"github.com/m3db/prometheus_client_golang/prometheus/promhttp"
	"github.com/uber-common/cadence-samples/cmd/samples/common"
	"go.uber.org/zap"
)

const (
	defaultPrometheusPort = "127.0.0.1:8004"
)

// startPrometheusHTTPServer starts an HTTP server to expose Prometheus metrics
func startPrometheusHTTPServer(h *common.SampleHelper) {
	port := defaultPrometheusPort
	if h.Config.Prometheus != nil && h.Config.Prometheus.ListenAddress != "" {
		port = h.Config.Prometheus.ListenAddress
	}

	h.Logger.Info("Starting Prometheus HTTP server", zap.String("port", port))

	// Prometheus metrics endpoint - exposes standard Cadence metrics
	http.Handle("/metrics", promhttp.Handler())

	// Start server in goroutine
	go func() {
		if err := http.ListenAndServe(port, nil); err != nil {
			h.Logger.Error("Failed to start Prometheus server", zap.Error(err))
			os.Exit(1)
		}
	}()

	h.Logger.Info("Prometheus server started successfully", zap.String("port", port))
}
