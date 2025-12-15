<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Child Workflow Sample

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

## Child Workflow Sample

This sample demonstrates **parent-child workflow relationships** and the **ContinueAsNew** pattern.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.ParentWorkflow
```

### What Happens

```
┌─────────────────────┐
│   ParentWorkflow    │
└──────────┬──────────┘
           │
           │ ExecuteChildWorkflow
           ▼
┌─────────────────────┐
│   ChildWorkflow     │──┐
│   (run 1 of 5)      │  │
└─────────────────────┘  │
           │             │ ContinueAsNew
           ▼             │
┌─────────────────────┐  │
│   ChildWorkflow     │──┤
│   (run 2 of 5)      │  │
└─────────────────────┘  │
           │             │
          ...           ...
           │             │
           ▼             │
┌─────────────────────┐  │
│   ChildWorkflow     │◀─┘
│   (run 5 of 5)      │
└─────────────────────┘
           │
           │ Returns result
           ▼
┌─────────────────────┐
│   ParentWorkflow    │
│   completes         │
└─────────────────────┘
```

1. Parent workflow starts a child workflow
2. Child workflow uses `ContinueAsNew` to restart itself 5 times
3. After 5 runs, child completes and returns result to parent

### Key Concept: Child Workflow Options

```go
cwo := workflow.ChildWorkflowOptions{
    WorkflowID:                   childID,
    ExecutionStartToCloseTimeout: time.Minute,
}
ctx = workflow.WithChildOptions(ctx, cwo)

err := workflow.ExecuteChildWorkflow(ctx, ChildWorkflow, args...).Get(ctx, &result)
```

### Key Concept: ContinueAsNew

```go
// Instead of recursion (which grows history), use ContinueAsNew
return "", workflow.NewContinueAsNewError(ctx, ChildWorkflow, newArgs...)
```

ContinueAsNew starts a new workflow run with fresh history, avoiding unbounded history growth.


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

