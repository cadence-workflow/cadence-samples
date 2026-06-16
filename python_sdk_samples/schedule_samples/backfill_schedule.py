"""Backfill a Cadence schedule — replay missed fires over a past time window.

Submits all fires that would have occurred in the last 2 hours using BUFFER
overlap so they queue rather than skip each other.

Usage:
    uv run python -m schedule_samples.backfill_schedule
"""

import argparse
import asyncio
from datetime import datetime, timedelta, timezone

from cadence.api.v1 import schedule_pb2
from cadence.client import Client

from schedule_samples.workflow import SCHEDULE_ID


async def main(args: argparse.Namespace) -> None:
    end = datetime.now(timezone.utc)
    start = end - timedelta(hours=args.hours)

    async with Client(domain=args.domain, target=args.target) as client:
        await client.backfill_schedule(
            SCHEDULE_ID,
            start_time=start,
            end_time=end,
            overlap_policy=schedule_pb2.SCHEDULE_OVERLAP_POLICY_BUFFER,
        )
        print(
            f"Backfilled schedule {SCHEDULE_ID!r} "
            f"({start:%H:%M} → {end:%H:%M} UTC, {args.hours}h window)"
        )


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Backfill a Cadence schedule")
    p.add_argument("--target", default="localhost:7833", help="Cadence frontend host:port")
    p.add_argument("--domain", default="default", help="Cadence domain")
    p.add_argument("--hours", type=int, default=2, help="How many hours back to backfill")
    return p


if __name__ == "__main__":
    asyncio.run(main(build_parser().parse_args()))
