# Child Workflow Sample

This sample demonstrates **parent-child workflow relationships** and the **ContinueAsNew** pattern for long-running workflows.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for a step-by-step tutorial with screenshots.

## How It Works

```
┌─────────────────────────────────────────────────────┐
│                 Parent Workflow                      │
│  ExecuteChildWorkflow(child, 0, 5)                  │
│                      │                               │
│                      ▼                               │
│  ┌─────────────────────────────────────────────┐    │
│  │           Child Workflow                     │    │
│  │  Run 1 → ContinueAsNew                       │    │
│  │  Run 2 → ContinueAsNew                       │    │
│  │  Run 3 → ContinueAsNew                       │    │
│  │  Run 4 → ContinueAsNew                       │    │
│  │  Run 5 → Complete ✓                          │    │
│  └─────────────────────────────────────────────┘    │
│                      │                               │
│                      ▼                               │
│  Parent receives: "completed after 5 runs"          │
└─────────────────────────────────────────────────────┘
```

**Use cases:**
- Breaking large workflows into modular pieces
- Long-running workflows that need to reset history (ContinueAsNew)
- Workflow decomposition for better organization

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Build the samples: `make`

## Running the Sample

```bash
# Terminal 1: Start worker
./bin/childworkflow -m worker

# Terminal 2: Trigger workflow
./bin/childworkflow -m trigger
```

## Key Code

### Parent Workflow
```go
cwo := workflow.ChildWorkflowOptions{
    WorkflowID:                   childID,
    ExecutionStartToCloseTimeout: time.Minute,
}
ctx = workflow.WithChildOptions(ctx, cwo)

var result string
err := workflow.ExecuteChildWorkflow(ctx, sampleChildWorkflow, 0, 5).Get(ctx, &result)
```

### Child Workflow with ContinueAsNew
```go
func sampleChildWorkflow(ctx workflow.Context, totalCount, runCount int) (string, error) {
    totalCount++
    runCount--
    
    if runCount == 0 {
        return fmt.Sprintf("completed after %v runs", totalCount), nil
    }
    
    // Restart workflow with new parameters (resets history)
    return "", workflow.NewContinueAsNewError(ctx, sampleChildWorkflow, totalCount, runCount)
}
```

## Why ContinueAsNew?

Workflow history grows with each event. For long-running workflows, use `ContinueAsNew` to:
- Reset the history size
- Prevent hitting history limits
- Keep workflows performant

## References

- [Child Workflows](https://cadenceworkflow.io/docs/go-client/child-workflows/)
- [ContinueAsNew](https://cadenceworkflow.io/docs/go-client/continue-as-new/)

