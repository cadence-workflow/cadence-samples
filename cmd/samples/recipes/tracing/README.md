# Tracing Sample

This sample demonstrates **distributed tracing integration** - adding observability to workflows with OpenTracing/Jaeger.

> **Looking for a visual guide?** See [new_samples/hello_world](../../../../new_samples/hello_world/) for screenshots.

## How It Works

```
┌─────────────────────────────────────────────────┐
│  Jaeger / OpenTracing                           │
│                                                 │
│  Trace: helloWorldWorkflow                      │
│    └── Span: ExecuteActivity                    │
│         └── Span: helloWorldActivity            │
│                                                 │
│  (traces propagate across workflow/activity)    │
└─────────────────────────────────────────────────┘
```

**Use case:** Performance monitoring, debugging, APM integration, observability.

## Prerequisites

1. Cadence server running (see [main README](../../../../README.md))
2. Jaeger or compatible tracing backend
3. Build the samples: `make`

## Running the Sample

```bash
# Terminal 1: Start worker (with tracing configured)
./bin/tracing -m worker

# Terminal 2: Trigger workflow
./bin/tracing -m trigger
```

View traces in Jaeger UI (typically http://localhost:16686).

## Key Code

```go
// Configure tracer in SampleHelper
h.Tracer = opentracing.GlobalTracer()

// Traces automatically propagate through:
// - Workflow execution
// - Activity execution
// - Child workflows
```

## References

- [Cadence Tracing](https://cadenceworkflow.io/docs/go-client/tracing/)
- [OpenTracing](https://opentracing.io/)

