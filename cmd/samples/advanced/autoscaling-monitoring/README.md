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

> **Note**: Make sure to select a Domain in Grafana for the dashboards to display data. The dashboards will be empty until a domain is selected from the dropdown.


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

### 2. Generate Load
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
    # Workflow-level settings
    workflows: 10             # Number of workflows to start
    workflowDelay: 1000       # Delay between starting workflows (milliseconds)
    
    # Activity-level settings (per workflow)
    activitiesPerWorkflow: 30 # Number of activities per workflow
    batchDelay: 2000          # Delay between activity batches within workflow (milliseconds)
    
    # Activity processing time range (milliseconds)
    minProcessingTime: 1000
    maxProcessingTime: 6000
```

### Configuration Usage

The configuration values are used throughout the sample:

1. **Worker Configuration** (`worker_config.go`):
   - `pollerMinCount`, `pollerMaxCount`, `pollerInitCount` → `AutoScalerOptions`

2. **Workflow Configuration** (`workflow.go`):
   - `activitiesPerWorkflow` → Number of activities to execute per workflow
   - `batchDelay` → Delay between activity batches within workflow

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
    workflows: 10
    workflowDelay: 1000
    activitiesPerWorkflow: 30
    batchDelay: 2000
    minProcessingTime: 1000
    maxProcessingTime: 6000
```

### Load Pattern Examples

The sample supports various load patterns for testing autoscaling behavior:

#### **1. Gradual Ramp-up (Default)**
```yaml
loadGeneration:
  workflows: 10
  workflowDelay: 1000
  activitiesPerWorkflow: 30
```
**Result**: 10 workflows starting 1 second apart, each with 30 activities (300 total activities)

#### **2. Burst Load**
```yaml
loadGeneration:
  workflows: 25
  workflowDelay: 0
  activitiesPerWorkflow: 60
```
**Result**: 25 workflows all starting immediately (1500 total activities)

#### **3. Sustained Load**
```yaml
loadGeneration:
  workflows: 50
  workflowDelay: 2000
  activitiesPerWorkflow: 100
```
**Result**: 5 long-running workflows with 2-second delays between starts (5000 total activities)

#### **4. Light Load**
```yaml
loadGeneration:
  workflows: 1
  workflowDelay: 0
  activitiesPerWorkflow: 20
```
**Result**: Single workflow with 20 activities for minimal load testing

## Monitoring

### Metrics Endpoints
- **Prometheus Metrics**: http://127.0.0.1:8004/metrics
  - Exposed automatically when running worker mode only
  - Real-time autoscaling and worker performance metrics
  - Prometheus-compatible format with sanitized names
  - **Note**: Metrics server is not started in trigger mode

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
The sample creates multiple workflows that execute activities in parallel, with each workflow:
- Starting with configurable delays (`workflowDelay`) to create sustained load patterns
- Executing a configurable number of activities (`activitiesPerWorkflow`) per workflow
- Each activity taking 1-6 seconds to complete (configurable via `minProcessingTime`/`maxProcessingTime`)
- Recording metrics about execution time
- Creating varying load patterns with configurable batch delays within each workflow

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

## Testing

The sample includes unit tests for the configuration loading functionality. Run these tests if you make any changes to the config:

### Running Tests
```bash
# Run all tests
go test -v

# Run specific test
go test -v -run TestLoadConfiguration_SuccessfulLoading

# Run tests with coverage
go test -v -cover
```

### Test Coverage
The tests cover:
- **Successful configuration loading** - Complete YAML files with all fields
- **Missing file fallback** - Graceful handling when config file doesn't exist
- **Default value application** - Ensuring all fields have sensible defaults

### Configuration Testing
The tests validate that the improved configuration system:
- Handles embedded struct issues properly
- Applies defaults correctly for missing fields
- Provides clear error messages for configuration problems
- Maintains backward compatibility

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
