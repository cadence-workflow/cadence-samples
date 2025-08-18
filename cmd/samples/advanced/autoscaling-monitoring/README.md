# Autoscaling Monitoring Sample

This sample demonstrates three advanced Cadence worker features:

1. **Worker Poller Autoscaling** - Dynamic adjustment of worker poller goroutines based on workload
2. **Prometheus Tally Reporter** - Metrics collection using Tally and Prometheus
3. **HTTP Endpoint for Prometheus Scraping** - Exposing metrics for monitoring

## Features

### Worker Poller Autoscaling
The worker uses `worker.NewV2` with `AutoScalerOptions` to enable true autoscaling behavior:
- **AutoScalerOptions.Enabled**: true - Enables the autoscaling feature
- **PollerMinCount**: 2 - Minimum number of poller goroutines
- **PollerMaxCount**: 8 - Maximum number of poller goroutines  
- **PollerInitCount**: 4 - Initial number of poller goroutines

The worker automatically adjusts the number of poller goroutines between the min and max values based on the current workload.

### Prometheus Metrics
The sample exposes comprehensive metrics about worker performance, activity execution, and autoscaling behavior:
- `autoscaling_workflows_started_total` - Counter of workflows started
- `autoscaling_activities_completed_total` - Counter of activities completed
- `autoscaling_activity_duration_seconds` - Histogram of activity execution times
- `autoscaling_worker_load` - Gauge showing current worker load

### Monitoring Dashboards
When running the Cadence server locally with Grafana, you can access the client dashboards at:

**Client Dashboards**: http://localhost:3000/d/dehkspwgabvuoc/cadence-client

## Prerequisites

1. **Cadence Server**: Running locally with Docker Compose
2. **Prometheus**: Configured to scrape metrics from the sample
3. **Grafana**: With Cadence dashboards (included with default Cadence server setup)

## Quick Start

### 1. Start the Worker
```bash
./bin/autoscaling-monitoring -m worker
```

### 2. Start the Prometheus Server (Optional)
```bash
./bin/autoscaling-monitoring -m server
```

### 3. Generate Load
```bash
./bin/autoscaling-monitoring -m trigger
```

## Configuration

The sample uses a custom configuration system that extends the base Cadence configuration. You can specify a configuration file using the `-config` flag:

```bash
./bin/autoscaling-monitoring -m worker -config /path/to/config.yaml
```

### Configuration File Structure

```yaml
# Cadence connection settings
domain: "default"
service: "cadence-frontend"
host: "localhost:7833"

# Prometheus configuration
prometheus:
  listenAddress: "127.0.0.1:8004"

# Autoscaling configuration
autoscaling:
  # Worker autoscaling settings
  pollerMinCount: 2
  pollerMaxCount: 8
  pollerInitCount: 4
  
  # Load generation settings
  loadGeneration:
    iterations: 50          # Number of activities to execute
    batchDelay: 2           # Delay between batches (seconds)
    minProcessingTime: 1000 # Min activity time (ms)
    maxProcessingTime: 6000 # Max activity time (ms)
```

### Configuration Usage

The configuration values are used throughout the sample:

1. **Worker Configuration** (`worker_config.go`):
   - `pollerMinCount`, `pollerMaxCount`, `pollerInitCount` → `AutoScalerOptions`

2. **Workflow Configuration** (`workflow.go`):
   - `iterations` → Number of activities to execute
   - `batchDelay` → Delay between activity batches

3. **Activity Configuration** (`activities.go`):
   - `minProcessingTime`, `maxProcessingTime` → Activity processing time range

4. **Prometheus Configuration** (`prometheus_server.go`):
   - `listenAddress` → HTTP server port

### Default Configuration

If no configuration file is provided or if the file cannot be read, the sample uses these defaults:

```yaml
domain: "default"
service: "cadence-frontend"
host: "localhost:7833"
autoscaling:
  pollerMinCount: 2
  pollerMaxCount: 8
  pollerInitCount: 4
  loadGeneration:
    iterations: 50
    batchDelay: 2
    minProcessingTime: 1000
    maxProcessingTime: 6000
```

## Monitoring

### Metrics Endpoints
- **Prometheus Metrics**: http://localhost:8004/metrics
- **Health Check**: http://localhost:8004/health
- **Status Page**: http://localhost:8004/

### Grafana Dashboard
Access the Cadence client dashboard at: http://localhost:3000/d/dehkspwgabvuoc/cadence-client

### Key Metrics to Monitor

1. **Worker Performance**:
   - Activity execution rate
   - Decision task processing rate
   - Worker load levels

2. **Autoscaling Behavior**:
   - Concurrent activity execution count
   - Task queue depth
   - Poller utilization
   - Number of active poller goroutines

3. **Custom Metrics**:
   - `autoscaling_workflows_started_total`
   - `autoscaling_activities_completed_total`
   - `autoscaling_activity_duration_seconds`

## How It Works

### Load Generation
The sample creates a workflow that executes activities in parallel, with each activity:
- Taking 1-6 seconds to complete (configurable via `minProcessingTime`/`maxProcessingTime`)
- Recording metrics about execution time
- Creating varying load patterns with configurable batch delays

### Autoscaling Demonstration
The worker uses `worker.NewV2` with `AutoScalerOptions` to:
- Start with configurable poller goroutines (`pollerInitCount`)
- Scale down to minimum pollers (`pollerMinCount`) when load is low
- Scale up to maximum pollers (`pollerMaxCount`) when load is high
- Automatically adjust based on task queue depth and processing time

### Metrics Collection
The sample collects and exposes:
- Workflow start/complete events
- Activity execution times and counts
- Worker load simulation metrics
- Standard Cadence metrics via Tally

## Production Considerations

### Scaling
- Adjust `pollerMinCount`, `pollerMaxCount`, and `pollerInitCount` based on your workload
- Monitor worker performance and adjust autoscaling parameters
- Use multiple worker instances for high availability

### Monitoring
- Configure Prometheus to scrape metrics regularly
- Set up alerts for worker performance issues
- Use Grafana dashboards to visualize autoscaling behavior
- Monitor poller count changes to verify autoscaling is working

### Security
- Secure the Prometheus endpoint in production
- Use authentication for metrics access
- Consider using HTTPS for metrics endpoints

## Troubleshooting

### Common Issues

1. **Worker Not Starting**:
   - Check Cadence server is running
   - Verify domain exists
   - Check configuration file
   - Ensure using compatible Cadence client version

2. **Autoscaling Not Working**:
   - Verify `worker.NewV2` is being used
   - Check `AutoScalerOptions.Enabled` is true
   - Monitor poller count changes in logs
   - Ensure sufficient load is being generated

3. **Configuration Issues**:
   - Verify configuration file path is correct
   - Check YAML syntax in configuration file
   - Review default values if config file is not found

4. **Metrics Not Appearing**:
   - Verify Prometheus server is running
   - Check metrics endpoint is accessible
   - Ensure Prometheus is configured to scrape the endpoint

5. **Dashboard Not Loading**:
   - Verify Grafana is running
   - Check dashboard URL is correct
   - Ensure Prometheus data source is configured

### Debug Mode
Enable debug logging by setting the log level in your environment:
```bash
export CADENCE_LOG_LEVEL=debug
```
