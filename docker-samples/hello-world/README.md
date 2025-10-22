# Cadence Golang Hello World

This is a simple hello world example for Cadence workflow engine using the Go client.

## Table of Contents

- [Quick Start with Docker](#quick-start-with-docker)
- [Running Locally](#running-locally)
- [What's Included](#whats-included)
- [Running Workflows](#running-workflows)
- [Monitoring](#monitoring)

---

## Quick Start with Docker

**The easiest way to run this sample!**

### 1. Start Everything

From the `samples` directory:

```bash
cd /path/to/cadence/samples
docker-compose up
```

This starts:
- ✅ Cadence server
- ✅ Cassandra database
- ✅ Web UI (http://localhost:8088)
- ✅ Hello World worker

### 2. Register Domain (first time only)

```bash
docker run --network=samples_cadence-network --rm ubercadence/cli:master \
  --address cadence:7833 \
  --domain test-domain domain register
```

### 3. Trigger a Workflow

```bash
docker run --network=samples_cadence-network --rm ubercadence/cli:master \
  --address cadence:7833 \
  --domain test-domain workflow start \
  --tasklist test-worker \
  --workflow_type main.helloWorldWorkflow \
  --execution_timeout 60 \
  --input '"World"'
```

### 4. View Results

- **Web UI**: http://localhost:8088 (enter domain: `test-domain`)
- **Logs**: `docker-compose logs -f hello-world-worker`

✅ **That's it!** Your workflow is running in Docker.

---

## Running Locally

### Prerequisites

Before running locally, make sure you have:

1. **Cadence server running**: You can start it using Docker:
   ```bash
   docker-compose -f ../../docker/docker-compose.yml up
   ```

2. **Domain registered**: Register the `test-domain` domain:
   ```bash
   cadence --domain test-domain domain register
   ```
   
   Or using the dockerized CLI:
   ```bash
   docker run --network=host --rm ubercadence/cli:master --domain test-domain domain register
   ```

### Building and Running the Worker

1. **Build the worker:**
   ```bash
   cd /path/to/cadence/samples/hello-world
   GOWORK=off go build -o hello-world
   ```

2. **Run the worker:**
   ```bash
   ./hello-world
   ```

   You should see logs like:
   ```
   INFO  Started Worker.  {"worker": "test-worker"}
   ```

### Running the Workflow

In a separate terminal, while the worker is running, execute the workflow:

#### Using Cadence CLI (if installed locally):
```bash
cadence --domain test-domain workflow start --et 60 --tl test-worker --workflow_type main.helloWorldWorkflow --input '"World"'
```

#### Using Docker:
```bash
docker run --network=host --rm ubercadence/cli:master --domain test-domain workflow start --et 60 --tl test-worker --workflow_type main.helloWorldWorkflow --input '"World"'
```

### Expected Output

In your worker terminal, you should see logs similar to:

```
INFO  helloworld workflow started
INFO  helloworld activity started
INFO  Workflow completed.  {"Result": "Hello World!"}
```

---

## What's Included

This sample includes:

### Workflows

1. **`helloWorldWorkflow`** - Basic workflow that executes a simple activity
   - Fast execution (~50ms)
   - Great for testing

2. **`longRunningWorkflow`** - Demonstrates a workflow that stays open
   - Executes activity then sleeps for 2 minutes
   - Useful for seeing running workflows in the UI

### Activities

1. **`helloWorldActivity`** - Takes a name and returns a greeting
2. **`longRunningActivity`** - Simulates a long-running task

### Files

- `main.go` - Worker service with workflow/activity implementations
- `long_running_example.go` - Long-running workflow example
- `Dockerfile` - Docker image configuration
- `go.mod`, `go.sum` - Go dependencies

---

## Running Workflows

### Hello World Workflow

```bash
# Docker
docker run --network=samples_cadence-network --rm ubercadence/cli:master \
  --address cadence:7833 \
  --domain test-domain workflow start \
  --tasklist test-worker \
  --workflow_type main.helloWorldWorkflow \
  --execution_timeout 60 \
  --input '"YourName"'

# Local
cadence --domain test-domain workflow start \
  --et 60 --tl test-worker \
  --workflow_type main.helloWorldWorkflow \
  --input '"YourName"'
```

### Long-Running Workflow

```bash
# Docker
docker run --network=samples_cadence-network --rm ubercadence/cli:master \
  --address cadence:7833 \
  --domain test-domain workflow start \
  --tasklist test-worker \
  --workflow_type main.longRunningWorkflow \
  --execution_timeout 300 \
  --input '"Test"'

# Local
cadence --domain test-domain workflow start \
  --et 300 --tl test-worker \
  --workflow_type main.longRunningWorkflow \
  --input '"Test"'
```

---

## Monitoring

### Cadence Web UI

Open your browser and navigate to:
```
http://localhost:8088
```

Enter `test-domain` in the domain field to view:
- Workflow execution history
- Running workflows
- Completed workflows
- Detailed event logs

### Command Line

#### List workflows:
```bash
# Docker
docker run --network=samples_cadence-network --rm ubercadence/cli:master \
  --address cadence:7833 \
  --domain test-domain workflow list

# Local
cadence --domain test-domain workflow list
```

#### Show workflow details:
```bash
# Docker
docker run --network=samples_cadence-network --rm ubercadence/cli:master \
  --address cadence:7833 \
  --domain test-domain workflow show \
  --workflow_id <WORKFLOW_ID>

# Local
cadence --domain test-domain workflow show --workflow_id <WORKFLOW_ID>
```

---

## Configuration

The worker can be configured via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `CADENCE_HOST` | `127.0.0.1:7833` | Cadence server address |
| `CADENCE_DOMAIN` | `test-domain` | Cadence domain name |
| `TASK_LIST` | `test-worker` | Task list name |
| `CLIENT_NAME` | `test-worker` | YARPC client name |

### Example with custom configuration:

```bash
# Local
CADENCE_HOST=localhost:7833 CADENCE_DOMAIN=my-domain ./hello-world

# Docker
docker run -e CADENCE_HOST=cadence:7833 -e CADENCE_DOMAIN=my-domain \
  --network=samples_cadence-network \
  samples_hello-world-worker
```

---

## Building the Docker Image

### Using docker-compose (recommended):
```bash
cd /path/to/cadence/samples
docker-compose build hello-world-worker
```

### Using docker build directly:
```bash
cd /path/to/cadence/samples/hello-world
docker build -t hello-world-sample .
```

### Running the built image:
```bash
docker run --network=samples_cadence-network \
  -e CADENCE_HOST=cadence:7833 \
  hello-world-sample
```

---

## Troubleshooting

### Worker Can't Connect to Cadence

**Docker**: Make sure you're using the correct network and address:
```bash
# Check if Cadence is running
docker-compose ps

# Check worker logs
docker-compose logs hello-world-worker

# Ensure using correct hostname: cadence:7833 (not localhost)
```

**Local**: Ensure Cadence server is running on `127.0.0.1:7833`

### Domain Not Found

Register the domain:
```bash
# Docker
docker run --network=samples_cadence-network --rm ubercadence/cli:master \
  --address cadence:7833 \
  --domain test-domain domain register

# Local
cadence --domain test-domain domain register
```

### Workflow Not Executing

1. Check if worker is running and registered
2. Verify domain is registered
3. Check task list name matches (`test-worker`)
4. View worker logs for errors

---

## What You'll Learn

This sample demonstrates:

- ✅ **Worker Service** - How to create and start a Cadence worker
- ✅ **Workflows** - How to implement durable workflow functions
- ✅ **Activities** - How to implement activity functions
- ✅ **Timeouts** - How to configure activity timeouts
- ✅ **Error Handling** - How to handle activity failures
- ✅ **Logging** - How to use structured logging in workflows
- ✅ **Docker Deployment** - How to containerize Cadence workers

---

## Next Steps

- Explore the [Cadence Documentation](https://cadenceworkflow.io/docs/)
- Check out more [Cadence Samples](https://github.com/uber-common/cadence-samples)
- Learn about [Go Client](https://cadenceworkflow.io/docs/go-client/)
- Join the [Cadence Community](https://cadenceworkflow.io/community/)

---

## License

See the main [LICENSE](../../LICENSE) file.
