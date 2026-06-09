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

