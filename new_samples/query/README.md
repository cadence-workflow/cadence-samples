<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Query Sample

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

## Query Workflow Sample

This sample demonstrates **workflow queries** - inspecting workflow state without affecting execution.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 180 \
  --workflow_type cadence_samples.QueryWorkflow
```

### Query the Workflow

While the workflow is running, query its state:

```bash
cadence --env development \
  --domain cadence-samples \
  workflow query \
  --wid <workflow_id> \
  --qt state
```

### What Happens

The workflow goes through states that you can query:

```
Time 0:   state = "started"
Time 1s:  state = "waiting on timer"
Time 2m:  state = "done" (workflow completes)
```

### Key Concept: Query Handler

```go
func QueryWorkflow(ctx workflow.Context) error {
    currentState := "started"
    
    // Register query handler for "state" query type
    workflow.SetQueryHandler(ctx, "state", func() (string, error) {
        return currentState, nil
    })
    
    currentState = "waiting on timer"
    workflow.NewTimer(ctx, 2*time.Minute).Get(ctx, nil)
    
    currentState = "done"
    return nil
}
```

### Use Cases

- Progress monitoring dashboards
- Debugging running workflows
- Health checks without affecting execution


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

