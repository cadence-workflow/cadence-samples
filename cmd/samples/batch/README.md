# Batch Processing Sample

This sample demonstrates how to process large batches of tasks with controlled concurrency using Cadence's `x.NewBatchFuture` functionality.

## What it does

- Creates a configurable number of activities (default: 10)
- Executes them with controlled concurrency (default: 3)
- Simulates work with random delays (900-999ms per task)
- Handles cancellation gracefully

## Real-world use cases

- Batch data processing
- Bulk operations
- ETL jobs
- Report generation
- File processing

## How to run

Start Worker:
```bash
./bin/batch -m worker
```

Start Workflow:
```bash
./bin/batch -m trigger
```

## Key concepts

- **Batch processing**: Process multiple tasks efficiently
- **Concurrency control**: Limit simultaneous executions
- **Activity factories**: Lazy evaluation of activities
- **Future-based execution**: Asynchronous task management
- **Context cancellation**: Graceful handling of timeouts