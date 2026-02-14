## Greetings Workflow Sample

This sample demonstrates **sequential activity execution**.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.GreetingsWorkflow
```

### Key Concept: Sequential Execution

```go
// Each .Get() blocks until complete
greeting, _ := workflow.ExecuteActivity(ctx, GetGreetingActivity).Get(ctx, &greeting)
name, _ := workflow.ExecuteActivity(ctx, GetNameActivity).Get(ctx, &name)
workflow.ExecuteActivity(ctx, SayGreetingActivity, greeting, name).Get(ctx, &result)
```

