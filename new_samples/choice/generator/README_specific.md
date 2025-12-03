## Choice Workflow Sample

This sample demonstrates **conditional execution** based on activity results.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.ChoiceWorkflow
```

### Key Concept: Conditional Branching

```go
var order string
workflow.ExecuteActivity(ctx, GetOrderActivity).Get(ctx, &order)
switch order {
case "apple":
    workflow.ExecuteActivity(ctx, ProcessAppleActivity)
case "banana":
    workflow.ExecuteActivity(ctx, ProcessBananaActivity)
}
```

