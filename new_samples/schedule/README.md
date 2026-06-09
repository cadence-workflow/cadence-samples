This sample demonstrates the Cadence **Schedule** SDK — a client-side API for creating and managing schedules that trigger workflow executions on a cron expression without requiring the caller to keep a process running.

## What's here

- `workflow.go` — the workflow and activity the schedule triggers on each tick. The workflow takes a `sleepSeconds` input so the same code can serve fast runs (lifecycle/catch-up/pagination) and long-running runs (overlap).
- `main.go` — two modes: `worker` (polls and executes runs) and `manage` (drives the API). `manage` selects a **scenario** with `-scenario`.
- `worker.go` — Cadence worker setup and client builder.
- `helpers.go` — shared helpers (client build, input encoding, cleanup).
- `scenario_*.go` — one file per scenario (see below).

## Scenarios (`-m manage -scenario <name>`)

| Scenario | What it exercises | Needs worker? |
|---|---|---|
| `lifecycle` (default) | Full-field **Create→Describe** round-trip, **Create non-idempotency**, **describe-then-update Update** (change cron + one policy sub-field with siblings preserved, then action Memo via `SetActionMemo`), **Pause** (reason) / **Unpause**, **Backfill** a past range, **List** entry fields, **Delete** + confirm absent from List | yes (for backfill runs) |
| `overlap` | `ScheduleOverlapPolicy` — SkipNew / Concurrent / CancelPrevious, with runs that outlast the cron interval so overlaps actually occur | yes |
| `catchup` | Catch-up on **Unpause** — `Skip` vs `All` over missed fire times | yes |
| `pagination` | `List` paging through all schedules via `NextPageToken` (pageSize=2), asserting each appears exactly once | no |
| `dataconverter` | **Memo** (schedule-level *and* action-level) encoded on write / decoded on read through a **custom** `DataConverter` (gob); shows the default JSON converter cannot decode it | no |

## Steps to run

**Step 1 — Start a Cadence server**

```bash
cd ~/src/cadence
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

Leave it running; it polls the `schedule-sample-worker` task list. Required for `lifecycle`, `overlap`, and `catchup`.

**Step 4 — Run a scenario (Terminal 2)**

```bash
cd new_samples/schedule
go run . -m manage -scenario lifecycle      # default
go run . -m manage -scenario overlap
go run . -m manage -scenario catchup
go run . -m manage -scenario pagination
go run . -m manage -scenario dataconverter
```

While a behavioral scenario runs, watch Terminal 1 (and the Cadence Web UI) to see the worker pick up and execute real workflow runs:

```
Scheduled workflow started
Scheduled activity executed
Scheduled workflow completed
```

**Step 5 — Stop the worker**

Press `CMD+C` in Terminal 1.
