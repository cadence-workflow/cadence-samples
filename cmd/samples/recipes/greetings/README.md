# Greetings Sample

This sample demonstrates **sequential activity execution** - running multiple activities one after another, passing results between them.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for a step-by-step tutorial with screenshots.


## How It Works

The workflow executes 3 activities in sequence:

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────────┐
│ getGreeting()   │───▶│ getName()       │───▶│ sayGreeting(g, n)   │
│ returns "Hello" │    │ returns "Cadence"│    │ returns "Hello      │
└─────────────────┘    └─────────────────┘    │          Cadence!"  │
                                              └─────────────────────┘
```

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Build the samples:
   ```bash
   make
   ```

## Running the Sample

### Step 1: Start the Worker

```bash
./bin/greetings -m worker
```

### Step 2: Trigger the Workflow

```bash
./bin/greetings -m trigger
```

### Step 3: View the Result

Check the worker terminal:
```
Workflow completed. {"Result": "Greeting: Hello Cadence!\n"}
```

Or view in the Cadence Web UI at [localhost:8088](http://localhost:8088).

## Key Code

### Sequential Execution Pattern

```go
// Activity 1: Get greeting
var greetResult string
err := workflow.ExecuteActivity(ctx, getGreetingActivity).Get(ctx, &greetResult)

// Activity 2: Get name
var nameResult string
err = workflow.ExecuteActivity(ctx, getNameActivity).Get(ctx, &nameResult)

// Activity 3: Combine results
var sayResult string
err = workflow.ExecuteActivity(ctx, sayGreetingActivity, greetResult, nameResult).Get(ctx, &sayResult)
```

Each `ExecuteActivity().Get()` blocks until the activity completes, ensuring sequential execution.

## Testing

```bash
# Unit tests
go test -v ./cmd/samples/recipes/greetings/ -run TestUnitTestSuite

# Replay tests
go test -v ./cmd/samples/recipes/greetings/ -run TestReplayWorkflowHistoryFromFile
```

## References

- [Cadence Documentation](https://cadenceworkflow.io)
- [Activity Basics](https://cadenceworkflow.io/docs/concepts/activities/)

