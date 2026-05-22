## Versioning Sample

This sample demonstrates **workflow versioning** for safe code deployments.

### The Problem

Changing workflow code can break running workflows during replay because the decision history no longer matches.

### The Solution: GetVersion

```go
version := workflow.GetVersion(ctx, "change-id", workflow.DefaultVersion, 1)
if version == workflow.DefaultVersion {
    // Old code path
} else {
    // New code path  
}
```

- **DefaultVersion (-1)**: Workflows started before the change
- **Version 1+**: Workflows started after the change

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.VersionedWorkflow
```

### Deployment Strategy

1. Deploy new code with GetVersion branching
2. New workflows use version 1, old workflows continue with DefaultVersion
3. Once all old workflows complete, remove DefaultVersion branch

