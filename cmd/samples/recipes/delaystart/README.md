# Delay Start Sample

This sample demonstrates **delayed workflow execution** - starting a workflow that waits before executing its logic.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for screenshots.

## How It Works

```
┌──────────────────┐
│ Workflow Start   │
│ (with delay)     │
└────────┬─────────┘
         │
    ⏳ Wait 10s
         │
         ▼
┌──────────────────┐
│ delayStartActivity│
│ executes         │
└──────────────────┘
```

**Use case:** Scheduled tasks, delayed notifications, batch processing at specific times.

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Build the samples: `make`

## Running the Sample

```bash
# Terminal 1: Start worker
./bin/delaystart -m worker

# Terminal 2: Trigger workflow (will execute after delay)
./bin/delaystart -m trigger
```

## Key Code

```go
func delayStartWorkflow(ctx workflow.Context, delayStart time.Duration) error {
    logger.Info("workflow started after waiting for " + delayStart.String())
    workflow.ExecuteActivity(ctx, delayStartActivity, delayStart).Get(ctx, &result)
}
```

## Testing

```bash
go test -v ./cmd/samples/recipes/delaystart/
```

