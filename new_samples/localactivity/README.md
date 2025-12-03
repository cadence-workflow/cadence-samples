<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Local Activity Sample

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

## Local Activity Sample

This sample demonstrates **local activities** - lightweight activities that run in the workflow worker process.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.LocalActivityWorkflow \
  --input '"test_0_1_data"'
```

### What Happens

The workflow uses local activities to quickly check conditions, then runs regular activities only for matching conditions:

```
Input: "test_0_1_data"

Local Activities (fast, no server round-trip):
  ├── CheckCondition0("test_0_1_data") → true  (contains "_0_")
  ├── CheckCondition1("test_0_1_data") → true  (contains "_1_")
  └── CheckCondition2("test_0_1_data") → false (no "_2_")

Regular Activities (only for matching conditions):
  ├── ProcessActivity(0) → runs
  └── ProcessActivity(1) → runs
```

### Key Concept: Local vs Regular Activity

```go
// Local activity - runs in worker process, no server round-trip
lao := workflow.LocalActivityOptions{
    ScheduleToCloseTimeout: time.Second,
}
ctx = workflow.WithLocalActivityOptions(ctx, lao)
workflow.ExecuteLocalActivity(ctx, checkCondition, data)

// Regular activity - scheduled through server
ao := workflow.ActivityOptions{
    ScheduleToStartTimeout: time.Minute,
    StartToCloseTimeout:    time.Minute,
}
ctx = workflow.WithActivityOptions(ctx, ao)
workflow.ExecuteActivity(ctx, processActivity, data)
```

### When to Use Local Activities

✅ **Good for:**
- Fast validations/checks
- Data transformations
- Condition evaluation
- Operations < 1 second

❌ **Avoid for:**
- Long-running operations
- Operations needing retries
- External API calls


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

