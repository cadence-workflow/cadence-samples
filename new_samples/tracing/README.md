<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Tracing Sample

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

## Tracing Sample

This sample demonstrates **distributed tracing** integration with Cadence workflows.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.TracingWorkflow \
  --input '"World"'
```

### Key Concept: Trace Propagation

Cadence automatically propagates trace context through:
- Workflow execution
- Activity execution  
- Child workflows
- Signals and queries

To enable tracing, configure a tracer when creating the worker:

```go
workerOptions := worker.Options{
    Tracer: opentracing.GlobalTracer(),
}
```

### Integration with Jaeger/Zipkin

1. Set up a tracing backend (Jaeger, Zipkin, etc.)
2. Configure the tracer in your worker
3. View traces in your tracing UI to see the full execution path


## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

