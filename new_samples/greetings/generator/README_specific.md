## Greetings Workflow

This sample demonstrates **sequential activity execution** - running activities one after another and passing results between them.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.GreetingsWorkflow
```

### What Happens

The workflow executes three activities in sequence:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────────┐
│ GetGreeting()   │───▶│ GetName()       │───▶│ SayGreeting(g, n)   │
│ returns "Hello" │    │ returns "Cadence"│    │ returns "Hello      │
└─────────────────┘    └─────────────────┘    │          Cadence!"  │
                                              └─────────────────────┘
```

1. `GetGreetingActivity` - Returns "Hello"
2. `GetNameActivity` - Returns "Cadence"  
3. `SayGreetingActivity` - Combines them into "Hello Cadence!"

### Key Concept: Sequential Execution

```go
// First activity
err := workflow.ExecuteActivity(ctx, GetGreetingActivity).Get(ctx, &greeting)

// Second activity (waits for first to complete)
err = workflow.ExecuteActivity(ctx, GetNameActivity).Get(ctx, &name)

// Third activity uses results from first two
err = workflow.ExecuteActivity(ctx, SayGreetingActivity, greeting, name).Get(ctx, &result)
```

Each `.Get()` call blocks until the activity completes, ensuring sequential execution.

