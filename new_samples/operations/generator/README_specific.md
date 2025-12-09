## Cancel Workflow Sample

This sample demonstrates how to cancel a running workflow and perform cleanup operations.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.CancelWorkflow
```

The workflow will start an activity that heartbeats every second.

### Cancel the Workflow

In another terminal, cancel the workflow:

```bash
cadence --env development \
  --domain cadence-samples \
  workflow cancel \
  --wid <workflow_id>
```

### What Happens

1. `ActivityToBeCanceled` starts and heartbeats every second
2. When you cancel the workflow, the activity receives a cancellation signal
3. The workflow runs `CleanupActivity` in a disconnected context
4. `ActivityToBeSkipped` is never executed (skipped due to cancellation)

### Key Concept: WaitForCancellation

```go
ao := workflow.ActivityOptions{
    WaitForCancellation: true,  // Wait for activity to acknowledge cancellation
}
```

When `WaitForCancellation` is true, Cadence waits for the activity to handle the cancellation before proceeding.

### Key Concept: Disconnected Context

```go
// When workflow is canceled, get a new disconnected context for cleanup
newCtx, _ := workflow.NewDisconnectedContext(ctx)
err := workflow.ExecuteActivity(newCtx, cleanupActivity).Get(ctx, nil)
```

A disconnected context allows cleanup activities to run even after the workflow is canceled.

