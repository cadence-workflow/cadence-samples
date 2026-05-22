## Retry Activity Sample

This sample demonstrates **retry policies** and **heartbeat progress tracking** for unreliable activities.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 600 \
  --workflow_type cadence_samples.RetryWorkflow
```

### What Happens

The activity processes 20 tasks but intentionally fails after every ~7 tasks (1/3 of batch). With retry policy, it resumes from the last heartbeated progress.

```
Attempt 1: Process tasks 0-6, fail, heartbeat progress=6
Attempt 2: Resume from 7, process 7-13, fail, heartbeat progress=13  
Attempt 3: Resume from 14, process 14-19, complete!
```

### Key Concept: Retry Policy

```go
ao := workflow.ActivityOptions{
    RetryPolicy: &cadence.RetryPolicy{
        InitialInterval:    time.Second,      // First retry after 1s
        BackoffCoefficient: 2.0,              // Double wait each retry
        MaximumInterval:    time.Minute,      // Cap at 1 minute
        MaximumAttempts:    5,                // Give up after 5 tries
        NonRetriableErrorReasons: []string{"bad-error"}, // Don't retry these
    },
}
```

### Key Concept: Heartbeat with Progress

```go
// Record progress during activity execution
activity.RecordHeartbeat(ctx, currentTaskID)

// On retry, resume from last heartbeated progress
if activity.HasHeartbeatDetails(ctx) {
    var lastCompletedID int
    activity.GetHeartbeatDetails(ctx, &lastCompletedID)
    startFrom = lastCompletedID + 1
}
```

This avoids reprocessing already-completed work after a failure.

