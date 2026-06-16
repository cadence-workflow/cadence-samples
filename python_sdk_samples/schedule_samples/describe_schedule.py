"""Describe a Cadence schedule — prints spec, state, policies, runtime info, and memo.

Usage:
    uv run python -m schedule_samples.describe_schedule
"""

import argparse
import asyncio
import json

from cadence.api.v1 import schedule_pb2
from cadence.client import Client

from schedule_samples.workflow import SCHEDULE_ID

_OVERLAP_NAMES = {
    schedule_pb2.SCHEDULE_OVERLAP_POLICY_INVALID: "INVALID",
    schedule_pb2.SCHEDULE_OVERLAP_POLICY_SKIP_NEW: "SKIP_NEW",
    schedule_pb2.SCHEDULE_OVERLAP_POLICY_BUFFER: "BUFFER",
    schedule_pb2.SCHEDULE_OVERLAP_POLICY_CONCURRENT: "CONCURRENT",
    schedule_pb2.SCHEDULE_OVERLAP_POLICY_CANCEL_PREVIOUS: "CANCEL_PREVIOUS",
    schedule_pb2.SCHEDULE_OVERLAP_POLICY_TERMINATE_PREVIOUS: "TERMINATE_PREVIOUS",
}


def _decode_payload(data: bytes) -> object:
    try:
        return json.loads(data)[0]
    except Exception:
        return data.decode(errors="replace")


async def main(args: argparse.Namespace) -> None:
    async with Client(domain=args.domain, target=args.target) as client:
        desc = await client.describe_schedule(SCHEDULE_ID)

        print(f"Schedule : {SCHEDULE_ID!r}")
        print(f"  cron           : {desc.spec.cron_expression!r}")
        print(f"  paused         : {desc.state.paused}")
        if desc.state.paused:
            print(f"  pause reason   : {desc.state.pause_info.reason!r}")

        print(f"  overlap policy : {_OVERLAP_NAMES.get(desc.policies.overlap_policy, str(desc.policies.overlap_policy))}")
        print(f"  pause_on_fail  : {desc.policies.pause_on_failure}")

        info = desc.info
        if info.HasField("next_run_time"):
            print(f"  next run       : {info.next_run_time.ToDatetime()} UTC")
        if info.HasField("last_run_time"):
            print(f"  last run       : {info.last_run_time.ToDatetime()} UTC")
        print(f"  total runs     : {info.total_runs}")

        if desc.memo.fields:
            decoded = {k: _decode_payload(v.data) for k, v in desc.memo.fields.items()}
            print(f"  memo           : {decoded}")


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Describe a Cadence schedule")
    p.add_argument("--target", default="localhost:7833", help="Cadence frontend host:port")
    p.add_argument("--domain", default="default", help="Cadence domain")
    return p


if __name__ == "__main__":
    asyncio.run(main(build_parser().parse_args()))
