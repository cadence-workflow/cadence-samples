This sample demonstrates the Cadence **Schedule** SDK — a client-side API for creating and managing schedules that trigger workflow executions on a cron expression without requiring the caller to keep a process running.

## What's here

- `workflow.go` — the workflow and activity the schedule triggers on each tick. The workflow takes a `sleepSeconds` input so the same code can serve fast runs (lifecycle/pagination) and long-running runs (overlap/concurrency).
- `main.go` — two modes: `worker` (polls and executes runs) and `manage` (drives the API). `manage` selects a **scenario** with `-scenario`.
- `worker.go` — Cadence worker setup and client builder.
- `helpers.go` — shared helpers (client build, input encoding, cleanup).
- `scenario_*.go` — one file per scenario (see below).

## Scenarios (`-m manage -scenario <name>`)

- **`lifecycle`** *(default)* — Full-field **Create→Describe** round-trip, **Create non-idempotency**, **describe-then-update Update** (change cron + one policy sub-field with siblings preserved, then action Memo via `SetActionMemo`), **Pause** (reason) / **Unpause**, **List** entry fields, **Delete** + confirm absent from List. Requires the worker.

- **`backfill`** — **Backfill** a past time range under three overlap policies: SkipNew (only the first slot starts; the rest see an open run and are skipped), Concurrent (all slots start in parallel), CancelPrevious (each slot cancels its predecessor). Also verifies **idempotency** via `BackfillID` — submitting the same range twice does not re-queue it. Requires the worker.

- **`overlap`** — Live demonstration of `ScheduleOverlapPolicy`: SkipNew, Concurrent, and CancelPrevious, each with runs that outlast the cron interval so overlaps actually occur. Observational — watch the worker logs or Cadence Web UI to compare policy behavior. For automated overlap verification, see `backfill`. Requires the worker.

- **`concurrency`** — Verifies that `ConcurrencyLimit` is a hard server-side bound. A schedule fires every 3 seconds with `ConcurrencyLimit=2`; each workflow runs for 20 seconds. After a 12-second window no run has completed, so `TotalRuns` equals the number of simultaneous starts — it must be exactly 2. Requires the worker.

- **`catchup`** — Catch-up behavior on **Unpause**: `Skip` (missed fires discarded) vs `All` (every missed fire replayed within the catch-up window). Requires the worker.

- **`pagination`** — `List` paging via `NextPageToken` across three boundary cases: multi-page (5 schedules, pageSize=2), single-page (3 schedules, pageSize=10), and exact-boundary (4 schedules, pageSize=2). Each case asserts no duplicates, no gaps, and that every page respects the page-size limit. No worker needed.

- **`dataconverter`** — **Memo** encoded on write / decoded on read through a **custom** `DataConverter` (gob); also shows the default JSON converter cannot decode the gob bytes. Note: schedule-level Memo is not returned by the server in Describe responses (server gap); only action-level Memo is verified. No worker needed.

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

Leave it running; it polls the `schedule-sample-worker` task list. Required for `lifecycle`, `backfill`, `overlap`, `concurrency`, and `catchup`.

**Step 4 — Run a scenario (Terminal 2)**

```bash
cd new_samples/schedule
go run . -m manage -scenario lifecycle      # default
go run . -m manage -scenario backfill
go run . -m manage -scenario overlap
go run . -m manage -scenario concurrency
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
