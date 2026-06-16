"""Create a Cadence schedule that fires ScheduleSampleWorkflow every minute.

Demonstrates:
  - ScheduleSpec: cron expression
  - ScheduleAction: StartWorkflowAction with workflow_id_prefix and memo
  - SchedulePolicies: overlap, catch-up, and pause_on_failure
  - Schedule-level memo (tagged metadata visible in describe/list)

Usage:
    uv run python -m schedule_samples.create_schedule
"""

import argparse
import asyncio
import json
from datetime import timedelta

from google.protobuf.duration import from_timedelta

from cadence.api.v1 import common_pb2, schedule_pb2, tasklist_pb2
from cadence.client import Client

from schedule_samples.workflow import SCHEDULE_ID, TASK_LIST, WORKFLOW_TYPE


def _payload(value: object) -> common_pb2.Payload:
    # Cadence encodes memo values as a JSON-wrapped single-element list,
    # matching the Go and Java SDK conventions.
    return common_pb2.Payload(data=json.dumps([value]).encode())


async def main(args: argparse.Namespace) -> None:
    async with Client(domain=args.domain, target=args.target) as client:
        await client.create_schedule(
            SCHEDULE_ID,
            spec=schedule_pb2.ScheduleSpec(cron_expression="* * * * *"),
            action=schedule_pb2.ScheduleAction(
                start_workflow=schedule_pb2.ScheduleAction.StartWorkflowAction(
                    workflow_type=common_pb2.WorkflowType(name=WORKFLOW_TYPE),
                    task_list=tasklist_pb2.TaskList(name=TASK_LIST),
                    # Prefix makes individual runs easy to find: <prefix><timestamp>
                    workflow_id_prefix=f"{SCHEDULE_ID}-",
                    execution_start_to_close_timeout=from_timedelta(timedelta(minutes=10)),
                    task_start_to_close_timeout=from_timedelta(timedelta(seconds=10)),
                )
            ),
            policies=schedule_pb2.SchedulePolicies(
                overlap_policy=schedule_pb2.SCHEDULE_OVERLAP_POLICY_SKIP_NEW,
                catch_up_policy=schedule_pb2.SCHEDULE_CATCH_UP_POLICY_SKIP,
                # Pause the schedule automatically if a started workflow fails,
                # preventing a cascade of failing runs until a human investigates.
                pause_on_failure=True,
            ),
            # Schedule-level memo: visible in DescribeSchedule and ListSchedules.
            # Values must be Payload (raw bytes); Cadence encodes them as JSON lists.
            memo=common_pb2.Memo(
                fields={
                    "owner": _payload("platform-team"),
                    "env": _payload("dev"),
                }
            ),
        )
        print(f"Created schedule {SCHEDULE_ID!r} (fires every minute)")


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Create a Cadence schedule")
    p.add_argument("--target", default="localhost:7833", help="Cadence frontend host:port")
    p.add_argument("--domain", default="default", help="Cadence domain")
    return p


if __name__ == "__main__":
    asyncio.run(main(build_parser().parse_args()))
