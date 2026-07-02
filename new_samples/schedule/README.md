This sample demonstrates the Cadence **Schedule** SDK — a client-side API for creating and managing schedules that trigger workflow executions on a cron expression.

Each operation is its own standalone command, matching the [Python schedule samples](../../python_sdk_samples/schedule_samples/README.md) so you can compare Go and Python side by side.

## What's here

| File | Operation |
|------|-----------|
| `workflow.go` | Shared workflow definition and constants (`ScheduleID`, `TaskListName`) |
| `worker.go` | Start the workflow worker |
| `create_schedule.go` | Create a schedule with cron, overlap/catch-up policies, `pause_on_failure`, and memo |
| `describe_schedule.go` | Print spec, state, policies, next/last run times, and total runs |
| `pause_schedule.go` | Pause with a reason |
| `unpause_schedule.go` | Unpause with `SKIP` catch-up |
| `backfill_schedule.go` | Replay a historical time window (last 2 hours) |
| `update_schedule.go` | Read-modify-write the cron expression |
| `list_schedules.go` | Paginate all schedules in the domain |
| `delete_schedule.go` | Delete the schedule |

## Steps to run

**Step 1 — Start a Cadence server**

```bash
cd <path-to-cadence-repo>
docker-compose up
```

Ports used by this sample:
- `localhost:7933` — TChannel (Cadence CLI)
- `localhost:7833` — gRPC (Go client)

**Step 2 — Register the domain (one-time)**

```bash
./cadence --address localhost:7933 --do cadence-samples domain register
```

**Step 3 — Start the worker (Terminal 1)**

```bash
cd new_samples/schedule
go run . -m worker
```

Leave it running; it polls the `schedule-sample-worker` task list.

**Step 4 — Run operations in order (Terminal 2)**

```bash
cd new_samples/schedule

go run . -m manage -op create       # Create the schedule
go run . -m manage -op describe     # Inspect the schedule
go run . -m manage -op pause        # Pause it
go run . -m manage -op describe     # Confirm paused
go run . -m manage -op unpause      # Resume it
go run . -m manage -op backfill     # Replay last 2 hours
go run . -m manage -op update       # Change cron to every 2 minutes
go run . -m manage -op describe     # Confirm updated cron
go run . -m manage -op list         # List all schedules
go run . -m manage -op delete       # Delete the schedule
```

**Step 5 — Stop the worker**

Press `Ctrl+C` in Terminal 1.
