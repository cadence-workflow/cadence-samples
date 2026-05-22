# Cadence Python SDK Samples

All samples under this folder demonstrate how to use Python SDK effectively.

## 🚀 Quick Start

1. We use uv to install dependencies of all samples

Refer to [UV installation Guide](https://docs.astral.sh/uv/getting-started/installation/)

2. build all samples
```bash
cd python_sdk_samples
uv sync
```

This downloads all dependencies so `uv run` will have all the dependent packages

3. Start Cadence Server

```bash
curl -LO https://raw.githubusercontent.com/cadence-workflow/cadence/refs/heads/master/docker/docker-compose.yml && docker-compose up --wait
```

This downloads and starts all required dependencies including Cadence server, database, and [Cadence Web UI](https://github.com/uber/cadence-web). You can view your sample workflows at [http://localhost:8088](http://localhost:8088).

4. **run one sample**:

```bash
uv run python -m openai_samples.agent_handoffs.main
```
