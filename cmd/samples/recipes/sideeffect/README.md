# Side Effect Sample

This sample demonstrates **workflow.SideEffect** - handling non-deterministic operations safely.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for screenshots.

## How It Works

```
┌─────────────────────────────────────────────────┐
│  Workflow                                       │
│                                                 │
│  SideEffect(func() {                            │
│      return uuid.New()  // Non-deterministic!   │
│  })                                             │
│                                                 │
│  First execution: Generates UUID, stores it     │
│  Replay: Returns stored UUID (no regeneration)  │
└─────────────────────────────────────────────────┘
```

**Use case:** UUID generation, random numbers, timestamps, external state queries.

## Why SideEffect?

Workflow code must be deterministic for replay. `SideEffect` captures non-deterministic values once and returns the same value on replay.

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Build the samples: `make`

## Running the Sample

```bash
./bin/sideeffect
```

**Note:** This sample starts worker, triggers workflow, and queries result in one command.

## Key Code

```go
var value string
workflow.SideEffect(ctx, func(ctx workflow.Context) interface{} {
    return uuid.New().String()  // Only executed once, stored for replay
}).Get(&value)
```

## References

- [Side Effects](https://cadenceworkflow.io/docs/go-client/side-effect/)

