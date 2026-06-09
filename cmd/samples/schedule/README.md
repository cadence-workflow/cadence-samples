This sample demonstrates the Cadence **Schedule** SDK — a client-side API for creating and managing schedules that trigger workflow executions on a cron expression without requiring the caller to keep a process running.

## What's here

- `schedule_workflow.go` — the workflow and activity the schedule triggers on each tick. The workflow takes a `sleepSeconds` input so the same code can serve fast runs (lifecycle/catch-up/pagination) and long-running runs (overlap).
- `main.go` — two modes: `worker` (polls and executes runs) and `manage` (drives the API). `manage` selects a **scenario** with `-scenario`.
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

### Notes on what to expect

- **Memo model:** both schedule-level `Memo` and action-level `Memo` take native Go values on write — the SDK encodes them with your `DataConverter`. On `Describe` they come back as **raw bytes** (`map[string][]byte`) which you decode yourself (`SearchAttributes` are JSON). The `dataconverter` scenario demonstrates the full round-trip for both levels; `lifecycle` just logs their presence.
- **Server gaps to expect** (the `lifecycle` scenario logs these as zero/nil — they are server-side gaps, not sample failures): `Info.CreateTime`/`LastUpdateTime` are zero and `Info.OngoingBackfills` is nil.
- **Describe-then-update:** `Update` takes a callback. The SDK calls `DescribeSchedule` for you, pre-populates `*client.ScheduleUpdate` with the current state, runs your callback, then diffs and sends only the fields you changed. In `lifecycle`, the Update flips one policy sub-field (`OverlapPolicy`) and the siblings (`PauseOnFailure`, `BufferLimit`, `ConcurrencyLimit`) are **preserved** — the SDK reads them from Describe and resends them, so the server's within-struct full replacement does not reset them. Untouched top-level fields (the `Action`) are omitted and preserved by the server's top-level merge. A second Update changes only the action-level `Memo` via `SetActionMemo`.
- **Behavioral scenarios** (`overlap`, `catchup`) depend on the cluster's scheduler worker. Run them against a current Cadence build; the precise outcome is best observed in the worker logs / Cadence Web UI (the sample logs best-effort counts via `Info.TotalRuns`).

## Steps to run

**Step 1 — Start a Cadence server**

```bash
cd ~/src/cadence
docker-compose up
```

Ports used by this sample:
- `localhost:7933` — TChannel (Cadence CLI)
- `localhost:7833` — gRPC (Go client)

The `cadence-samples` domain is registered automatically when the binary starts.

**Step 2 — Build**

```bash
cd ~/src/cadence-samples
make schedule          # produces bin/schedule
```

> This sample's `go.mod` currently `replace`s `go.uber.org/cadence` with a local checkout of
> the Go client and pins the matching `cadence-idl`. Remove the `replace` once the SDK
> changes are released.

**Step 3 — Start the worker (Terminal 1)**

```bash
./bin/schedule -m worker
```

Leave it running; it polls the `scheduleGroup` task list. Required for `lifecycle`, `overlap`, and `catchup`.

**Step 4 — Run a scenario (Terminal 2)**

```bash
./bin/schedule -m manage -scenario lifecycle      # default
./bin/schedule -m manage -scenario overlap
./bin/schedule -m manage -scenario catchup
./bin/schedule -m manage -scenario pagination
./bin/schedule -m manage -scenario dataconverter
```

While a behavioral scenario runs, watch Terminal 1 (and the Cadence Web UI) to see the worker pick up and execute real workflow runs:

```
Scheduled workflow started
Scheduled activity executed
Scheduled workflow completed
```

**Step 5 — Stop the worker**

Press `CMD+C` in Terminal 1.
