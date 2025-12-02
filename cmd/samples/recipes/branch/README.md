# Branch Sample

This sample demonstrates **parallel execution** - running multiple activities concurrently using two different approaches.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for a step-by-step tutorial with screenshots.

## Two Approaches to Parallelism

### 1. Futures (`-c branch`)
Execute activities in parallel, collect futures, then wait for all:

```
     ┌─────────────┐
     │   Start     │
     └──────┬──────┘
   ┌────────┼────────┐
   ▼        ▼        ▼
┌─────┐  ┌─────┐  ┌─────┐
│ A1  │  │ A2  │  │ A3  │   (parallel)
└──┬──┘  └──┬──┘  └──┬──┘
   └────────┼────────┘
            ▼
     ┌─────────────┐
     │ Wait all    │
     └─────────────┘
```

### 2. Coroutines (`workflow.Go`)
Spawn goroutine-like coroutines with channels for coordination:

```
     ┌─────────────┐
     │   Start     │
     └──────┬──────┘
        ┌───┴───┐
        ▼       ▼
   ┌─────────┐ ┌─────┐
   │ A1 → A2 │ │ A3  │   (parallel branches)
   └────┬────┘ └──┬──┘
        └────┬────┘
             ▼
     ┌─────────────┐
     │ channel.Receive │
     └─────────────┘
```

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Build the samples: `make`

## Running the Sample

```bash
# Terminal 1: Start worker
./bin/branch -m worker

# Terminal 2: Trigger workflows
./bin/branch -m trigger              # Coroutines approach (default)
./bin/branch -m trigger -c branch    # Futures approach
```

## Key Code

### Futures Pattern
```go
var futures []workflow.Future
for i := 1; i <= totalBranches; i++ {
    future := workflow.ExecuteActivity(ctx, sampleActivity, input)
    futures = append(futures, future)
}
// Wait for all
for _, future := range futures {
    future.Get(ctx, nil)
}
```

### Coroutines Pattern
```go
waitChannel := workflow.NewChannel(ctx)

workflow.Go(ctx, func(ctx workflow.Context) {
    workflow.ExecuteActivity(ctx, activity1).Get(ctx, nil)
    workflow.ExecuteActivity(ctx, activity2).Get(ctx, nil)
    waitChannel.Send(ctx, "done")
})

workflow.Go(ctx, func(ctx workflow.Context) {
    workflow.ExecuteActivity(ctx, activity3).Get(ctx, nil)
    waitChannel.Send(ctx, "done")
})

// Wait for both coroutines
waitChannel.Receive(ctx, nil)
waitChannel.Receive(ctx, nil)
```

## Testing

```bash
go test -v ./cmd/samples/recipes/branch/
```

## References

- [Parallel Execution](https://cadenceworkflow.io/docs/go-client/execute-activity/)
- [Workflow Coroutines](https://cadenceworkflow.io/docs/go-client/workflow-patterns/)

