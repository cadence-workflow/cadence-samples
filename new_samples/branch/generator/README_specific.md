## Branch Workflow

This sample demonstrates **parallel activity execution** - running multiple activities concurrently.

### Start Branch Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.BranchWorkflow
```

### Start Parallel Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.ParallelWorkflow
```

### What Happens

**BranchWorkflow** - Executes activities in parallel and waits for all:

```
         ┌─────────────────┐
         │ BranchWorkflow  │
         └────────┬────────┘
                  │
    ┌─────────────┼─────────────┐
    ▼             ▼             ▼
┌───────┐    ┌───────┐    ┌───────┐
│Branch1│    │Branch2│    │Branch3│
└───┬───┘    └───┬───┘    └───┬───┘
    │            │            │
    └─────────────┼───────────┘
                  ▼
         Wait for all to complete
```

**ParallelWorkflow** - Uses `workflow.Go()` for coroutines:

```
         ┌──────────────────┐
         │ ParallelWorkflow │
         └────────┬─────────┘
                  │
       ┌──────────┴──────────┐
       ▼                     ▼
  workflow.Go()         workflow.Go()
       │                     │
  ┌────┴────┐            ┌───┴───┐
  ▼         ▼            ▼       
branch1.1  branch1.2   branch2   
  │         │            │       
  └────┬────┘            │       
       └────────┬────────┘
                ▼
        Wait for both coroutines
```

### Key Concept: Parallel with Futures

```go
var futures []workflow.Future
for i := 1; i <= totalBranches; i++ {
    future := workflow.ExecuteActivity(ctx, BranchActivity, input)
    futures = append(futures, future)
}
// Wait for all
for _, future := range futures {
    future.Get(ctx, nil)
}
```

### Key Concept: Parallel with workflow.Go()

```go
waitChannel := workflow.NewChannel(ctx)

workflow.Go(ctx, func(ctx workflow.Context) {
    // Run activities sequentially in this branch
    workflow.ExecuteActivity(ctx, activity1).Get(ctx, nil)
    workflow.ExecuteActivity(ctx, activity2).Get(ctx, nil)
    waitChannel.Send(ctx, "done")
})

workflow.Go(ctx, func(ctx workflow.Context) {
    // Run in parallel
    workflow.ExecuteActivity(ctx, activity3).Get(ctx, nil)
    waitChannel.Send(ctx, "done")
})

// Wait for both coroutines
for i := 0; i < 2; i++ {
    waitChannel.Receive(ctx, nil)
}
```

