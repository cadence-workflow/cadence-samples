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

