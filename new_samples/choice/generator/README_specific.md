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

