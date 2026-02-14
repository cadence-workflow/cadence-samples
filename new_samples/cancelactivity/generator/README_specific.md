## How It Works

This sample demonstrates graceful workflow cancellation with cleanup:

```
┌──────────────────────┐
│   CancelWorkflow     │
│                      │
│  ┌────────────────┐  │     Cancel Signal
│  │ LongRunning    │◀─┼─────────────────────
│  │ Activity       │  │
│  │ (heartbeating) │  │
│  └───────┬────────┘  │
│          │           │
│    On Cancel:        │
│          ▼           │
│  ┌────────────────┐  │
│  │ CleanupActivity│  │  ← Runs in disconnected context
│  └────────────────┘  │
└──────────────────────┘
```

Key concepts:
- **WaitForCancellation**: Activity option that waits for activity to acknowledge cancel
- **NewDisconnectedContext**: Creates a context unaffected by workflow cancellation
- **IsCanceledError**: Check if an error is due to cancellation

## Running the Sample

Start the worker:
```bash
go run .
```

Trigger a workflow:
```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.CancelWorkflow \
  --tl cadence-samples-worker \
  --et 600
```

Cancel the workflow (copy workflow ID from above):
```bash
cadence --env development \
  --domain cadence-samples \
  workflow cancel \
  --wid <workflow_id>
```

Watch the worker logs to see the cleanup activity run.
