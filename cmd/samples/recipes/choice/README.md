# Choice Sample

This sample demonstrates **conditional workflow logic** - executing different activities based on runtime decisions.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for a step-by-step tutorial with screenshots.

## Two Patterns

### 1. Exclusive Choice (`-c single`)
Execute ONE activity based on a decision:

```
┌───────────────┐
│ getOrder()    │
│ returns "apple"│
└───────┬───────┘
        │
   ┌────┴────┬────────┬────────┐
   ▼         ▼        ▼        ▼
┌──────┐ ┌──────┐ ┌──────┐ ┌──────┐
│apple │ │banana│ │cherry│ │orange│
│  ✓   │ │      │ │      │ │      │
└──────┘ └──────┘ └──────┘ └──────┘
```

### 2. Multi Choice (`-c multi`)
Execute MULTIPLE activities in parallel based on a basket:

```
┌───────────────────┐
│ getBasketOrder()  │
│ returns ["apple", │
│  "cherry","orange"]│
└─────────┬─────────┘
    ┌─────┼─────┐
    ▼     ▼     ▼
┌──────┐┌──────┐┌──────┐
│apple ││cherry││orange│  (parallel)
│  ✓   ││  ✓   ││  ✓   │
└──────┘└──────┘└──────┘
```

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Build the samples: `make`

## Running the Sample

```bash
# Terminal 1: Start worker
./bin/choice -m worker

# Terminal 2: Trigger workflows
./bin/choice -m trigger -c single    # Exclusive choice (default)
./bin/choice -m trigger -c multi     # Multi choice
```

## Key Code

### Exclusive Choice
```go
var orderChoice string
workflow.ExecuteActivity(ctx, getOrderActivity).Get(ctx, &orderChoice)

switch orderChoice {
case "apple":
    workflow.ExecuteActivity(ctx, orderAppleActivity, orderChoice)
case "banana":
    workflow.ExecuteActivity(ctx, orderBananaActivity, orderChoice)
// ...
}
```

### Multi Choice
```go
var choices []string
workflow.ExecuteActivity(ctx, getBasketOrderActivity).Get(ctx, &choices)

var futures []workflow.Future
for _, item := range choices {
    switch item {
    case "apple":
        f = workflow.ExecuteActivity(ctx, orderAppleActivity, item)
    // ...
    }
    futures = append(futures, f)
}

// Wait for all
for _, future := range futures {
    future.Get(ctx, nil)
}
```

## Testing

```bash
go test -v ./cmd/samples/recipes/choice/
```

## References

- [Cadence Documentation](https://cadenceworkflow.io)

