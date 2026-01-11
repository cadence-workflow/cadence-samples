# Cancel Activity Sample

This sample demonstrates **graceful activity cancellation** with cleanup operations.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for screenshots.

## How It Works

```
┌──────────────────┐
│ Start Workflow   │
└────────┬─────────┘
         ▼
┌──────────────────┐     ┌──────────────────┐
│ activityToBe     │────▶│ User cancels     │
│ Canceled()       │     │ workflow         │
│ (heartbeating)   │     └────────┬─────────┘
└──────────────────┘              │
         │◀───────────────────────┘
         ▼
┌──────────────────┐
│ cleanupActivity()│  (runs via NewDisconnectedContext)
└──────────────────┘
```

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Build the samples: `make`

## Running the Sample

```bash
# Terminal 1: Start worker
./bin/cancelactivity -m worker

# Terminal 2: Trigger workflow
./bin/cancelactivity -m trigger

# Terminal 3: Cancel the workflow (copy WorkflowID from trigger output)
./bin/cancelactivity -m cancel -w <WorkflowID>
```

## Key Code

```go
defer func() {
    if cadence.IsCanceledError(retError) {
        newCtx, _ := workflow.NewDisconnectedContext(ctx)
        workflow.ExecuteActivity(newCtx, cleanupActivity).Get(ctx, nil)
    }
}()
```

## Testing

```bash
go test -v ./cmd/samples/recipes/cancelactivity/
```

