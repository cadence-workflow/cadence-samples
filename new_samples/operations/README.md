<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Operations Sample

## Prerequisites

0. Install Cadence CLI. See instruction [here](https://cadenceworkflow.io/docs/cli/).
1. Run the Cadence server:
    1. Clone the [Cadence](https://github.com/cadence-workflow/cadence) repository if you haven't done already: `git clone https://github.com/cadence-workflow/cadence.git`
    2. Run `docker compose -f docker/docker-compose.yml up` to start Cadence server
    3. See more details at https://github.com/uber/cadence/blob/master/README.md
2. Once everything is up and running in Docker, open [localhost:8088](localhost:8088) to view Cadence UI.
3. Register the `cadence-samples` domain:

```bash
cadence --domain cadence-samples domain register
```

Refresh the [domains page](http://localhost:8088/domains) from step 2 to verify `cadence-samples` is registered.

## Steps to run sample

Inside the folder this sample is defined, run the following command:

```bash
go run .
```

This will call the main function in main.go which starts the worker, which will be execute the sample workflow code

## Samples in this folder

This folder contains samples demonstrating workflow operations and lifecycle management in Cadence.

### Cancel Workflow

The `CancelWorkflow` demonstrates how to properly handle workflow cancellation, including:
- Graceful cleanup when a workflow is cancelled
- Using a disconnected context to run cleanup activities after cancellation
- Heartbeating in long-running activities to detect cancellation

#### Start the workflow

```bash
cadence --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.CancelWorkflow \
  --tl cadence-samples-worker \
  --et 600 \
  --input '{}'
```

Copy the workflow ID from the output.

#### Cancel the workflow

```bash
cadence --domain cadence-samples \
  workflow cancel \
  --workflow_id <YOUR_WORKFLOW_ID>
```

#### What to observe

After cancellation:
1. The `ActivityToBeCanceled` will detect the cancellation via `ctx.Done()` and return
2. The `ActivityToBeSkipped` will not be scheduled (context already cancelled)
3. The `CleanupActivity` will run using a disconnected context to perform cleanup

This pattern is essential for workflows that need to release resources or perform cleanup operations when cancelled.

## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

