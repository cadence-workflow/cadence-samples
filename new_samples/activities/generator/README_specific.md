## Dynamic Activity Workflow

This sample demonstrates invoking activities by string name rather than function reference.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.DynamicWorkflow \
  --input '{"message":"Cadence"}'
```

Verify that your workflow started and completed successfully.

### Key Concept

Instead of passing the function directly:
```go
workflow.ExecuteActivity(ctx, MyActivity, input)
```

Pass the activity name as a string:
```go
workflow.ExecuteActivity(ctx, "cadence_samples.DynamicGreetingActivity", input)
```

This is useful for plugin systems or configuration-driven workflows.

