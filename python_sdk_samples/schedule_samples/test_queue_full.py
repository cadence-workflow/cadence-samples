"""Test backfill queue-full rejection: >maxPendingBackfills=10 signals are dropped.

Sends 11 rapid backfill signals to a schedule. The scheduler workflow accepts
the first 10 (maxPendingBackfills) and silently rejects the 11th, emitting a
scheduler_backfill_rejected_count_per_domain metric tagged reason=queue_full.

After sending the signals, describe_schedule is called to show that only up
to 10 backfills are pending.

Usage:
    uv run python -m schedule_samples.test_queue_full [--cleanup]
"""

import argparse
import asyncio
from datetime import datetime, timedelta, timezone

from google.protobuf.duration import from_timedelta

from cadence.api.v1 import common_pb2, schedule_pb2, tasklist_pb2
from cadence.client import Client

from schedule_samples.workflow import SCHEDULE_ID, TASK_LIST, WORKFLOW_TYPE

QUEUE_FULL_SCHEDULE_ID = "queue-full-test-schedule"
SIGNALS_TO_SEND = 11  # one more than maxPendingBackfills


async def _delete_if_exists(client: Client, schedule_id: str) -> None:
    try:
        await client.delete_schedule(schedule_id)
        print(f"Deleted existing schedule {schedule_id!r}")
    except Exception:
        pass


async def main(args: argparse.Namespace) -> None:
    async with Client(domain=args.domain, target=args.target) as client:
        if args.cleanup:
            await _delete_if_exists(client, QUEUE_FULL_SCHEDULE_ID)
            return

        await _delete_if_exists(client, QUEUE_FULL_SCHEDULE_ID)

        # Create a schedule that's paused so the scheduler doesn't drain backfills
        # between signals, giving us a deterministic queue-full scenario.
        await client.create_schedule(
            QUEUE_FULL_SCHEDULE_ID,
            spec=schedule_pb2.ScheduleSpec(cron_expression="* * * * *"),
            action=schedule_pb2.ScheduleAction(
                start_workflow=schedule_pb2.ScheduleAction.StartWorkflowAction(
                    workflow_type=common_pb2.WorkflowType(name=WORKFLOW_TYPE),
                    task_list=tasklist_pb2.TaskList(name=TASK_LIST),
                    workflow_id_prefix=f"{QUEUE_FULL_SCHEDULE_ID}-",
                    execution_start_to_close_timeout=from_timedelta(timedelta(minutes=10)),
                    task_start_to_close_timeout=from_timedelta(timedelta(seconds=10)),
                )
            ),
            policies=schedule_pb2.SchedulePolicies(
                overlap_policy=schedule_pb2.SCHEDULE_OVERLAP_POLICY_BUFFER,
                catch_up_policy=schedule_pb2.SCHEDULE_CATCH_UP_POLICY_ALL,
            ),
        )
        print(f"Created schedule {QUEUE_FULL_SCHEDULE_ID!r}")

        # Pause so the scheduler sits idle while we send all 11 signals.
        await client.pause_schedule(QUEUE_FULL_SCHEDULE_ID, reason="queue-full test")
        print("Paused schedule — sending backfill signals\n")

        # Send 11 backfill signals in rapid succession.
        # The schedule is paused so the scheduler sits idle between signals;
        # all 11 arrive before a single decision task drains them.
        print(f"Sending {SIGNALS_TO_SEND} backfill signals...")
        now = datetime.now(timezone.utc)
        for i in range(SIGNALS_TO_SEND):
            window_end = now - timedelta(hours=i)
            window_start = window_end - timedelta(hours=1)
            await client.backfill_schedule(
                QUEUE_FULL_SCHEDULE_ID,
                start_time=window_start,
                end_time=window_end,
                overlap_policy=schedule_pb2.SCHEDULE_OVERLAP_POLICY_BUFFER,
            )
            print(f"  sent backfill {i + 1}/{SIGNALS_TO_SEND}")

        # Give the scheduler one decision task to process the signals.
        print("\nWaiting 5s for scheduler to process signals...")
        await asyncio.sleep(5)

        desc = await client.describe_schedule(QUEUE_FULL_SCHEDULE_ID)
        pending = desc.info.ongoing_backfills
        print(f"\nPending backfills after {SIGNALS_TO_SEND} signals: {len(pending)}")
        for bf in pending:
            print(f"  {bf.backfill_id or '(no id)'}  runs_completed={bf.runs_completed}/{bf.runs_total}")

        if len(pending) <= 10:
            print(
                f"\nQueue-full rejection confirmed: {SIGNALS_TO_SEND - len(pending)} signal(s) dropped."
                " Check scheduler_backfill_rejected_count_per_domain{reason=queue_full} metric."
            )
        else:
            print(f"\nUnexpected: {len(pending)} backfills queued (expected ≤10).")


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Test backfill queue-full rejection")
    p.add_argument("--target", default="localhost:7833", help="Cadence frontend host:port")
    p.add_argument("--domain", default="default", help="Cadence domain")
    p.add_argument("--cleanup", action="store_true", help="Delete the test schedule and exit")
    return p


if __name__ == "__main__":
    try:
        asyncio.run(main(build_parser().parse_args()))
    except KeyboardInterrupt:
        pass
