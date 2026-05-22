<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Dynamic Invocation Sample

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

This sample demonstrates invoking activities by **string name** rather than function reference:

```go
// Instead of:
workflow.ExecuteActivity(ctx, GetGreetingActivity)

// Use string name:
workflow.ExecuteActivity(ctx, "main.getGreetingActivity")
```

This enables:
- Plugin architectures where activities are loaded at runtime
- Configuration-driven workflows
- Cross-language activity invocation

```
┌─────────────────────────┐
│ DynamicGreetingsWorkflow│
│                         │
│  ExecuteActivity(ctx,   │
│    "main.getGreeting")  │──▶ GetGreetingActivity
│         │               │
│  ExecuteActivity(ctx,   │
│    "main.getName")      │──▶ GetNameActivity
│         │               │
│  ExecuteActivity(ctx,   │
│    "main.sayGreeting")  │──▶ SayGreetingActivity
└─────────────────────────┘
```

## Running the Sample

Start the worker:
```bash
go run .
```

Trigger the workflow:
```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.DynamicGreetingsWorkflow \
  --tl cadence-samples-worker \
  --et 60
```


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

