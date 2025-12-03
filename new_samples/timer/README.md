<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Timer Sample

## Prerequisites

0. Install Cadence CLI. See instruction [here](https://cadenceworkflow.io/docs/cli/).
1. Run the Cadence server:
    1. Clone the [Cadence](https://github.com/cadence-workflow/cadence) repository if you haven't done already: `git clone https://github.com/cadence-workflow/cadence.git`
    2. Run `docker compose -f docker/docker-compose.yml up` to start Cadence server
    3. See more details at https://github.com/uber/cadence/blob/master/README.md
2. Once everything is up and running in Docker, open [localhost:8088](localhost:8088) to view Cadence UI.
3. Register the `cadence-samples` domain:

```bash
cadence --env development --domain cadence-samples domain register
```

Refresh the [domains page](http://localhost:8088/domains) from step 2 to verify `cadence-samples` is registered.

## Steps to run sample

Inside the folder this sample is defined, run the following command:

```bash
go run .
```

This will call the main function in main.go which starts the worker, which will be execute the sample workflow code

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


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

