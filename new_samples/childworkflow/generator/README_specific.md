## Child Workflow Sample

This sample demonstrates **parent-child workflow relationships** and the **ContinueAsNew** pattern.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.ParentWorkflow
```

### What Happens

```
┌─────────────────────┐
│   ParentWorkflow    │
└──────────┬──────────┘
           │
           │ ExecuteChildWorkflow
           ▼
┌─────────────────────┐
│   ChildWorkflow     │──┐
│   (run 1 of 5)      │  │
└─────────────────────┘  │
           │             │ ContinueAsNew
           ▼             │
┌─────────────────────┐  │
│   ChildWorkflow     │──┤
│   (run 2 of 5)      │  │
└─────────────────────┘  │
           │             │
          ...           ...
           │             │
           ▼             │
┌─────────────────────┐  │
│   ChildWorkflow     │◀─┘
│   (run 5 of 5)      │
└─────────────────────┘
           │
           │ Returns result
           ▼
┌─────────────────────┐
│   ParentWorkflow    │
│   completes         │
└─────────────────────┘
```

1. Parent workflow starts a child workflow
2. Child workflow uses `ContinueAsNew` to restart itself 5 times
3. After 5 runs, child completes and returns result to parent

### Key Concept: Child Workflow Options

```go
cwo := workflow.ChildWorkflowOptions{
    WorkflowID:                   childID,
    ExecutionStartToCloseTimeout: time.Minute,
}
ctx = workflow.WithChildOptions(ctx, cwo)

err := workflow.ExecuteChildWorkflow(ctx, ChildWorkflow, args...).Get(ctx, &result)
```

### Key Concept: ContinueAsNew

```go
// Instead of recursion (which grows history), use ContinueAsNew
return "", workflow.NewContinueAsNewError(ctx, ChildWorkflow, newArgs...)
```

ContinueAsNew starts a new workflow run with fresh history, avoiding unbounded history growth.

