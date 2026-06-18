"""Create a Cadence schedule that fires ScheduleSampleWorkflow every minute.

Usage:
    uv run python -m schedule_samples.create_schedule
    uv run python -m schedule_samples.create_schedule --schedule-id my-other-schedule
"""

import argparse
import asyncio
import json
from datetime import timedelta

from google.protobuf.duration import from_timedelta

from cadence.api.v1 import common_pb2, schedule_pb2, tasklist_pb2
from cadence.client import Client
from cadence.error import CadenceRpcError

from schedule_samples.workflow import SCHEDULE_ID, TASK_LIST, WORKFLOW_TYPE


def _payload(value: object) -> common_pb2.Payload:
    # Cadence memo values are JSON-encoded as a single-element list, matching Go/Java SDK behavior.
    return common_pb2.Payload(data=json.dumps([value]).encode())


async def main(args: argparse.Namespace) -> None:
    schedule_id = args.schedule_id
    async with Client(domain=args.domain, target=args.target) as client:
        try:
            await client.create_schedule(
                schedule_id,
                spec=schedule_pb2.ScheduleSpec(cron_expression="* * * * *"),
                action=schedule_pb2.ScheduleAction(
                    start_workflow=schedule_pb2.ScheduleAction.StartWorkflowAction(
                        workflow_type=common_pb2.WorkflowType(name=WORKFLOW_TYPE),
                        task_list=tasklist_pb2.TaskList(name=TASK_LIST),
                        workflow_id_prefix=f"{schedule_id}-",
                        execution_start_to_close_timeout=from_timedelta(timedelta(minutes=10)),
                        task_start_to_close_timeout=from_timedelta(timedelta(seconds=10)),
                    )
                ),
                policies=schedule_pb2.SchedulePolicies(
                    overlap_policy=schedule_pb2.SCHEDULE_OVERLAP_POLICY_SKIP_NEW,
                    catch_up_policy=schedule_pb2.SCHEDULE_CATCH_UP_POLICY_SKIP,
                    pause_on_failure=True,
                ),
                memo=common_pb2.Memo(
                    fields={
                        "owner": _payload("platform-team"),
                        "env": _payload("dev"),
                    }
                ),
            )
            print(f"Created schedule {schedule_id!r} (fires every minute)")
        except CadenceRpcError as e:
            if "already exists" in str(e):
                print(f"Schedule {schedule_id!r} already exists. Run delete_schedule first or use --schedule-id to pick a different name.")
            else:
                raise


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Create a Cadence schedule")
    p.add_argument("--target", default="localhost:7833", help="Cadence frontend host:port")
    p.add_argument("--domain", default="default", help="Cadence domain")
    p.add_argument("--schedule-id", default=SCHEDULE_ID, help="Schedule ID to create (default: %(default)s)")
    return p


if __name__ == "__main__":
    asyncio.run(main(build_parser().parse_args()))
