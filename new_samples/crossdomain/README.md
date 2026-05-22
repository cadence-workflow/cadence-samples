<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Cross Domain Sample

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

## Cross Domain Sample

This sample demonstrates executing **child workflows across different Cadence domains**.

### Prerequisites

Register a second domain for the child workflow:

```bash
cadence --env development --domain child-domain domain register
```

Start a worker in the child domain (separate terminal):

```bash
# Worker for child-domain would need to be configured separately
```

### Start the Parent Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 120 \
  --workflow_type cadence_samples.CrossDomainWorkflow
```

### Key Concept: Cross-Domain Child Options

```go
childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
    Domain:                       "other-domain",  // Different domain!
    WorkflowID:                   "child-wf-123",
    TaskList:                     "other-task-list",
    ExecutionStartToCloseTimeout: time.Minute,
})

workflow.ExecuteChildWorkflow(childCtx, ChildWorkflow, args...)
```

### Use Cases

- Multi-tenant architectures
- Isolation between teams/services
- Cross-cluster workflow execution


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

