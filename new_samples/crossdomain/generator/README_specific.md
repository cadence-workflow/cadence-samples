## Cross Domain Sample

This sample demonstrates executing **child workflows across different Cadence domains**.

### Prerequisites

Register a second domain for the child workflow:

```bash
cadence --env development --domain child-domain domain register
```

Start a worker in the child domain (separate terminal):

```bash
# Worker for child-domain would need to be configured separately
```

### Start the Parent Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 120 \
  --workflow_type cadence_samples.CrossDomainWorkflow
```

### Key Concept: Cross-Domain Child Options

```go
childCtx := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
    Domain:                       "other-domain",  // Different domain!
    WorkflowID:                   "child-wf-123",
    TaskList:                     "other-task-list",
    ExecutionStartToCloseTimeout: time.Minute,
})

workflow.ExecuteChildWorkflow(childCtx, ChildWorkflow, args...)
```

### Use Cases

- Multi-tenant architectures
- Isolation between teams/services
- Cross-cluster workflow execution

