<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Mutex Sample

## Prerequisites

0. Install Cadence CLI. See instruction [here](https://cadenceworkflow.io/docs/cli/).
1. Run the Cadence server:
    1. Clone the [Cadence](https://github.com/cadence-workflow/cadence) repository if you haven't done already: `git clone https://github.com/cadence-workflow/cadence.git`
    2. Run `docker compose -f docker/docker-compose.yml up` to start Cadence server
    3. See more details at https://github.com/uber/cadence/blob/master/README.md
2. Once everything is up and running in Docker, open [localhost:8088](localhost:8088) to view Cadence UI.
3. Register the `cadence-samples` domain:

```bash
cadence --domain cadence-samples domain register
```

Refresh the [domains page](http://localhost:8088/domains) from step 2 to verify `cadence-samples` is registered.

## Steps to run sample

Inside the folder this sample is defined, run the following command:

```bash
go run .
```

This will call the main function in main.go which starts the worker, which will be execute the sample workflow code

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

## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

