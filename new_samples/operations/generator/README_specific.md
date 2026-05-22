## Samples in this folder

This folder contains samples demonstrating workflow operations and lifecycle management in Cadence.

### Cancel Workflow

The `CancelWorkflow` demonstrates how to properly handle workflow cancellation, including:
- Graceful cleanup when a workflow is cancelled
- Using a disconnected context to run cleanup activities after cancellation
- Heartbeating in long-running activities to detect cancellation

#### Start the workflow

```bash
cadence --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.CancelWorkflow \
  --tl cadence-samples-worker \
  --et 600 \
  --input '{}'
```

Copy the workflow ID from the output.

#### Cancel the workflow

```bash
cadence --domain cadence-samples \
  workflow cancel \
  --workflow_id <YOUR_WORKFLOW_ID>
```

#### What to observe

After cancellation:
1. The `ActivityToBeCanceled` will detect the cancellation via `ctx.Done()` and return
2. The `ActivityToBeSkipped` will not be scheduled (context already cancelled)
3. The `CleanupActivity` will run using a disconnected context to perform cleanup

This pattern is essential for workflows that need to release resources or perform cleanup operations when cancelled.
