## Saga Sample

This sample shows the **Saga** pattern: a workflow runs several steps, registers a
compensation after each step that succeeds, and, if a later step fails, runs those
compensations in **reverse order** to undo the work that was already done.

The Go client has no built-in Saga type, so `saga.go` contains a small helper:

```go
saga := &Saga{}

// after a step succeeds
saga.AddCompensation(func(ctx workflow.Context) error {
    return workflow.ExecuteActivity(ctx, CompensateStepActivity, step).Get(ctx, nil)
})

// when a later step fails
saga.Compensate(ctx) // runs the compensations, last registered first
```

`Compensate` uses a disconnected context, so compensations still run if the workflow
is cancelled, and it keeps going when one compensation fails so every step gets a
chance to be undone.

### Success path

All three steps succeed and nothing is compensated:

```bash
cadence \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.SagaWorkflow \
  --input '{"FailAtStep": 0}'
```

### Failure path

Step 3 fails, so steps 2 and 1 are compensated, in that order:

```bash
cadence \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.SagaWorkflow \
  --input '{"FailAtStep": 3}'
```

The worker logs show:

```
Step done            step=1
Step done            step=2
Step failing on purpose  step=3
Step compensated     step=2
Step compensated     step=1
```

The workflow then ends as failed, because the business operation did not complete.
In the Cadence UI ([localhost:8088](http://localhost:8088)) the history shows each
`CompensateStepActivity` after the failed step. Try `"FailAtStep": 1` to see that a
step that never succeeded has nothing to compensate. Values outside `0`–`3` are
rejected and the workflow fails before running any step.

### Run the tests

```bash
go test .
```

### Note: when to register a compensation

This sample registers a compensation **after** its step succeeds, which is the simplest
rule to follow. If a step can fail after partly applying its effect (for example, a
timeout after the remote system already accepted the request), register the
compensation **before** running the step instead, and make the compensation safe to
run for something that may not exist.

