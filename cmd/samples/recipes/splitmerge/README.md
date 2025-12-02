# Split Merge Sample

This sample demonstrates **divide and conquer** - splitting work into chunks, processing in parallel, then merging results.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for screenshots.

## How It Works

```
              ┌─────────────────┐
              │  Split Work     │
              │  (N chunks)     │
              └────────┬────────┘
         ┌─────────────┼─────────────┐
         ▼             ▼             ▼
   ┌──────────┐  ┌──────────┐  ┌──────────┐
   │ Chunk 1  │  │ Chunk 2  │  │ Chunk 3  │  (parallel)
   └────┬─────┘  └────┬─────┘  └────┬─────┘
         └─────────────┼─────────────┘
                       ▼
              ┌─────────────────┐
              │  Merge Results  │
              │  {count, sum}   │
              └─────────────────┘
```

**Use case:** Large file processing, batch analytics, ETL pipelines, map-reduce patterns.

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Build the samples: `make`

## Running the Sample

```bash
# Terminal 1: Start worker
./bin/splitmerge -m worker

# Terminal 2: Trigger workflow
./bin/splitmerge -m trigger
```

## Key Code

```go
chunkResultChannel := workflow.NewChannel(ctx)

for i := 1; i <= workerCount; i++ {
    workflow.Go(ctx, func(ctx workflow.Context) {
        workflow.ExecuteActivity(ctx, chunkProcessingActivity, chunkID).Get(ctx, &result)
        chunkResultChannel.Send(ctx, result)
    })
}

// Merge results
for i := 1; i <= workerCount; i++ {
    chunkResultChannel.Receive(ctx, &result)
    totalSum += result.SumInChunk
}
```

## Testing

```bash
go test -v ./cmd/samples/recipes/splitmerge/
```

