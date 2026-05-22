## Local Activity Sample

This sample demonstrates **local activities** - lightweight activities that run in the workflow worker process.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.LocalActivityWorkflow \
  --input '"test_0_1_data"'
```

### What Happens

The workflow uses local activities to quickly check conditions, then runs regular activities only for matching conditions:

```
Input: "test_0_1_data"

Local Activities (fast, no server round-trip):
  ├── CheckCondition0("test_0_1_data") → true  (contains "_0_")
  ├── CheckCondition1("test_0_1_data") → true  (contains "_1_")
  └── CheckCondition2("test_0_1_data") → false (no "_2_")

Regular Activities (only for matching conditions):
  ├── ProcessActivity(0) → runs
  └── ProcessActivity(1) → runs
```

### Key Concept: Local vs Regular Activity

```go
// Local activity - runs in worker process, no server round-trip
lao := workflow.LocalActivityOptions{
    ScheduleToCloseTimeout: time.Second,
}
ctx = workflow.WithLocalActivityOptions(ctx, lao)
workflow.ExecuteLocalActivity(ctx, checkCondition, data)

// Regular activity - scheduled through server
ao := workflow.ActivityOptions{
    ScheduleToStartTimeout: time.Minute,
    StartToCloseTimeout:    time.Minute,
}
ctx = workflow.WithActivityOptions(ctx, ao)
workflow.ExecuteActivity(ctx, processActivity, data)
```

### When to Use Local Activities

✅ **Good for:**
- Fast validations/checks
- Data transformations
- Condition evaluation
- Operations < 1 second

❌ **Avoid for:**
- Long-running operations
- Operations needing retries
- External API calls

