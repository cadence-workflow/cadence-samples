## Timer Workflow

This sample demonstrates **timer usage** for timeouts and delayed notifications.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.TimerWorkflow \
  --input '5000000000'
```

The input is the processing threshold in nanoseconds (5 seconds = 5000000000).

### What Happens

```
┌──────────────────────────────────────────────────────────────┐
│                      TimerWorkflow                            │
│                                                               │
│  ┌─────────────────┐         ┌─────────────────┐             │
│  │ OrderProcessing │         │ Timer (5s)      │             │
│  │ (random 0-10s)  │         │                 │             │
│  └────────┬────────┘         └────────┬────────┘             │
│           │                           │                       │
│           ▼                           ▼                       │
│    If completes first:         If fires first:               │
│    Cancel timer                Send notification email       │
│                                Wait for processing           │
└──────────────────────────────────────────────────────────────┘
```

1. Starts a long-running `OrderProcessingActivity` (takes random 0-10 seconds)
2. Starts a timer for the threshold duration
3. **If processing finishes first**: Timer is cancelled
4. **If timer fires first**: Sends notification email, then waits for processing

### Key Concept: Timer with Cancellation

```go
childCtx, cancelHandler := workflow.WithCancel(ctx)

// Start processing
f := workflow.ExecuteActivity(ctx, orderProcessingActivity)
selector.AddFuture(f, func(f workflow.Future) {
    processingDone = true
    cancelHandler()  // Cancel the timer if processing completes
})

// Start timer
timerFuture := workflow.NewTimer(childCtx, threshold)
selector.AddFuture(timerFuture, func(f workflow.Future) {
    if !processingDone {
        workflow.ExecuteActivity(ctx, sendEmailActivity)
    }
})
```

### Real-World Use Cases

- Order processing with SLA monitoring
- Payment processing with timeout alerts
- API calls with fallback mechanisms

