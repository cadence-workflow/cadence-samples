# Schedule Samples

Demonstrates the full Cadence Schedules API. A **schedule** is a server-side
cron that fires a workflow on a recurring interval without any client process
needing to stay alive.

Each operation is its own runnable script. All scripts share the schedule ID
`"my-cadence-schedule"` defined in `workflow.py` and accept `--target`
(default `localhost:7833`) and `--domain` (default `default`).

## Running the samples

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

## Expected output

**`create_schedule`**
```
Created schedule 'my-cadence-schedule' (fires every minute)
```
If the schedule already exists:
```
Schedule 'my-cadence-schedule' already exists. Run delete_schedule first or use --schedule-id to pick a different name.
```

**`describe_schedule`** (run shortly after create, once the first cron tick fires):
```
Schedule : 'my-cadence-schedule'
  cron           : '* * * * *'
  paused         : False
  overlap policy : SKIP_NEW
  pause_on_fail  : True
  next run       : 2026-01-01 09:01:00 UTC
  last run       : 2026-01-01 09:00:00 UTC
  total runs     : 1
  memo           : {'owner': 'platform-team', 'env': 'dev'}
```

**`pause_schedule`**
```
Paused schedule 'my-cadence-schedule' (reason: 'paused via sample')
```

**`describe_schedule`** (after pause — note `paused: True`, `next run` absent):
```
Schedule : 'my-cadence-schedule'
  cron           : '* * * * *'
  paused         : True
  pause reason   : 'paused via sample'
  overlap policy : SKIP_NEW
  pause_on_fail  : True
  last run       : 2026-01-01 09:00:00 UTC
  total runs     : 1
  memo           : {'owner': 'platform-team', 'env': 'dev'}
```

**`unpause_schedule`**
```
Unpaused schedule 'my-cadence-schedule' (reason: 'resuming via sample')
```

**`backfill_schedule`** — replays the last 2 hours of missed fires immediately using BUFFER overlap:
```
Backfilled schedule 'my-cadence-schedule' (07:00 → 09:00 UTC, 2h window)
```

**`update_schedule`** — changes the cron to every 2 minutes:
```
Updated schedule 'my-cadence-schedule' cron='*/2 * * * *'
```

**`list_schedules`**
```
Schedules in domain 'default':
  my-cadence-schedule
```

**`delete_schedule`**
```
Deleted schedule 'my-cadence-schedule'
```

## What each script does

| Script | What happens |
|--------|-------------|
| `workflow.py` | Shared workflow definition and constants (`SCHEDULE_ID`, `TASK_LIST`) |
| `run_worker.py` | Starts a worker that executes `ScheduleSampleWorkflow` |
| `create_schedule.py` | Registers a schedule firing every minute (`* * * * *`), with `SkipNew` overlap, `Skip` catch-up, `pause_on_failure`, and a schedule-level memo. Accepts `--schedule-id` to create under a different name. |
| `describe_schedule.py` | Prints spec, state, policies, next/last run times, total runs, and memo |
| `pause_schedule.py` | Suspends firing; the schedule stays registered but no new runs start |
| `unpause_schedule.py` | Resumes firing with `Skip` catch-up; missed fires while paused are discarded |
| `backfill_schedule.py` | Submits the last 2 hours of schedule fires immediately using `Buffer` overlap |
| `update_schedule.py` | Changes the cron to every 2 minutes (`*/2 * * * *`) via a read-modify-write callback |
| `list_schedules.py` | Paginates and prints all schedule IDs in the domain |
| `delete_schedule.py` | Removes the schedule; running workflows complete normally |
