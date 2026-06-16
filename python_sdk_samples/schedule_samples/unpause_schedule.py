"""Unpause a Cadence schedule — resumes firing after a pause.

Uses SKIP catch-up policy so fires missed during the pause are discarded.

Usage:
    uv run python -m schedule_samples.unpause_schedule
"""

import argparse
import asyncio

from cadence.api.v1 import schedule_pb2
from cadence.client import Client

from schedule_samples.workflow import SCHEDULE_ID


async def main(args: argparse.Namespace) -> None:
    async with Client(domain=args.domain, target=args.target) as client:
        await client.unpause_schedule(
            SCHEDULE_ID,
            reason=args.reason,
            catch_up_policy=schedule_pb2.SCHEDULE_CATCH_UP_POLICY_SKIP,
        )
        print(f"Unpaused schedule {SCHEDULE_ID!r} (reason: {args.reason!r})")


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Unpause a Cadence schedule")
    p.add_argument("--target", default="localhost:7833", help="Cadence frontend host:port")
    p.add_argument("--domain", default="default", help="Cadence domain")
    p.add_argument("--reason", default="resuming via sample", help="Unpause reason recorded on the schedule")
    return p


if __name__ == "__main__":
    asyncio.run(main(build_parser().parse_args()))
