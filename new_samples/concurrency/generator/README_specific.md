## Samples in this folder

This folder contains samples demonstrating concurrency control patterns in Cadence.

### Batch Processing Workflow

The `BatchWorkflow` demonstrates how to process large batches of activities with controlled concurrency using Cadence's `workflow.NewBatchFuture` functionality.

#### The Problem It Solves

When processing large datasets (thousands of records, files, or API calls), you face a dilemma:
- **Sequential processing**: Too slow, poor user experience
- **Unlimited concurrency**: Overwhelms databases, APIs, or downstream services
- **Manual concurrency control**: Complex error handling and resource management
- **Cadence limits**: Max 1024 pending activities per workflow

#### The Solution

`workflow.NewBatchFuture` provides a robust solution:

- **Controlled Concurrency**: Process items in parallel while respecting system limits
- **Automatic Error Handling**: Failed activities don't crash the entire batch
- **Resource Efficiency**: Optimal throughput without overwhelming downstream services
- **Built-in Observability**: Monitoring, retries, and failure tracking
- **Workflow Integration**: Seamless integration with Cadence's workflow engine

#### Real-World Scenarios

- Processing 10,000 user records for a migration
- Sending emails to 50,000 subscribers
- Generating reports for 1,000 customers
- Processing files in a data pipeline

#### Sample Behavior

- Creates a configurable number of activities (default: 10)
- Executes them with controlled concurrency (default: 3)
- Simulates work with random delays (900-999ms per activity)
- Handles cancellation gracefully

#### Technical Considerations

- **Cadence limit**: Maximum 1024 pending activities per workflow
- **Resource management**: Controlled concurrency prevents system overload
- **Error handling**: Failed activities don't crash the entire batch

#### How to Start the Workflow

```bash
cadence --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.BatchWorkflow \
  --tl cadence-samples-worker \
  --et 300 \
  --input '{"Concurrency":3,"TotalSize":10}'
```

You can adjust the parameters:
- `Concurrency`: Maximum number of activities running in parallel
- `TotalSize`: Total number of activities to process
