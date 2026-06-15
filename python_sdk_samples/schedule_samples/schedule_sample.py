"""Cadence Schedules sample.

Walks through the full schedule lifecycle against a running Cadence server:

    create → describe → pause → unpause → backfill → update → list → delete

Prerequisites:
    1. A running Cadence server (see python_sdk_samples/README.md).
    2. A worker running for the target task list. Start one in a separate
       terminal before running this script:
           uv run python -m schedule_samples.schedule_sample worker

Usage (in a second terminal):
    uv run python -m schedule_samples.schedule_sample demo

Both subcommands accept --target and --domain flags.
"""

import argparse
import asyncio
import sys
import uuid
from datetime import datetime, timedelta, timezone

from google.protobuf.duration import from_timedelta

from cadence import Registry, workflow
from cadence.client import Client
from cadence.worker import Worker
from cadence.api.v1 import common_pb2, schedule_pb2, tasklist_pb2

TASK_LIST = "schedule-sample-tl"
WORKFLOW_TYPE = "ScheduleSampleWorkflow"

registry = Registry()


@registry.workflow()
class ScheduleSampleWorkflow:
    """No-op workflow started by the schedule on each trigger."""

    @workflow.run
    async def run(self) -> str:
        return "ok"


# ---------------------------------------------------------------------------
# worker subcommand
# ---------------------------------------------------------------------------


async def run_worker(args: argparse.Namespace) -> None:
    async with Client(domain=args.domain, target=args.target) as client:
        print(f"Worker running on task list {TASK_LIST!r} — Ctrl-C to stop")
        async with Worker(client, TASK_LIST, registry):
            await asyncio.Event().wait()


# ---------------------------------------------------------------------------
# demo subcommand
# ---------------------------------------------------------------------------


async def run_demo(args: argparse.Namespace) -> None:
    sid = f"schedule-sample-{uuid.uuid4().hex[:8]}"

    action = schedule_pb2.ScheduleAction(
        start_workflow=schedule_pb2.ScheduleAction.StartWorkflowAction(
            workflow_type=common_pb2.WorkflowType(name=WORKFLOW_TYPE),
            task_list=tasklist_pb2.TaskList(name=TASK_LIST),
            execution_start_to_close_timeout=from_timedelta(timedelta(minutes=10)),
            task_start_to_close_timeout=from_timedelta(timedelta(seconds=10)),
        )
    )
    spec = schedule_pb2.ScheduleSpec(cron_expression="* * * * *")
    policies = schedule_pb2.SchedulePolicies(
        overlap_policy=schedule_pb2.SCHEDULE_OVERLAP_POLICY_SKIP_NEW,
        catch_up_policy=schedule_pb2.SCHEDULE_CATCH_UP_POLICY_SKIP,
    )

    async with Client(domain=args.domain, target=args.target) as client:
        try:
            # --- create ---
            await client.create_schedule(
                sid, spec=spec, action=action, policies=policies
            )
            print(f"Created  : {sid}")

            # --- describe ---
            desc = await client.describe_schedule(sid)
            print(
                f"Describe : cron={desc.spec.cron_expression!r}  paused={desc.state.paused}"
            )

            # --- pause ---
            await client.pause_schedule(sid, reason="sample demo")
            desc = await client.describe_schedule(sid)
            print(
                f"Paused   : paused={desc.state.paused}  reason={desc.state.pause_info.reason!r}"
            )

            # --- unpause ---
            await client.unpause_schedule(
                sid,
                reason="resuming",
                catch_up_policy=schedule_pb2.SCHEDULE_CATCH_UP_POLICY_SKIP,
            )
            desc = await client.describe_schedule(sid)
            print(f"Unpaused : paused={desc.state.paused}")

            # --- backfill (replays the per-minute schedule over the last 2 hours) ---
            end = datetime.now(timezone.utc)
            start = end - timedelta(hours=2)
            await client.backfill_schedule(
                sid,
                start_time=start,
                end_time=end,
                overlap_policy=schedule_pb2.SCHEDULE_OVERLAP_POLICY_BUFFER,
            )
            print(
                f"Backfill : submitted 2-hour window ({start:%H:%M} → {end:%H:%M} UTC)"
            )

            # --- update (change cron) ---
            def set_hourly(d: object) -> None:
                d.spec.cron_expression = "0 * * * *"

            await client.update_schedule(sid, set_hourly)
            desc = await client.describe_schedule(sid)
            print(f"Updated  : new cron={desc.spec.cron_expression!r}")

            # --- list ---
            print("List     : schedules in domain:")
            async for entry in client.list_schedules():
                marker = " ← this one" if entry.schedule_id == sid else ""
                print(f"           {entry.schedule_id}{marker}")

            # --- delete ---
            await client.delete_schedule(sid)
            print(f"Deleted  : {sid}")
        except Exception:
            try:
                await client.delete_schedule(sid)
            except Exception as cleanup_err:
                print(f"Cleanup failed for {sid}: {cleanup_err}")
            raise


# ---------------------------------------------------------------------------
# CLI
# ---------------------------------------------------------------------------


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Cadence Schedules sample")
    p.add_argument(
        "--target", default="localhost:7833", help="Cadence frontend host:port"
    )
    p.add_argument("--domain", default="default", help="Cadence domain")

    sub = p.add_subparsers(dest="cmd", required=True)
    sub.add_parser("worker", help="Run the workflow worker (keep running)")
    sub.add_parser("demo", help="Run the schedule lifecycle demo")

    return p


def main() -> int:
    args = build_parser().parse_args()
    try:
        if args.cmd == "worker":
            asyncio.run(run_worker(args))
        else:
            asyncio.run(run_demo(args))
    except KeyboardInterrupt:
        pass
    return 0


if __name__ == "__main__":
    sys.exit(main())
