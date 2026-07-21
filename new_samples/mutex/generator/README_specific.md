## How It Works

This sample implements a distributed mutex using Cadence workflows and signals. Two competing workflows acquire the same lock sequentially — the second one waits until the first releases it.

```
Workflow A                  mutexWorkflow              Workflow B
    │                           │                          │
    │── SignalWithStart ────────▶│                          │
    │◀─ AcquireLock signal ─────│                          │
    │   (lock granted)          │                          │
    │                           │◀── SignalWithStart ──────│
    │   (critical section)      │    (queued, waiting)     │
    │── unlock signal ─────────▶│                          │
    │                           │── AcquireLock signal ───▶│
    │                           │   (lock granted)         │
```

Key concepts:
- **SignalWithStart**: Atomically starts the mutex workflow if it does not exist, or signals the existing one
- **SideEffect**: Used inside the mutex workflow to deterministically generate the per-caller release channel name
- **Local activity**: `signalWithStartMutexWorkflowActivity` runs in-process, keeping the SignalWithStart call close to the worker

## Running the Sample

Start the worker:
```bash
go run .
```

Start two competing workflows with the same resource ID to observe the mutex in action:
```bash
RESOURCE_ID=$(uuidgen | tr '[:upper:]' '[:lower:]')

cadence --env development \
  --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.sampleWorkflowWithMutex \
  --tl cadence-samples-worker \
  --et 600 \
  --input "\"$RESOURCE_ID\""

cadence --env development \
  --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.sampleWorkflowWithMutex \
  --tl cadence-samples-worker \
  --et 600 \
  --input "\"$RESOURCE_ID\""
```

Watch the worker logs — the second workflow will log "waiting for mutex" until the first completes its 10-second critical section and releases the lock.
