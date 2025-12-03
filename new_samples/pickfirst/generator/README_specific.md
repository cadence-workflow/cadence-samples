## Pick First Workflow Sample

This sample demonstrates **racing activities** and using the first result.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.PickFirstWorkflow
```

### Key Concept: Race and Cancel

```go
childCtx, cancelHandler := workflow.WithCancel(ctx)
selector := workflow.NewSelector(ctx)
selector.AddFuture(f1, handler).AddFuture(f2, handler)
selector.Select(ctx)  // Wait for first
cancelHandler()       // Cancel others
```

