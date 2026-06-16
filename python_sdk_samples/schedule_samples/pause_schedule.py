"""Pause a Cadence schedule. No new workflow runs are started until unpaused.

Usage:
    uv run python -m schedule_samples.pause_schedule
"""

import argparse
import asyncio

from cadence.client import Client

from schedule_samples.workflow import SCHEDULE_ID


async def main(args: argparse.Namespace) -> None:
    async with Client(domain=args.domain, target=args.target) as client:
        await client.pause_schedule(SCHEDULE_ID, reason=args.reason)
        print(f"Paused schedule {SCHEDULE_ID!r} (reason: {args.reason!r})")


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Pause a Cadence schedule")
    p.add_argument("--target", default="localhost:7833", help="Cadence frontend host:port")
    p.add_argument("--domain", default="default", help="Cadence domain")
    p.add_argument("--reason", default="paused via sample", help="Pause reason recorded on the schedule")
    return p


if __name__ == "__main__":
    asyncio.run(main(build_parser().parse_args()))
