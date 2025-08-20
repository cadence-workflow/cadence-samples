# Autoscaling Monitoring Sample

This sample demonstrates three advanced Cadence worker features:

1. **Worker Poller Autoscaling** - Dynamic adjustment of worker poller goroutines based on workload
2. **Integrated Prometheus Metrics** - Real-time metrics collection using Tally with Prometheus reporter
3. **Autoscaling Metrics** - Comprehensive autoscaling behavior metrics exposed via HTTP endpoint

## Features

### Worker Poller Autoscaling
The worker uses `worker.NewV2` with `AutoScalerOptions` to enable true autoscaling behavior:
- **AutoScalerOptions.Enabled**: true - Enables the autoscaling feature
- **PollerMinCount**: 2 - Minimum number of poller goroutines
- **PollerMaxCount**: 8 - Maximum number of poller goroutines  
- **PollerInitCount**: 4 - Initial number of poller goroutines

The worker automatically adjusts the number of poller goroutines between the min and max values based on the current workload.

### Prometheus Metrics
The sample uses Tally with Prometheus reporter to expose comprehensive metrics:
- **Real-time autoscaling metrics** - Poller count changes, quota adjustments, wait times
- **Worker performance metrics** - Task processing rates, poller utilization, queue depths
- **Standard Cadence metrics** - All metrics automatically emitted by the Cadence Go client
- **Sanitized metric names** - Prometheus-compatible metric names and labels

### Monitoring Dashboards
When running the Cadence server locally with Grafana, you can access the client dashboards at:

**Client Dashboards**: http://localhost:3000/d/dehkspwgabvuoc/cadence-client

## Prerequisites

1. **Cadence Server**: Running locally with Docker Compose.
2. **Prometheus**: Configured to scrape metrics from the sample.
3. **Grafana**: With Cadence dashboards (included with default Cadence server setup). Dashboards in the latest version of the server.

## Quick Start

### 1. Start the Worker
```bash
./bin/autoscaling-monitoring -m worker
```

The worker automatically exposes metrics at: http://127.0.0.1:8004/metrics

### 2. Access Metrics (Optional)
For dedicated metrics server:
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

4. **Prometheus Configuration** (integrated):
   - `listenAddress` → Metrics endpoint port (default: 127.0.0.1:8004)

### Default Configuration

If no configuration file is provided or if the file cannot be read, the sample uses these defaults:

```yaml
domain: "default"
service: "cadence-frontend"
host: "localhost:7833"
prometheus:
  listenAddress: "127.0.0.1:8004"
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
- **Prometheus Metrics**: http://127.0.0.1:8004/metrics
  - Exposed automatically when running worker or server mode
  - Real-time autoscaling and worker performance metrics
  - Prometheus-compatible format with sanitized names

### Grafana Dashboard
Access the Cadence client dashboard at: http://localhost:3000/d/dehkspwgabvuoc/cadence-client

### Key Metrics to Monitor

1. **Worker Performance Metrics**:
   - `cadence_worker_decision_poll_success_count` - Successful decision task polls
   - `cadence_worker_activity_poll_success_count` - Successful activity task polls
   - `cadence_worker_decision_poll_count` - Total decision task poll attempts
   - `cadence_worker_activity_poll_count` - Total activity task poll attempts

2. **Autoscaling Behavior Metrics**:
   - `cadence_worker_poller_count` - Number of active poller goroutines (key autoscaling indicator)
   - `cadence_concurrency_auto_scaler_poller_quota` - Current poller quota for autoscaling
   - `cadence_concurrency_auto_scaler_poller_wait_time` - Time pollers wait for tasks
   - `cadence_concurrency_auto_scaler_scale_up_count` - Number of scale-up events
   - `cadence_concurrency_auto_scaler_scale_down_count` - Number of scale-down events

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
The sample uses Tally with Prometheus reporter for comprehensive metrics:
- **Real-time autoscaling metrics** - Poller count changes, quota adjustments, scale events
- **Worker performance metrics** - Task processing rates, poller utilization, queue depths
- **Standard Cadence metrics** - All metrics automatically emitted by the Cadence Go client
- **Sanitized metric names** - Prometheus-compatible format with proper character replacement

## Production Considerations

### Scaling
- Adjust `pollerMinCount`, `pollerMaxCount`, and `pollerInitCount` based on your workload
- Monitor worker performance and adjust autoscaling parameters
- Use multiple worker instances for high availability

### Monitoring
- Configure Prometheus to scrape metrics regularly (latest version of Cadence server is configured to do this)
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
   - Verify worker is running (metrics are exposed automatically)
   - Check metrics endpoint is accessible: http://127.0.0.1:8004/metrics
   - Ensure Prometheus is configured to scrape the endpoint
   - Check for metric name sanitization issues

5. **Dashboard Not Loading**:
   - Verify Grafana is running
   - Check dashboard URL is correct
   - Ensure Prometheus data source is configured
