<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Side Effect Sample

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

## Side Effect Sample

This sample demonstrates **workflow.SideEffect** for handling non-deterministic operations.

### The Problem

Workflows must be deterministic for replay. But sometimes you need non-deterministic values like:
- UUIDs
- Random numbers
- Current time
- External state

### The Solution: SideEffect

```go
workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
    return uuid.New().String()  // Non-deterministic!
}).Get(&value)
```

On first execution, SideEffect runs the function and records the result.
On replay, it returns the recorded value without re-executing.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.SideEffectWorkflow
```

### Query the Generated Value

```bash
cadence --env development \
  --domain cadence-samples \
  workflow query \
  --wid <workflow_id> \
  --qt value
```

The same UUID is returned every time you query, demonstrating deterministic replay.


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

