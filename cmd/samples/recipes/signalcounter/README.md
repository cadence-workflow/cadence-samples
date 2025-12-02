# Signal Counter Sample

This sample demonstrates **signal handling with ContinueAsNew** - processing unlimited signals by resetting history.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for screenshots.

## How It Works

```
┌─────────────────────────────────────────────────┐
│  Workflow (counter = 0)                         │
│                                                 │
│  Signal channelA: +5  → counter = 5             │
│  Signal channelB: +3  → counter = 8             │
│  Signal channelA: +2  → counter = 10            │
│                                                 │
│  (maxSignalsPerExecution reached)               │
│              │                                  │
│              ▼                                  │
│  ContinueAsNew(counter = 10)                    │
│              │                                  │
│              ▼                                  │
│  New Workflow (counter = 10)                    │
│  ... continues receiving signals ...            │
└─────────────────────────────────────────────────┘
```

**Use case:** Event aggregation, counters, long-running signal processors.

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Build the samples: `make`

## Running the Sample

```bash
# Terminal 1: Start worker
./bin/signalcounter -m worker

# Terminal 2: Trigger workflow
./bin/signalcounter -m trigger

# Terminal 3: Send signals (copy WorkflowID)
./bin/signalcounter -m signal -w <WorkflowID> -n channelA -i 5
./bin/signalcounter -m signal -w <WorkflowID> -n channelB -i 3
```

## Key Code

```go
for {
    selector.AddReceive(workflow.GetSignalChannel(ctx, "channelA"), handler)
    selector.AddReceive(workflow.GetSignalChannel(ctx, "channelB"), handler)
    
    if signalsPerExecution >= maxSignalsPerExecution {
        return workflow.NewContinueAsNewError(ctx, workflow, counter)
    }
    selector.Select(ctx)
}
```

## Testing

```bash
go test -v ./cmd/samples/recipes/signalcounter/
```

