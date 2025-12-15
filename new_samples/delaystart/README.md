<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Delay Start Sample

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

This sample demonstrates deferred workflow execution using `DelayStart` option:

```
workflow start --delay_start 30s
        │
        ▼
┌───────────────────┐
│  Workflow waits   │  ← Cadence delays start by 30s
│  in pending state │
└───────────────────┘
        │
        ▼ (after delay)
┌───────────────────┐
│DelayStartWorkflow │
│        │          │
│        ▼          │
│DelayStartActivity │
└───────────────────┘
```

The delay is handled by Cadence, not by the workflow code.

## Running the Sample

Start the worker:
```bash
go run .
```

Start a workflow with 30-second delay:
```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.DelayStartWorkflow \
  --tl cadence-samples-worker \
  --et 600 \
  --delay_start 30s \
  --input '"30s"'
```

The workflow will remain in "pending" state for 30 seconds before starting.

## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

