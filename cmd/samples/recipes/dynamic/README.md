# Dynamic Sample

This sample demonstrates **dynamic activity invocation** - calling activities by string name instead of function reference.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for screenshots.

## How It Works

```
┌──────────────────────────────────────────────┐
│            Workflow                          │
│                                              │
│  ExecuteActivity(ctx, "main.getGreeting")    │
│  ExecuteActivity(ctx, "main.getName")        │
│  ExecuteActivity(ctx, "main.sayGreeting")    │
│                                              │
│  (activities called by STRING name)          │
└──────────────────────────────────────────────┘
```

**Use case:** Plugin systems, configuration-driven workflows, runtime activity selection.

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Build the samples: `make`

## Running the Sample

```bash
# Terminal 1: Start worker
./bin/dynamic -m worker

# Terminal 2: Trigger workflow
./bin/dynamic -m trigger
```

## Key Code

```go
const getGreetingActivityName = "main.getGreetingActivity"

// Call activity by name string instead of function
workflow.ExecuteActivity(ctx, getGreetingActivityName).Get(ctx, &result)
```

## Testing

```bash
go test -v ./cmd/samples/recipes/dynamic/
```

