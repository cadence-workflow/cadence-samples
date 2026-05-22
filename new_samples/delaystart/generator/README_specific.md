## How It Works

This sample demonstrates deferred workflow execution using `DelayStart` option:

```
workflow start --delay_start 30s
        │
        ▼
┌───────────────────┐
│  Workflow waits   │  ← Cadence delays start by 30s
│  in pending state │
└───────────────────┘
        │
        ▼ (after delay)
┌───────────────────┐
│DelayStartWorkflow │
│        │          │
│        ▼          │
│DelayStartActivity │
└───────────────────┘
```

The delay is handled by Cadence, not by the workflow code.

## Running the Sample

Start the worker:
```bash
go run .
```

Start a workflow with 30-second delay:
```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.DelayStartWorkflow \
  --tl cadence-samples-worker \
  --et 600 \
  --delay_start 30s \
  --input '"30s"'
```

The workflow will remain in "pending" state for 30 seconds before starting.
