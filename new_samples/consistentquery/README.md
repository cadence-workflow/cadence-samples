<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Consistent Query Sample

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

## Consistent Query Sample

This sample demonstrates **consistent queries** with signal handling.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 180 \
  --workflow_type cadence_samples.ConsistentQueryWorkflow
```

### Query the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow query \
  --wid <workflow_id> \
  --qt state
```

### Send Signals to Update State

```bash
cadence --env development \
  --domain cadence-samples \
  workflow signal \
  --wid <workflow_id> \
  --name increase
```

Each signal increments the counter. Query to see the updated value.

### Key Concept: Query + Signal

```go
queryResult := 0

// Register query handler
workflow.SetQueryHandler(ctx, "state", func() (int, error) {
    return queryResult, nil
})

// Handle signals that modify state
signalChan := workflow.GetSignalChannel(ctx, "increase")
signalChan.Receive(ctx, nil)
queryResult += 1  // State changes are visible to queries
```


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

