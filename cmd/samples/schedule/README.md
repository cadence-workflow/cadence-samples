This sample demonstrates the Cadence Schedule SDK — a client-side API for creating and managing schedules that trigger workflow executions on a cron expression without requiring the caller to keep a process running.

`schedule_workflow.go` defines the workflow and activity that the schedule triggers on each cron tick.

`main.go` provides two modes:
- `worker` — long-running process that polls for and executes scheduled workflow runs
- `manage` — walks through the full schedule lifecycle: Create, Describe, Update, Pause, Unpause, Backfill, List, and Delete

## Steps to run this sample

**Step 1 — Start a Cadence server**

Start a local Cadence server using Docker:

```bash
cd ~/src/cadence
docker-compose up
```

The server exposes two ports used by this sample:
- `localhost:7933` — TChannel (used by the Cadence CLI)
- `localhost:7833` — gRPC (used by the Go client)

The `cadence-samples` domain is registered automatically when the binary starts — no manual CLI step needed.

**Step 2 — Build the schedule binary**

```bash
cd ~/src/cadence-samples
make schedule
```

This compiles `cmd/samples/schedule/*.go` into `bin/schedule`.

**Step 3 — Start the worker (Terminal 1)**

```bash
cd ~/src/cadence-samples
./bin/schedule -m worker
```

Leave this running. The worker polls the `scheduleGroup` task list and executes any workflow runs triggered by the schedule. You should see:

```
Started Workflow Worker   {"TaskList": "scheduleGroup", ...}
Started Activity Worker   {"TaskList": "scheduleGroup", ...}
```

**Step 4 — Run manage mode (Terminal 2)**

```bash
cd ~/src/cadence-samples
./bin/schedule -m manage
```

This walks through the full schedule lifecycle in sequence:

```
=== Step 1: Create schedule ===
=== Step 2: Describe schedule ===
=== Step 3: Update schedule (change cron to every 2 hours) ===
=== Step 4: Pause schedule ===
=== Step 5: Unpause schedule ===
=== Step 6: Backfill past 3 hours (triggers one run) ===
=== Step 7: List schedules in domain ===
=== Step 8: Delete schedule ===
```

When Step 6 (Backfill) runs, switch back to Terminal 1 — you will see the worker pick up and execute a real workflow run. The log lines include extra context fields (WorkflowID, RunID, etc.) and look like:

```
Scheduled workflow started  
Scheduled activity executed
Scheduled workflow completed
```

**Step 5 — Stop the worker**

Press `CMD+C` in Terminal 1.
