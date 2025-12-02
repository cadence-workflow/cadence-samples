# Cross Domain Sample

This sample demonstrates **cross-domain child workflows** - executing child workflows in different Cadence domains.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for screenshots.

## How It Works

```
┌─────────────────────────────────────────┐
│          Parent Workflow (domain0)       │
│                    │                     │
│    ┌───────────────┴───────────────┐    │
│    ▼                               ▼    │
│ ┌─────────┐                 ┌─────────┐ │
│ │ Child   │                 │ Child   │ │
│ │ domain1 │                 │ domain2 │ │
│ │(cluster1)│                │(cluster0)│ │
│ └─────────┘                 └─────────┘ │
└─────────────────────────────────────────┘
```

**Use case:** Multi-tenant systems, cross-region workflows, domain isolation.

## Prerequisites

1. Cadence server with multiple domains configured
2. Build the samples: `make`

## Running the Sample

```bash
./bin/crossdomain
```

**Note:** This sample requires specific multi-domain setup. See main.go for domain configuration.

## Key Code

```go
ctx1 := workflow.WithChildOptions(ctx, workflow.ChildWorkflowOptions{
    Domain:   domain1,
    TaskList: tasklist1,
})
workflow.ExecuteChildWorkflow(ctx1, wf1, data).Get(ctx1, nil)
```

