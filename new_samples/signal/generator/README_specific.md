## Simple Signal Workflow

This workflow takes an input message and greet you as response. Try the following CLI

```bash
cadence --env development \
  --domain cadence-samples \
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
cadence --env development \
  --domain cadence-samples \
  workflow signal \
  --wid <workflow_id> \
  --name complete \
  --input 'true'
```
