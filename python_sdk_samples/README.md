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

---

## Samples

### OpenAI Samples (`openai_samples/`)

LLM-powered workflow samples using the Cadence OpenAI integration.

- `agent_handoffs` — multi-agent handoff pattern
- `human_in_the_loop` — pause a workflow and resume it based on human input

### Schedule Samples (`schedule_samples/`)

Demonstrates the full Cadence Schedules API. A **schedule** is a server-side
cron that fires a workflow on a recurring interval without any client process
needing to stay alive.

**1. Start a worker** (keep running in one terminal):
```bash
uv run python -m schedule_samples.schedule_sample worker
```

**2. Run the demo** (in a second terminal):
```bash
uv run python -m schedule_samples.schedule_sample demo
```

Both subcommands accept `--target` (default `localhost:7833`) and `--domain` (default `default`).

**Expected output:**
```
Created  : schedule-sample-<id>
Describe : cron='* * * * *'  paused=False
Paused   : paused=True  reason='sample demo'
Unpaused : paused=False
Backfill : submitted 2-hour window (HH:MM → HH:MM UTC)
Updated  : new cron='0 * * * *'
List     : schedules in domain:
           schedule-sample-<id> ← this one
Deleted  : schedule-sample-<id>
```

**What each step does:**

| Step | What happens |
|------|-------------|
| **create** | Registers a new schedule that fires `ScheduleSampleWorkflow` every minute (`* * * * *`), with `SkipNew` overlap (skips a fire if the previous run is still active) and `Skip` catch-up (discards missed fires after downtime) |
| **describe** | Fetches the schedule's current spec and state from the server |
| **pause** | Suspends firing — the schedule stays registered but no new workflow runs are started until unpaused |
| **unpause** | Resumes firing with `Skip` catch-up — fires that were missed while paused are discarded |
| **backfill** | Asks the server to replay the last 2 hours of schedule fires immediately, using `Buffer` overlap so they queue rather than skip each other |
| **update** | Changes the cron expression to hourly (`0 * * * *`) using a read-modify-write — the client fetches the current schedule, the callback mutates the spec, and the full updated schedule is sent back |
| **list** | Paginates all schedules in the domain and prints their IDs |
| **delete** | Permanently removes the schedule; any already-started workflow runs continue to completion |
