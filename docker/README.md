# Cadence Samples with Docker Compose

This project provides a complete Cadence development environment using Docker Compose, including the core Cadence services and cadence-samples application.

## Prerequisites

- Docker and Docker Compose installed
- Basic understanding of Cadence workflows

## Project Structure

```
.
├── docker/
│   └── Dockerfile          # Your cadence-samples Dockerfile
├── docker-compose.yml      # Complete Cadence stack + samples
├── config/                 # Configuration files
├── bin/                    # Built binaries (after build)
└── README.md              # This file
```

## Getting Started

### Step 1: Start the Complete Stack

```bash
# Start all services (Cadence, Cassandra, Web UI, Grafana, and cadence-samples)
docker-compose up -d

# Check if all services are running
docker-compose ps

# View logs to ensure everything started correctly
docker-compose logs cadence
```

### Step 2: Wait for Services to be Ready

Wait for Cadence to be fully initialized:

```bash
# Monitor Cadence logs until you see "Started"
docker-compose logs -f cadence
```

### Step 3: Access the Cadence Web UI

Open your browser and navigate to:
- **Cadence Web UI**: http://localhost:8088
- **Grafana**: http://localhost:3000
- **Prometheus**: http://localhost:9090

## Using the Samples

### Step 1: Access the Container

```bash
docker-compose exec cadence-samples /bin/bash
```

### Step 2: Run Workflow Examples

You need **two terminals** for most examples - one for the worker and another for triggering workflows.

#### Terminal 1 - Start the Worker
```bash
# Access the container
docker-compose exec cadence-samples /bin/bash

# Example: Hello World worker
./bin/helloworld -m worker
```

#### Terminal 2 - Trigger the Workflow
Open a second terminal and execute:
```bash
# Access the container in a new session
docker-compose exec cadence-samples /bin/bash

# Trigger the workflow
./bin/helloworld -m trigger
```

#### Stop the Worker
In Terminal 1, press `Ctrl+C` to stop the worker.

### Available Sample Commands

Once inside the container, you can run various sample workflows:

```bash
# List available binaries
ls -la ./bin/

# Examples (replace with actual sample names)
./bin/helloworld -m worker      # Start worker
./bin/helloworld -m trigger     # Trigger workflow

./bin/timer -m worker           # Timer example worker
./bin/timer -m trigger          # Timer example trigger
```

## Updating the Docker Compose

### Option 1: Update from Original Source

To update the base Cadence services:

1. Download the latest version:
```bash
curl -o docker-compose-base.yml https://raw.githubusercontent.com/cadence-workflow/cadence/refs/heads/master/docker/docker-compose.yml
```

2. Manually merge the `cadence-samples` service from the current `docker-compose.yml`

3. Test the updated configuration

### Option 2: Modify the Current Configuration

To update your cadence-samples service:

1. **Change the Docker image**: Modify the `build` section in `docker-compose.yml`
2. **Update environment variables**: Add or modify variables in the `environment` section
3. **Rebuild the service**: `docker-compose build cadence-samples`
4. **Restart the service**: `docker-compose up -d cadence-samples`

### Example: Adding a New Environment Variable

```yaml
cadence-samples:
  # ... existing configuration
  environment:
    - CADENCE_HOST=cadence:7833
    - YOUR_NEW_VAR=your_value  # Add this line
```

## Troubleshooting

### Service Not Starting

```bash
# Check service status
docker-compose ps

# View logs for specific service
docker-compose logs cadence-samples
docker-compose logs cadence

# Restart specific service
docker-compose restart cadence-samples
```

### Connection Issues

```bash
# Test connectivity from samples to Cadence
docker-compose exec cadence-samples ping cadence

# Check if Cadence ports are accessible
docker-compose exec cadence-samples telnet cadence 7833
```

### Rebuilding After Code Changes

```bash
# Rebuild cadence-samples after code changes
docker-compose build cadence-samples

# Restart with new build
docker-compose up -d cadence-samples
```

## Development Workflow

1. **Make code changes** in your local files
2. **Rebuild the service**: `docker-compose build cadence-samples`
3. **Restart the container**: `docker-compose up -d cadence-samples`
4. **Test your changes** using the workflow examples

## Stopping the Environment

```bash
# Stop all services
docker-compose down

# Stop and remove volumes (WARNING: This will delete data)
docker-compose down -v

# Stop and remove everything including images
docker-compose down --rmi all -v
```

## Service Ports

- **Cadence Frontend**: 7833
- **Cadence Web UI**: 8088
- **Grafana**: 3000
- **Prometheus**: 9090
- **Cassandra**: 9042

## Notes

- The cadence-samples container runs in interactive mode by default
- Configuration files are automatically updated to point to the correct Cadence host
- All services use Docker internal networking for communication
- Data persists in Docker volumes between restarts (unless explicitly removed)