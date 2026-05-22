## How It Works

This sample demonstrates invoking activities by **string name** rather than function reference:

```go
// Instead of:
workflow.ExecuteActivity(ctx, GetGreetingActivity)

// Use string name:
workflow.ExecuteActivity(ctx, "main.getGreetingActivity")
```

This enables:
- Plugin architectures where activities are loaded at runtime
- Configuration-driven workflows
- Cross-language activity invocation

```
┌─────────────────────────┐
│ DynamicGreetingsWorkflow│
│                         │
│  ExecuteActivity(ctx,   │
│    "main.getGreeting")  │──▶ GetGreetingActivity
│         │               │
│  ExecuteActivity(ctx,   │
│    "main.getName")      │──▶ GetNameActivity
│         │               │
│  ExecuteActivity(ctx,   │
│    "main.sayGreeting")  │──▶ SayGreetingActivity
└─────────────────────────┘
```

## Running the Sample

Start the worker:
```bash
go run .
```

Trigger the workflow:
```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.DynamicGreetingsWorkflow \
  --tl cadence-samples-worker \
  --et 60
```

