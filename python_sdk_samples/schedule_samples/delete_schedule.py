"""Delete a Cadence schedule.

Already-started workflow runs continue to completion; only future fires
are cancelled.

Usage:
    uv run python -m schedule_samples.delete_schedule
"""

import argparse
import asyncio

from cadence.client import Client

from schedule_samples.workflow import SCHEDULE_ID


async def main(args: argparse.Namespace) -> None:
    async with Client(domain=args.domain, target=args.target) as client:
        await client.delete_schedule(SCHEDULE_ID)
        print(f"Deleted schedule {SCHEDULE_ID!r}")


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Delete a Cadence schedule")
    p.add_argument("--target", default="localhost:7833", help="Cadence frontend host:port")
    p.add_argument("--domain", default="default", help="Cadence domain")
    return p


if __name__ == "__main__":
    asyncio.run(main(build_parser().parse_args()))
