# Pick First Sample

This sample demonstrates **racing activities** - running multiple activities in parallel and using the first result.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for screenshots.

## How It Works

```
     ┌─────────────────┐
     │  Start          │
     └────────┬────────┘
        ┌─────┴─────┐
        ▼           ▼
   ┌─────────┐ ┌─────────┐
   │ Task 1  │ │ Task 2  │   (parallel)
   │ 2 sec   │ │ 10 sec  │
   └────┬────┘ └────┬────┘
        │           │
   First ✓     Cancel ✗
        │
        ▼
   ┌─────────┐
   │ Result  │
   └─────────┘
```

**Use case:** Failover, load balancing, fastest-response-wins patterns.

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Build the samples: `make`

## Running the Sample

```bash
# Terminal 1: Start worker
./bin/pickfirst -m worker

# Terminal 2: Trigger workflow
./bin/pickfirst -m trigger
```

## Key Code

```go
f1 := workflow.ExecuteActivity(childCtx, sampleActivity, 0, time.Second*2)
f2 := workflow.ExecuteActivity(childCtx, sampleActivity, 1, time.Second*10)

selector.AddFuture(f1, handler).AddFuture(f2, handler)
selector.Select(ctx)  // Wait for first
cancelHandler()       // Cancel others
```

## Testing

```bash
go test -v ./cmd/samples/recipes/pickfirst/
```

