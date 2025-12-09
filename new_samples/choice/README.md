<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Choice Sample

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

## Choice Workflow

This sample demonstrates **conditional execution** - running different activities based on the result of a previous activity.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.ChoiceWorkflow
```

### What Happens

```
         ┌─────────────────┐
         │  ChoiceWorkflow │
         └────────┬────────┘
                  │
                  ▼
         ┌─────────────────┐
         │ GetOrderActivity│
         │ (returns random │
         │  fruit order)   │
         └────────┬────────┘
                  │
     ┌────────────┼────────────┬────────────┐
     │            │            │            │
     ▼            ▼            ▼            ▼
  "apple"     "banana"    "cherry"     (other)
     │            │            │            │
     ▼            ▼            ▼            ▼
┌─────────┐ ┌─────────┐ ┌─────────┐    ┌─────┐
│ProcessA │ │ProcessB │ │ProcessC │    │Error│
│pple     │ │anana    │ │herry    │    │     │
└─────────┘ └─────────┘ └─────────┘    └─────┘
```

1. `GetOrderActivity` returns a random fruit (apple, banana, or cherry)
2. Based on the result, the workflow executes the corresponding activity
3. Only one processing activity runs (exclusive choice)

### Key Concept: Conditional Branching

```go
var orderChoice string
err := workflow.ExecuteActivity(ctx, GetOrderActivity).Get(ctx, &orderChoice)

switch orderChoice {
case "apple":
    workflow.ExecuteActivity(ctx, ProcessAppleActivity, orderChoice)
case "banana":
    workflow.ExecuteActivity(ctx, ProcessBananaActivity, orderChoice)
case "cherry":
    workflow.ExecuteActivity(ctx, ProcessCherryActivity, orderChoice)
default:
    return errors.New("unknown order type")
}
```

### Real-World Use Cases

- Order routing based on product type
- User authentication with different providers
- Document processing based on file type


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

