<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Split Merge Sample

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

This sample demonstrates the **split-merge** pattern using Cadence coroutines:

```
                    ┌─────────────────┐
                    │SplitMergeWorkflow│
                    │  (workerCount=3) │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              ▼              ▼              ▼
        workflow.Go    workflow.Go    workflow.Go
              │              │              │
              ▼              ▼              ▼
        ┌─────────┐    ┌─────────┐    ┌─────────┐
        │ Chunk 1 │    │ Chunk 2 │    │ Chunk 3 │
        │Activity │    │Activity │    │Activity │
        └────┬────┘    └────┬────┘    └────┬────┘
              │              │              │
              └──────────────┼──────────────┘
                             ▼
                    ┌─────────────────┐
                    │  Merge Results  │
                    │ (totalSum, etc) │
                    └─────────────────┘
```

Key concepts:
- **workflow.Go**: Launch concurrent coroutines (NOT goroutines)
- **workflow.NewChannel**: Create channels for coroutine communication
- Results are collected and merged after all chunks complete

## Running the Sample

Start the worker:
```bash
go run .
```

Trigger the workflow with 3 parallel workers:
```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.SplitMergeWorkflow \
  --tl cadence-samples-worker \
  --et 60 \
  --input '3'
```


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

