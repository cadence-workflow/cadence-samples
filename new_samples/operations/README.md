<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Cancel Workflow Sample

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

## Cancel Workflow Sample

This sample demonstrates how to cancel a running workflow and perform cleanup operations.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.CancelWorkflow
```

The workflow will start an activity that heartbeats every second.

### Cancel the Workflow

In another terminal, cancel the workflow:

```bash
cadence --env development \
  --domain cadence-samples \
  workflow cancel \
  --wid <workflow_id>
```

### What Happens

1. `ActivityToBeCanceled` starts and heartbeats every second
2. When you cancel the workflow, the activity receives a cancellation signal
3. The workflow runs `CleanupActivity` in a disconnected context
4. `ActivityToBeSkipped` is never executed (skipped due to cancellation)

### Key Concept: WaitForCancellation

```go
ao := workflow.ActivityOptions{
    WaitForCancellation: true,  // Wait for activity to acknowledge cancellation
}
```

When `WaitForCancellation` is true, Cadence waits for the activity to handle the cancellation before proceeding.

### Key Concept: Disconnected Context

```go
// When workflow is canceled, get a new disconnected context for cleanup
newCtx, _ := workflow.NewDisconnectedContext(ctx)
err := workflow.ExecuteActivity(newCtx, cleanupActivity).Get(ctx, nil)
```

A disconnected context allows cleanup activities to run even after the workflow is canceled.


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

