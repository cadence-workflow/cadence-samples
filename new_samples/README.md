# Cadence Samples

This directory contains samples demonstrating various Cadence workflow concepts. Each sample is self-contained in its own concept folder.

## Available Samples

| Folder | Description |
|--------|-------------|
| [activities/](activities/) | Activity patterns: dynamic execution by name, parallel execution with pick-first |
| [client_tls/](client_tls/) | Client-side TLS configuration for secure Cadence connections |
| [data/](data/) | Custom DataConverters: gzip compression, AES-256-GCM encryption, and S3 "claim-check" offload |
| [hello_world/](hello_world/) | Basic "Hello World" workflow and activity |
| [operations/](operations/) | Workflow operations: cancellation and cleanup patterns |
| [query/](query/) | Workflow query patterns |
| [signal/](signal/) | Workflow signal patterns |

## Prerequisites

1. Install Cadence CLI: [https://cadenceworkflow.io/docs/cli/](https://cadenceworkflow.io/docs/cli/)
2. Run the Cadence server:
   ```bash
   git clone https://github.com/cadence-workflow/cadence.git
   cd cadence
   docker compose -f docker/docker-compose.yml up
   ```
3. Open [localhost:8088](http://localhost:8088) to view Cadence UI
4. Register the `cadence-samples` domain:
   ```bash
   cadence --domain cadence-samples domain register
   ```

## Running a Sample

Each sample folder is self-contained. Navigate to any sample folder and run:

```bash
go run .
```

This starts the worker for that sample. Then use the Cadence CLI to start workflows as described in each sample's README.

---

## Adding a New Sample

New samples should follow the template-based structure for consistency. The `template/` directory contains Go templates that generate boilerplate code.

### Step 1: Create Your Sample Folder

```bash
mkdir my_sample
cd my_sample
```

### Step 2: Create Your Workflow Code

Create a file (e.g., `my_workflow.go`) with `package main`:

```go
package main

import (
    "context"
    "go.uber.org/cadence/activity"
    "go.uber.org/cadence/workflow"
    "time"
)

func MyWorkflow(ctx workflow.Context) (string, error) {
    ao := workflow.ActivityOptions{
        ScheduleToStartTimeout: time.Minute,
        StartToCloseTimeout:    time.Minute,
    }
    ctx = workflow.WithActivityOptions(ctx, ao)

    var result string
    err := workflow.ExecuteActivity(ctx, MyActivity).Get(ctx, &result)
    return result, err
}

func MyActivity(ctx context.Context) (string, error) {
    return "Hello from my sample!", nil
}
```

### Step 3: Create the Generator

Create `generator/generate.go`:

```go
package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
    data := template.TemplateData{
        SampleName: "My Sample",
        Workflows:  []string{"MyWorkflow"},
        Activities: []string{"MyActivity"},
    }

    template.GenerateAll(data)
}
```

### Step 4: Create Sample-Specific Documentation

Create `generator/README_specific.md` with documentation specific to your sample:

```markdown
## My Sample

Description of what this sample demonstrates...

### Start the workflow

\```bash
cadence --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.MyWorkflow \
  --tl cadence-samples-worker \
  --et 60 \
  --input '{}'
\```
```

### Step 5: Run the Generator

```bash
cd generator
go run generate.go
```

This generates:
- `../worker.go` - Worker setup and registration
- `../main.go` - Entry point that starts the worker
- `../README.md` - Combined documentation
- `README.md` - Generator-specific README

### Template Files

The `template/` directory contains:

| File | Purpose |
|------|---------|
| `generator.go` | Go code that powers the generation |
| `worker.tmpl` | Template for worker.go |
| `main.tmpl` | Template for main.go |
| `README.tmpl` | Template for README header (prerequisites) |
| `README_references.tmpl` | Template for README footer (references) |
| `README_generator.tmpl` | Template for generator/README.md |

## Learn More

- [Cadence Documentation](https://cadenceworkflow.io/docs)
- [Cadence Go Client](https://github.com/uber-go/cadence-client)
- [Cadence Server](https://github.com/uber/cadence)
