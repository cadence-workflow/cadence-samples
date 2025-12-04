## Context Propagation Sample

This sample demonstrates **custom context propagation** across workflow and activity execution.

### Key Concept: Context Propagator

A context propagator allows you to pass custom values (like user IDs, trace IDs, tenant info) through:
- Workflow execution
- Activity execution
- Child workflows

```go
type propagator struct{}

func (s *propagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
    // Serialize and inject values into headers
}

func (s *propagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
    // Extract values from headers into context
}
```

### Configuring the Worker

Register the propagator when creating the worker:

```go
workerOptions := worker.Options{
    ContextPropagators: []workflow.ContextPropagator{
        NewContextPropagator(),
    },
}
```

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 60 \
  --workflow_type cadence_samples.CtxPropagationWorkflow
```

### Use Cases

- Distributed tracing (trace IDs)
- Multi-tenancy (tenant IDs)
- User context (user IDs, auth tokens)
- Request correlation

