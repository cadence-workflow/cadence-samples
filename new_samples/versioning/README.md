<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Versioning Sample

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

## Versioning Sample

This sample demonstrates **workflow versioning** for safe code deployments.

### The Problem

Changing workflow code can break running workflows during replay because the decision history no longer matches.

### The Solution: GetVersion

```go
version := workflow.GetVersion(ctx, "change-id", workflow.DefaultVersion, 1)
if version == workflow.DefaultVersion {
    // Old code path
} else {
    // New code path  
}
```

- **DefaultVersion (-1)**: Workflows started before the change
- **Version 1+**: Workflows started after the change

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.VersionedWorkflow
```

### Deployment Strategy

1. Deploy new code with GetVersion branching
2. New workflows use version 1, old workflows continue with DefaultVersion
3. Once all old workflows complete, remove DefaultVersion branch


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

