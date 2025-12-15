<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Sleep Sample

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

## Sleep Workflow Sample

This sample demonstrates **workflow.Sleep** - pausing workflow execution for a specified duration.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 120 \
  --workflow_type cadence_samples.SleepWorkflow \
  --input '10000000000'
```

The input is sleep duration in nanoseconds (10 seconds = 10000000000).

### What Happens

```
Time 0:    Workflow starts
           └── workflow.Sleep(10s) begins

Time 10s:  Sleep completes
           └── MainSleepActivity executes

Time ~10s: Workflow completes
```

### Key Concept: workflow.Sleep

```go
func SleepWorkflow(ctx workflow.Context, sleepDuration time.Duration) error {
    // Sleep for the specified duration
    err := workflow.Sleep(ctx, sleepDuration)
    if err != nil {
        return err
    }
    
    // Continue with workflow logic after sleep
    return workflow.ExecuteActivity(ctx, MainSleepActivity).Get(ctx, nil)
}
```

### Sleep vs Timer

- `workflow.Sleep(ctx, duration)` - Simple, blocks workflow execution
- `workflow.NewTimer(ctx, duration)` - Returns a Future, can be used with Selector for racing

### Use Cases

- Delayed processing
- Rate limiting
- Scheduled tasks
- Waiting periods between retries


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

