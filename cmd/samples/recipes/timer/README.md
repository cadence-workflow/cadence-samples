# Timer Sample

This sample demonstrates **timer-based notifications** - running a long operation with a timeout that triggers an alert if processing takes too long.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for a step-by-step tutorial with screenshots.

## How It Works

```
┌──────────────────────────────────────────────────────────────┐
│                      Workflow Start                          │
└──────────────────────────────────────────────────────────────┘
                              │
              ┌───────────────┴───────────────┐
              ▼                               ▼
   ┌─────────────────────┐         ┌─────────────────────┐
   │ orderProcessing()   │         │ Timer (3 seconds)   │
   │ (random 0-10 sec)   │         │                     │
   └─────────────────────┘         └─────────────────────┘
              │                               │
              │   If timer fires first:       │
              │   ◀─────────────────────────────
              │         sendEmail()           │
              │                               │
              ▼                               │
   ┌─────────────────────┐                    │
   │ Wait for processing │◀───────────────────┘
   │ to complete         │
   └─────────────────────┘
```

**Use case:** Order processing with SLA monitoring - notify customer if processing is delayed, but don't cancel the operation.

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Build the samples: `make`

## Running the Sample

```bash
# Terminal 1: Start worker
./bin/timer -m worker

# Terminal 2: Trigger workflow
./bin/timer -m trigger
```

**Possible outcomes:**
- Processing finishes in < 3 seconds → Timer cancelled, no email
- Processing takes > 3 seconds → Email sent, then wait for completion

## Key Code

```go
// Start long-running activity
f := workflow.ExecuteActivity(ctx, orderProcessingActivity)
selector.AddFuture(f, func(f workflow.Future) {
    processingDone = true
    cancelHandler()  // Cancel timer if processing finishes first
})

// Start timer for notification threshold
timerFuture := workflow.NewTimer(childCtx, processingTimeThreshold)
selector.AddFuture(timerFuture, func(f workflow.Future) {
    if !processingDone {
        workflow.ExecuteActivity(ctx, sendEmailActivity)  // Send notification
    }
})

selector.Select(ctx)  // Wait for either to complete
```

## Testing

```bash
go test -v ./cmd/samples/recipes/timer/
```

## References

- [Cadence Timers](https://cadenceworkflow.io/docs/go-client/timers/)
- [Workflow Selectors](https://cadenceworkflow.io/docs/go-client/selectors/)

