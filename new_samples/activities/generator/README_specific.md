## Samples in this folder

This folder contains samples demonstrating various activity patterns in Cadence.

### Dynamic Workflow

The `DynamicWorkflow` demonstrates executing an activity by its registered string name rather than passing the function reference directly. This pattern is useful when you need to dynamically determine which activity to execute at runtime.

```bash
cadence --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.DynamicWorkflow \
  --tl cadence-samples-worker \
  --et 60 \
  --input '{"message":"Cadence"}'
```

### Parallel Branch Pick First Workflow

The `ParallelBranchPickFirstWorkflow` demonstrates running multiple activities in parallel and returning the result of the first one to complete. This pattern is useful for scenarios like:
- Racing multiple data sources
- Implementing timeouts with fallbacks
- Redundant execution for reliability

```bash
cadence --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.ParallelBranchPickFirstWorkflow \
  --tl cadence-samples-worker \
  --et 60 \
  --input '{}'
```

The workflow will:
1. Start two parallel activities with different delays
2. Wait for the first one to complete
3. Cancel the remaining activity
4. Return the first successful result

Note: The `WaitForCancellation` option is set to `true` to demonstrate proper cleanup of cancelled activities. In production, you may set this to `false` if you don't need to wait for cancellation acknowledgment.
