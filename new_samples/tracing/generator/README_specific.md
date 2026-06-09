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

