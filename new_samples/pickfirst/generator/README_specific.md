## Pick First Workflow

This sample demonstrates **race condition handling** - running multiple activities in parallel and using the result from whichever completes first.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.PickFirstWorkflow
```

### What Happens

```
         ┌──────────────────┐
         │ PickFirstWorkflow│
         └────────┬─────────┘
                  │
       ┌──────────┴──────────┐
       ▼                     ▼
┌─────────────┐       ┌─────────────┐
│ RaceActivity│       │ RaceActivity│
│ (2 seconds) │       │ (10 seconds)│
└──────┬──────┘       └──────┬──────┘
       │                     │
       ▼                     │
   Completes first!          │
       │                     │
       ▼                     ▼
   Use result            CANCELLED
```

1. Two activities start in parallel with different durations
2. The first one to complete "wins"
3. All other pending activities are cancelled
4. Workflow uses the winner's result

### Key Concept: Selector with Cancellation

```go
childCtx, cancelHandler := workflow.WithCancel(ctx)

// Start activities in parallel
f1 := workflow.ExecuteActivity(childCtx, RaceActivity, 0, 2*time.Second)
f2 := workflow.ExecuteActivity(childCtx, RaceActivity, 1, 10*time.Second)

selector := workflow.NewSelector(ctx)
selector.AddFuture(f1, func(f workflow.Future) {
    f.Get(ctx, &result)
})
selector.AddFuture(f2, func(f workflow.Future) {
    f.Get(ctx, &result)
})

// Wait for first to complete
selector.Select(ctx)

// Cancel all others
cancelHandler()
```

### Key Concept: Activity Cancellation Handling

```go
func RaceActivity(ctx context.Context, ...) (string, error) {
    for {
        activity.RecordHeartbeat(ctx, "status")
        
        select {
        case <-ctx.Done():
            // We've been cancelled
            return "cancelled", ctx.Err()
        default:
            // Continue working
        }
    }
}
```

### Real-World Use Cases

- Multi-provider API calls (use fastest response)
- Redundant service calls for reliability
- Load balancing with failover

