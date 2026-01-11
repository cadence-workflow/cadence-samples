# Retry Activity Sample

This sample demonstrates **activity retry with heartbeat progress** - automatic retries with resume from last checkpoint.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for screenshots.

## How It Works

```
┌────────────────────────────────────────────────┐
│  Batch Processing (20 items)                   │
│                                                │
│  Attempt 1: Process 1-6 → FAIL (simulated)     │
│             Heartbeat: item 6 ✓                │
│                                                │
│  Attempt 2: Resume from 7 → Process 7-12 → FAIL│
│             Heartbeat: item 12 ✓               │
│                                                │
│  Attempt 3: Resume from 13 → Process 13-20 ✓   │
└────────────────────────────────────────────────┘
```

**Use case:** Batch processing, unreliable external APIs, resumable long-running tasks.

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Build the samples: `make`

## Running the Sample

```bash
# Terminal 1: Start worker
./bin/retryactivity -m worker

# Terminal 2: Trigger workflow
./bin/retryactivity -m trigger
```

## Key Code

```go
RetryPolicy: &cadence.RetryPolicy{
    InitialInterval:    time.Second,
    BackoffCoefficient: 2.0,
    MaximumAttempts:    5,
}

// Resume from heartbeat progress
if activity.HasHeartbeatDetails(ctx) {
    activity.GetHeartbeatDetails(ctx, &completedIdx)
    i = completedIdx + 1  // Resume from last checkpoint
}

activity.RecordHeartbeat(ctx, i)  // Save progress
```

## Testing

```bash
go test -v ./cmd/samples/recipes/retryactivity/
```

