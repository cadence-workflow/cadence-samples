<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Signal Workflow Sample

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

## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

