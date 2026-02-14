## How It Works

This sample demonstrates handling multiple signals with **ContinueAsNew** to prevent history bloat:

```
┌──────────────────────────┐
│  SignalCounterWorkflow   │
│                          │
│  ┌────────────────────┐  │
│  │ Listen on channelA │◀─┼── signal channelA --input 5
│  │ Listen on channelB │◀─┼── signal channelB --input 10
│  └─────────┬──────────┘  │
│            │             │
│  counter += signal value │
│            │             │
│  if signals >= 3:        │
│      ContinueAsNew ──────┼──▶ New execution with counter
│                          │
└──────────────────────────┘
```

Key concepts:
- **Multiple signal channels**: Workflow listens on both channelA and channelB
- **ContinueAsNew**: Restarts workflow with current state to prevent history growth
- **MaxSignalsPerExecution**: Limits signals before ContinueAsNew (default: 3 for demo)

## Running the Sample

Start the worker:
```bash
go run .
```

Start the workflow:
```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.SignalCounterWorkflow \
  --tl cadence-samples-worker \
  --et 600 \
  --input '0'
```

Send signals (copy workflow ID from above):
```bash
# Signal on channelA
cadence --env development \
  --domain cadence-samples \
  workflow signal \
  --wid <workflow_id> \
  --name channelA \
  --input '5'

# Signal on channelB
cadence --env development \
  --domain cadence-samples \
  workflow signal \
  --wid <workflow_id> \
  --name channelB \
  --input '10'
```

After 3 signals, the workflow will ContinueAsNew with the accumulated counter.

