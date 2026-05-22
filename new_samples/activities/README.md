<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Activities Sample

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

This folder contains samples demonstrating various activity patterns in Cadence.

### Dynamic Workflow

The `DynamicWorkflow` demonstrates executing an activity by its registered string name rather than passing the function reference directly. This pattern is useful when you need to dynamically determine which activity to execute at runtime.

```bash
cadence --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.DynamicWorkflow \
  --tl cadence-samples-worker \
  --et 60 \
  --input '{"message":"Cadence"}'
```

### Parallel Branch Pick First Workflow

The `ParallelBranchPickFirstWorkflow` demonstrates running multiple activities in parallel and returning the result of the first one to complete. This pattern is useful for scenarios like:
- Racing multiple data sources
- Implementing timeouts with fallbacks
- Redundant execution for reliability

```bash
cadence --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.ParallelBranchPickFirstWorkflow \
  --tl cadence-samples-worker \
  --et 60 \
  --input '{}'
```

The workflow will:
1. Start two parallel activities with different delays
2. Wait for the first one to complete
3. Cancel the remaining activity
4. Return the first successful result

Note: The `WaitForCancellation` option is set to `true` to demonstrate proper cleanup of cancelled activities. In production, you may set this to `false` if you don't need to wait for cancellation acknowledgment.

## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

