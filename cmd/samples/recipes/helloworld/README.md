# Hello World Sample

This is the foundational Cadence workflow sample demonstrating basic workflow and activity execution.

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Build the samples:
   ```bash
   make
   ```

## Running the Sample

### Step 1: Start the Worker

```bash
./bin/helloworld -m worker
```

The worker registers the workflow and activity, then polls for tasks. You should see:
```
Started Worker. {"worker": "helloWorldGroup"}
```

### Step 2: Trigger the Workflow

In a new terminal:
```bash
./bin/helloworld -m trigger
```

This starts a new workflow execution. You should see:
```
Started Workflow {"WorkflowID": "helloworld_<uuid>", "RunID": "<uuid>"}
```

### Step 3: View the Result

Check the worker terminal - you'll see the workflow complete:
```
helloworld workflow started
helloworld activity started
Workflow completed. {"Result": "Hello Cadence!"}
```

Or view in the Cadence Web UI at [localhost:8088](http://localhost:8088).

## Worker Modes

This sample supports three modes via the `-m` flag:

### `worker` - Normal execution
```bash
./bin/helloworld -m worker
```
Starts a worker that processes workflow and activity tasks.

### `trigger` - Start workflow
```bash
./bin/helloworld -m trigger
```
Starts a new workflow execution (requires a running worker).

### `shadower` - Shadow testing
```bash
./bin/helloworld -m shadower
```
Replays completed workflows to verify code changes are deterministic. Useful for validating workflow modifications before deployment.

## Testing

### Unit Tests
```bash
go test -v ./cmd/samples/recipes/helloworld/ -run Test_Workflow
```

### Replay Tests
```bash
go test -v ./cmd/samples/recipes/helloworld/ -run TestReplayWorkflowHistoryFromFile
```

Replay tests use `helloworld.json` (a recorded workflow history) to verify that workflow code produces the same decisions. This catches non-deterministic changes.

### Shadow Tests
```bash
go test -v ./cmd/samples/recipes/helloworld/ -run TestWorkflowShadowing
```

Shadow tests replay recent production workflows against your local code to detect breaking changes.

## Understanding the Code

### Workflow Definition (`helloworld_workflow.go`)

```go
func helloWorldWorkflow(ctx workflow.Context, name string) error {
    // Configure activity options
    ao := workflow.ActivityOptions{
        ScheduleToStartTimeout: time.Minute,
        StartToCloseTimeout:    time.Minute,
        HeartbeatTimeout:       time.Second * 20,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)
    
    // Execute the activity
    var result string
    err := workflow.ExecuteActivity(ctx, helloWorldActivity, name).Get(ctx, &result)
    // ...
}
```

### Non-Determinism Warning

The workflow code contains commented-out code demonstrating a **non-deterministic change**:

```go
// Un-commenting the following code will cause replay tests to fail
// because it changes the workflow's decision history
// err := workflow.ExecuteActivity(ctx, helloWorldActivity, name).Get(ctx, &helloworldResult)
```

This teaches an important lesson: adding/removing activities changes the workflow's behavior and breaks running workflows. Use [workflow versioning](https://cadenceworkflow.io/docs/go-client/workflow-versioning/) for safe changes.

## Configuration

This sample uses `config/development.yaml` for connection settings:
- Domain name
- Cadence server host/port
- Metrics configuration

## References

- [Cadence Documentation](https://cadenceworkflow.io)
- [Workflow Versioning](https://cadenceworkflow.io/docs/go-client/workflow-versioning/)
- [Testing Workflows](https://cadenceworkflow.io/docs/go-client/workflow-testing/)

