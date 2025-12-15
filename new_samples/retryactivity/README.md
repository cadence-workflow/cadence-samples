<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Retry Activity Sample

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

## Retry Activity Sample

This sample demonstrates **retry policies** and **heartbeat progress tracking** for unreliable activities.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 600 \
  --workflow_type cadence_samples.RetryWorkflow
```

### What Happens

The activity processes 20 tasks but intentionally fails after every ~7 tasks (1/3 of batch). With retry policy, it resumes from the last heartbeated progress.

```
Attempt 1: Process tasks 0-6, fail, heartbeat progress=6
Attempt 2: Resume from 7, process 7-13, fail, heartbeat progress=13  
Attempt 3: Resume from 14, process 14-19, complete!
```

### Key Concept: Retry Policy

```go
ao := workflow.ActivityOptions{
    RetryPolicy: &cadence.RetryPolicy{
        InitialInterval:    time.Second,      // First retry after 1s
        BackoffCoefficient: 2.0,              // Double wait each retry
        MaximumInterval:    time.Minute,      // Cap at 1 minute
        MaximumAttempts:    5,                // Give up after 5 tries
        NonRetriableErrorReasons: []string{"bad-error"}, // Don't retry these
    },
}
```

### Key Concept: Heartbeat with Progress

```go
// Record progress during activity execution
activity.RecordHeartbeat(ctx, currentTaskID)

// On retry, resume from last heartbeated progress
if activity.HasHeartbeatDetails(ctx) {
    var lastCompletedID int
    activity.GetHeartbeatDetails(ctx, &lastCompletedID)
    startFrom = lastCompletedID + 1
}
```

This avoids reprocessing already-completed work after a failure.


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

