<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Signal Counter Sample

## Prerequisites

0. Install Cadence CLI. See instruction [here](https://cadenceworkflow.io/docs/cli/).
1. Run the Cadence server:
    1. Clone the [Cadence](https://github.com/cadence-workflow/cadence) repository if you haven't done already: `git clone https://github.com/cadence-workflow/cadence.git`
    2. Run `docker compose -f docker/docker-compose.yml up` to start Cadence server
    3. See more details at https://github.com/uber/cadence/blob/master/README.md
2. Once everything is up and running in Docker, open [localhost:8088](localhost:8088) to view Cadence UI.
3. Register the `cadence-samples` domain:

```bash
cadence --env development --domain cadence-samples domain register
```

Refresh the [domains page](http://localhost:8088/domains) from step 2 to verify `cadence-samples` is registered.

## Steps to run sample

Inside the folder this sample is defined, run the following command:

```bash
go run .
```

This will call the main function in main.go which starts the worker, which will be execute the sample workflow code

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


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

