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

