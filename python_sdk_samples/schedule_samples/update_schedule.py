"""Update a Cadence schedule using a read-modify-write callback.

Usage:
    uv run python -m schedule_samples.update_schedule
"""

import argparse
import asyncio

from cadence.client import Client

from schedule_samples.workflow import SCHEDULE_ID


async def main(args: argparse.Namespace) -> None:
    def set_cron(schedule: object) -> None:
        schedule.spec.cron_expression = args.cron

    async with Client(domain=args.domain, target=args.target) as client:
        await client.update_schedule(SCHEDULE_ID, set_cron)
        print(f"Updated schedule {SCHEDULE_ID!r} cron={args.cron!r}")


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Update a Cadence schedule")
    p.add_argument("--target", default="localhost:7833", help="Cadence frontend host:port")
    p.add_argument("--domain", default="default", help="Cadence domain")
    p.add_argument("--cron", default="*/2 * * * *", help="New cron expression (default: every 2 minutes)")
    return p


if __name__ == "__main__":
    asyncio.run(main(build_parser().parse_args()))
