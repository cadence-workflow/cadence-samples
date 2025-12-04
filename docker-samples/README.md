# Cadence Docker Samples

This directory contains Dockerized Cadence workflow samples that are ready to run with a single command. These samples are designed for easy onboarding and demonstration purposes.

## 🚀 Quick Start

Get started in under 2 minutes!

### Using the Quick Start Script

```bash
cd docker-samples
./quick-start.sh
```

This script will:
1. Start Cadence server and dependencies
2. Register the test domain
3. Start the hello-world worker
4. Trigger a sample workflow
5. Show you the results

### Manual Start

```bash
cd docker-samples
docker-compose up
```

Then visit:
- **Cadence Web UI**: http://localhost:8088 (domain: `test-domain`)
- **Worker Logs**: `docker-compose logs -f hello-world-worker`

## 📁 Available Samples

### [Hello World](hello-world/)

A simple, standalone Cadence workflow sample that demonstrates:
- Basic workflow and activity execution
- Environment-based configuration
- Docker containerization
- Multi-stage Docker builds

**Features:**
- ✅ Zero dependencies (runs completely in Docker)
- ✅ Environment variable configuration
- ✅ Production-ready Docker setup
- ✅ Comprehensive documentation

See [hello-world/README.md](hello-world/README.md) for detailed instructions.

## 🏗️ Architecture

The Docker setup includes:

- **Cassandra** (4.1.1) - Persistence layer
- **Cadence Server** (master-auto-setup) - Workflow engine
- **Cadence Web UI** (latest) - Web interface on port 8088
- **Hello World Worker** - Sample worker implementation

All services communicate via a custom Docker network (`docker-samples_cadence-network`).

## 📚 What's Included

```
docker-samples/
├── docker-compose.yml          # Orchestrates all services
├── quick-start.sh              # One-command demo
├── list-workflows.sh           # List all workflows
├── README.md                   # This file
├── DOCKERIZATION_COMPLETE.md   # Implementation details
└── hello-world/
    ├── main.go                 # Worker with env var support
    ├── long_running_example.go # Long-running workflow example
    ├── Dockerfile              # Multi-stage build
    ├── .dockerignore           # Build optimization
    ├── go.mod                  # Go dependencies
    ├── go.sum
    └── README.md               # Detailed sample docs
```

## 🎯 Use Cases

These Docker samples are perfect for:

- **Quick Demos**: Show Cadence capabilities in minutes
- **Onboarding**: Help new team members get started
- **Development**: Test workflows locally without complex setup
- **Learning**: Understand Cadence concepts through working examples
- **Prototyping**: Quickly experiment with workflow patterns

## ⚙️ Configuration

All samples support environment variable configuration for flexibility:

| Variable | Default | Description |
|----------|---------|-------------|
| `CADENCE_HOST` | `cadence:7833` | Cadence server address |
| `CADENCE_DOMAIN` | `test-domain` | Workflow domain |
| `TASK_LIST` | `test-worker` | Task list name |
| `CLIENT_NAME` | `test-worker` | YARPC client name |

## 🔧 Common Commands

### Start Services
```bash
docker-compose up -d
```

### Stop Services
```bash
docker-compose down
```

### View Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f hello-world-worker
docker-compose logs -f cadence
```

### Rebuild Worker
```bash
docker-compose build hello-world-worker
docker-compose up -d hello-world-worker
```

### Register Domain
```bash
docker run --network=docker-samples_cadence-network --rm ubercadence/cli:master \
  --address cadence:7833 \
  --domain test-domain domain register
```

### Trigger Workflow
```bash
docker run --network=docker-samples_cadence-network --rm ubercadence/cli:master \
  --address cadence:7833 \
  --domain test-domain workflow start \
  --tasklist test-worker \
  --workflow_type main.helloWorldWorkflow \
  --execution_timeout 60 \
  --input '"World"'
```

### List Workflows
```bash
./list-workflows.sh
# OR
docker run --network=docker-samples_cadence-network --rm ubercadence/cli:master \
  --address cadence:7833 \
  --domain test-domain workflow list
```

## 🛠️ Troubleshooting

### Services Won't Start

Check if ports are already in use:
```bash
lsof -i :7833  # Cadence
lsof -i :8088  # Web UI
lsof -i :9042  # Cassandra
```

### Worker Can't Connect

1. Check if Cadence is healthy:
   ```bash
   docker-compose ps
   docker-compose logs cadence
   ```

2. Verify network connectivity:
   ```bash
   docker network inspect docker-samples_cadence-network
   ```

3. Ensure correct hostname (`cadence:7833`, not `localhost:7833`)

### Domain Not Found

Register it manually:
```bash
docker run --network=docker-samples_cadence-network --rm ubercadence/cli:master \
  --address cadence:7833 \
  --domain test-domain domain register
```

### Clean Restart

```bash
docker-compose down -v
docker-compose up --build
```

## 📖 Learn More

- [Cadence Documentation](https://cadenceworkflow.io/docs/)
- [Cadence Go Client](https://github.com/uber/cadence-go-client)
- [Main Samples Repository](https://github.com/uber-common/cadence-samples)
- [Cadence Server](https://github.com/uber/cadence)

## 🤝 Contributing

Want to add more Docker samples? Check our [Contributing Guide](../CONTRIBUTING.md).

## 📄 License

Apache 2.0 - See [LICENSE](../LICENSE) for details.

---

**Ready to get started? Run `./quick-start.sh` and see Cadence in action!** 🚀

