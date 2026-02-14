## Branch Workflow Sample

This sample demonstrates **parallel activity execution** using Futures.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.BranchWorkflow
```

### Key Concept: Parallel Execution with Futures

```go
var futures []workflow.Future
for i := 1; i <= 3; i++ {
    future := workflow.ExecuteActivity(ctx, BranchActivity, input)
    futures = append(futures, future)
}
// Wait for all
for _, f := range futures {
    f.Get(ctx, nil)
}
```

