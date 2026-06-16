"""Describe a Cadence schedule — prints its spec and current state.

Usage:
    uv run python -m schedule_samples.describe_schedule
"""

import argparse
import asyncio

from cadence.client import Client

from schedule_samples.workflow import SCHEDULE_ID


async def main(args: argparse.Namespace) -> None:
    async with Client(domain=args.domain, target=args.target) as client:
        desc = await client.describe_schedule(SCHEDULE_ID)
        print(f"Schedule : {SCHEDULE_ID!r}")
        print(f"  cron   : {desc.spec.cron_expression!r}")
        print(f"  paused : {desc.state.paused}")
        if desc.state.paused:
            print(f"  reason : {desc.state.pause_info.reason!r}")


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Describe a Cadence schedule")
    p.add_argument("--target", default="localhost:7833", help="Cadence frontend host:port")
    p.add_argument("--domain", default="default", help="Cadence domain")
    return p


if __name__ == "__main__":
    asyncio.run(main(build_parser().parse_args()))
