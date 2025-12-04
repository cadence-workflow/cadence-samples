## Sleep Workflow Sample

This sample demonstrates **workflow.Sleep** - pausing workflow execution for a specified duration.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 120 \
  --workflow_type cadence_samples.SleepWorkflow \
  --input '10000000000'
```

The input is sleep duration in nanoseconds (10 seconds = 10000000000).

### What Happens

```
Time 0:    Workflow starts
           └── workflow.Sleep(10s) begins

Time 10s:  Sleep completes
           └── MainSleepActivity executes

Time ~10s: Workflow completes
```

### Key Concept: workflow.Sleep

```go
func SleepWorkflow(ctx workflow.Context, sleepDuration time.Duration) error {
    // Sleep for the specified duration
    err := workflow.Sleep(ctx, sleepDuration)
    if err != nil {
        return err
    }
    
    // Continue with workflow logic after sleep
    return workflow.ExecuteActivity(ctx, MainSleepActivity).Get(ctx, nil)
}
```

### Sleep vs Timer

- `workflow.Sleep(ctx, duration)` - Simple, blocks workflow execution
- `workflow.NewTimer(ctx, duration)` - Returns a Future, can be used with Selector for racing

### Use Cases

- Delayed processing
- Rate limiting
- Scheduled tasks
- Waiting periods between retries

