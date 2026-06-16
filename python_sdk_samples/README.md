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

Each operation is its own runnable script. All scripts share the schedule ID
`"my-cadence-schedule"` defined in `workflow.py` and accept `--target`
(default `localhost:7833`) and `--domain` (default `default`).

**1. Start a worker** (keep running in one terminal):
```bash
uv run python -m schedule_samples.run_worker
```

**2. Run each operation** (in a second terminal, in order):
```bash
uv run python -m schedule_samples.create_schedule
uv run python -m schedule_samples.describe_schedule
uv run python -m schedule_samples.pause_schedule
uv run python -m schedule_samples.describe_schedule
uv run python -m schedule_samples.unpause_schedule
uv run python -m schedule_samples.backfill_schedule
uv run python -m schedule_samples.update_schedule
uv run python -m schedule_samples.list_schedules
uv run python -m schedule_samples.delete_schedule
```

**What each script does:**

| Script | What happens |
|--------|-------------|
| `workflow.py` | Shared workflow definition and constants (`SCHEDULE_ID`, `TASK_LIST`) |
| `run_worker.py` | Starts a worker that executes `ScheduleSampleWorkflow` |
| `create_schedule.py` | Registers a schedule firing every minute (`* * * * *`), with `SkipNew` overlap and `Skip` catch-up |
| `describe_schedule.py` | Fetches and prints the schedule's current spec and state |
| `pause_schedule.py` | Suspends firing — the schedule stays registered but no new runs are started |
| `unpause_schedule.py` | Resumes firing with `Skip` catch-up — missed fires while paused are discarded |
| `backfill_schedule.py` | Replays the last 2 hours of schedule fires immediately using `Buffer` overlap |
| `update_schedule.py` | Changes the cron to hourly (`0 * * * *`) via a read-modify-write callback |
| `list_schedules.py` | Paginates and prints all schedule IDs in the domain |
| `delete_schedule.py` | Permanently removes the schedule; running workflows complete normally |
