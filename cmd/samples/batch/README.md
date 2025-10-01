# Batch Processing Sample

This sample demonstrates how to process large batches of activities with controlled concurrency using Cadence's `workflow.NewBatchFuture` functionality, while respecting the 1024 pending activities limit per workflow.

## The problem it solves

**The Problem**: When processing large datasets (thousands of records, files, or API calls), you face a dilemma:
- **Sequential processing**: Too slow, poor user experience
- **Unlimited concurrency**: Overwhelms databases, APIs, or downstream services
- **Manual concurrency control**: Complex error handling and resource management
- **Cadence limits**: Max 1024 pending activities per workflow

**Real-world scenarios**:
- Processing 10,000 user records for a migration
- Sending emails to 50,000 subscribers
- Generating reports for 1,000 customers
- Processing files in a data pipeline

### The Solution
`workflow.NewBatchFuture` provides a robust solution:

**Controlled Concurrency**: Process items in parallel while respecting system limits
**Automatic Error Handling**: Failed activities don't crash the entire batch
**Resource Efficiency**: Optimal throughput without overwhelming downstream services
**Built-in Observability**: Monitoring, retries, and failure tracking
**Workflow Integration**: Seamless integration with Cadence's workflow engine

This eliminates the need to build custom concurrency control, error handling, and monitoring systems.

## Sample behavior

- Creates a configurable number of activities (default: 10)
- Executes them with controlled concurrency (default: 3)
- Simulates work with random delays (900-999ms per activity)
- Handles cancellation gracefully

## Technical considerations

- **Cadence limit**: Maximum 1024 pending activities per workflow
- **Resource management**: Controlled concurrency prevents system overload
- **Error handling**: Failed activities don't crash the entire batch

## How to run

1. Build the sample:
```bash
make batch
```

2. Start Worker:
```bash
./bin/batch -m worker
```

3. Start Workflow:
```bash
./bin/batch -m trigger
```
