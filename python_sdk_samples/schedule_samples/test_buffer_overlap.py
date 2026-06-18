"""Test BUFFER overlap policy: fires queue up and drain one at a time.

Creates a schedule backed by SlowScheduleSampleWorkflow (25s sleep) with
BUFFER overlap. Pauses it before any cron fires, backfills 3 minutes (3
fires), then unpauses. Polls describe every 5s and prints a timestamp each
time total_runs increments. Sequential increments spaced ~25s apart confirm
that buffered fires wait for the previous run to finish before starting.

Usage:
    uv run python -m schedule_samples.test_buffer_overlap [--cleanup]
"""

import argparse
import asyncio
import time
from datetime import datetime, timedelta, timezone

from google.protobuf.duration import from_timedelta

from cadence.api.v1 import common_pb2, schedule_pb2, tasklist_pb2
from cadence.client import Client

from schedule_samples.workflow import BUFFER_SCHEDULE_ID, SLOW_WORKFLOW_TYPE, TASK_LIST

BACKFILL_FIRES = 3  # number of fires to backfill


async def _delete_if_exists(client: Client) -> None:
    try:
        await client.delete_schedule(BUFFER_SCHEDULE_ID)
        print(f"  deleted existing schedule {BUFFER_SCHEDULE_ID!r}")
    except Exception:
        pass


async def main(args: argparse.Namespace) -> None:
    async with Client(domain=args.domain, target=args.target) as client:
        if args.cleanup:
            await _delete_if_exists(client)
            return

        print("=== BUFFER overlap policy test ===\n")
        await _delete_if_exists(client)

        await client.create_schedule(
            BUFFER_SCHEDULE_ID,
            spec=schedule_pb2.ScheduleSpec(cron_expression="* * * * *"),
            action=schedule_pb2.ScheduleAction(
                start_workflow=schedule_pb2.ScheduleAction.StartWorkflowAction(
                    workflow_type=common_pb2.WorkflowType(name=SLOW_WORKFLOW_TYPE),
                    task_list=tasklist_pb2.TaskList(name=TASK_LIST),
                    workflow_id_prefix=f"{BUFFER_SCHEDULE_ID}-",
                    execution_start_to_close_timeout=from_timedelta(timedelta(minutes=10)),
                    task_start_to_close_timeout=from_timedelta(timedelta(seconds=10)),
                )
            ),
            policies=schedule_pb2.SchedulePolicies(
                overlap_policy=schedule_pb2.SCHEDULE_OVERLAP_POLICY_BUFFER,
                catch_up_policy=schedule_pb2.SCHEDULE_CATCH_UP_POLICY_SKIP,
            ),
        )

        # Pause immediately so no regular cron fires contaminate the test.
        await client.pause_schedule(BUFFER_SCHEDULE_ID, reason="buffer test setup")
        print(f"Created and paused schedule {BUFFER_SCHEDULE_ID!r} (BUFFER overlap, slow workflow)")

        # Backfill N fires so they queue up in the buffer.
        end = datetime.now(timezone.utc)
        start = end - timedelta(minutes=BACKFILL_FIRES)
        await client.backfill_schedule(
            BUFFER_SCHEDULE_ID,
            start_time=start,
            end_time=end,
            overlap_policy=schedule_pb2.SCHEDULE_OVERLAP_POLICY_BUFFER,
        )
        print(f"Backfilled {BACKFILL_FIRES} minutes ({start:%H:%M} → {end:%H:%M} UTC)")
        print(f"Expected: {BACKFILL_FIRES} fires run sequentially, each ~25s apart\n")

        # Small delay to let the scheduler receive and queue the backfill signal.
        await asyncio.sleep(3)

        # Unpause so the scheduler starts draining the buffer.
        await client.unpause_schedule(BUFFER_SCHEDULE_ID, reason="buffer test start")
        print("Unpaused — watching total_runs (Ctrl-C to stop)\n")

        prev_runs = -1
        fire_times: list[float] = []
        start_wall = time.monotonic()

        while True:
            await asyncio.sleep(5)
            desc = await client.describe_schedule(BUFFER_SCHEDULE_ID)
            total = desc.info.total_runs

            if total != prev_runs:
                elapsed = time.monotonic() - start_wall
                gap = elapsed - fire_times[-1] if fire_times else 0.0
                fire_times.append(elapsed)
                gap_str = f"  (+{gap:.0f}s since last)" if len(fire_times) > 1 else "  (first fire)"
                print(f"  total_runs={total}  elapsed={elapsed:.0f}s{gap_str}")
                prev_runs = total

            if total >= BACKFILL_FIRES:
                print(f"\nAll {BACKFILL_FIRES} fires completed.")
                if len(fire_times) >= 2:
                    gaps = [fire_times[i] - fire_times[i - 1] for i in range(1, len(fire_times))]
                    avg_gap = sum(gaps) / len(gaps)
                    print(f"Average gap between fires: {avg_gap:.0f}s (BUFFER drains at each cron tick after previous run finishes)")
                break

        # Pause again to prevent ongoing cron fires after the test.
        await client.pause_schedule(BUFFER_SCHEDULE_ID, reason="buffer test complete")
        print("Schedule paused. Run with --cleanup to delete it.")


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Test BUFFER overlap policy")
    p.add_argument("--target", default="localhost:7833", help="Cadence frontend host:port")
    p.add_argument("--domain", default="default", help="Cadence domain")
    p.add_argument("--cleanup", action="store_true", help="Delete the test schedule and exit")
    return p


if __name__ == "__main__":
    try:
        asyncio.run(main(build_parser().parse_args()))
    except KeyboardInterrupt:
        pass
