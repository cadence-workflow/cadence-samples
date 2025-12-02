<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Greetings Sample

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

## Greetings Workflow

This sample demonstrates **sequential activity execution** - running activities one after another and passing results between them.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.GreetingsWorkflow
```

### What Happens

The workflow executes three activities in sequence:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────────┐
│ GetGreeting()   │───▶│ GetName()       │───▶│ SayGreeting(g, n)   │
│ returns "Hello" │    │ returns "Cadence"│    │ returns "Hello      │
└─────────────────┘    └─────────────────┘    │          Cadence!"  │
                                              └─────────────────────┘
```

1. `GetGreetingActivity` - Returns "Hello"
2. `GetNameActivity` - Returns "Cadence"  
3. `SayGreetingActivity` - Combines them into "Hello Cadence!"

### Key Concept: Sequential Execution

```go
// First activity
err := workflow.ExecuteActivity(ctx, GetGreetingActivity).Get(ctx, &greeting)

// Second activity (waits for first to complete)
err = workflow.ExecuteActivity(ctx, GetNameActivity).Get(ctx, &name)

// Third activity uses results from first two
err = workflow.ExecuteActivity(ctx, SayGreetingActivity, greeting, name).Get(ctx, &result)
```

Each `.Get()` call blocks until the activity completes, ensuring sequential execution.


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

