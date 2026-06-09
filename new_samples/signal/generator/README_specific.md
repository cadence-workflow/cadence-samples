## Simple Signal Workflow

This workflow takes an input message and greet you as response. Try the following CLI

```bash
cadence --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.SimpleSignalWorkflow
```

Verify that your workflow started. Your can find your worklow by looking at the "Workflow type" column.

If this is your first sample, please refer to [HelloWorkflow sample](https://github.com/cadence-workflow/cadence-samples/tree/master/new_samples/hello_world) about how to view your workflows.


### Signal your workflow

This workflow will need a signal to complete successfully. Below is how you can send a signal. In this example, we are sending a `bool` value `true` (JSON formatted) via the signal called `complete`

```bash
cadence --domain cadence-samples \
  workflow signal \
  --wid <workflow_id> \
  --name complete \
  --input 'true'
```

## Await Signal Workflow

The `AwaitSignalWorkflow` demonstrates how to handle multiple signals that may
arrive **out of order** but must be **processed in a fixed sequential order**,
while enforcing two timeouts.

This sample waits for three signals: `Signal1`, `Signal2`, `Signal3`. They can
be sent in any order, but the workflow always processes them in the order
`Signal1` -> `Signal2` -> `Signal3`. Two timeouts are enforced:

- `signalToSignalTimeout` (30s): the maximum time allowed between two signals.
- `fromFirstSignalTimeout` (60s): the maximum total time from the first signal.

The workflow runs a separate goroutine (`Listen`) that only records which
signals have arrived, while the main workflow uses `workflow.Await` to process
them in order. After each signal is processed, a corresponding activity
(`Signal1Activity`, `Signal2Activity`, `Signal3Activity`) is executed.

### Start the workflow

Use a fixed workflow id (`-w`) so you can send signals to it afterwards. Give it
an execution timeout (`--et`) large enough to send all signals (e.g. 180s).

```bash
cadence --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.AwaitSignalWorkflow \
  --tl cadence-samples-worker \
  --et 180 \
  -w await-signal-demo
```

### Send the signals

The signals carry no payload here, so no `--input` is needed. The workflow
processes them in the order 1 -> 2 -> 3 no matter the order you send them.

#### In order (1, 2, 3)

```bash
cadence --domain cadence-samples workflow signal --wid await-signal-demo --name Signal1
cadence --domain cadence-samples workflow signal --wid await-signal-demo --name Signal2
cadence --domain cadence-samples workflow signal --wid await-signal-demo --name Signal3
```

#### Out of order (3, 1, 2)

```bash
cadence --domain cadence-samples workflow signal --wid await-signal-demo --name Signal3
cadence --domain cadence-samples workflow signal --wid await-signal-demo --name Signal1
cadence --domain cadence-samples workflow signal --wid await-signal-demo --name Signal2
```

Even though `Signal3` arrives first, the worker logs show the workflow still
runs the activities in order: `Signal1Activity` -> `Signal2Activity` -> `Signal3Activity`.

### Trigger a timeout (optional)

To see the timeout behavior, send `Signal1` and then wait more than 30 seconds
before sending `Signal2`. The workflow fails with a custom error
(`Signal2 not received`) because the signal-to-signal timeout is exceeded.

```bash
cadence --domain cadence-samples workflow signal --wid await-signal-demo --name Signal1
# wait > 30s, then:
cadence --domain cadence-samples workflow signal --wid await-signal-demo --name Signal2
```

### Credits

The Await Signal Workflow is a Cadence port of the Temporal Go SDK
[await-signals sample](https://github.com/temporalio/samples-go/tree/main/await-signals).
