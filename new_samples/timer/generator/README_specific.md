## Timer Workflow Sample

This sample demonstrates **timers** for timeouts and delayed notifications.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 120 \
  --workflow_type cadence_samples.TimerWorkflow \
  --input '5000000000'
```

### Key Concept: Timer with Selector

```go
childCtx, cancelHandler := workflow.WithCancel(ctx)
timerFuture := workflow.NewTimer(childCtx, threshold)

selector := workflow.NewSelector(ctx)
selector.AddFuture(activityFuture, func(f workflow.Future) {
    cancelHandler() // Cancel timer if activity completes first
})
selector.AddFuture(timerFuture, func(f workflow.Future) {
    // Timer fired - send notification
})
selector.Select(ctx)
```

