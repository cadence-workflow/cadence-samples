<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Pick First Sample

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


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

