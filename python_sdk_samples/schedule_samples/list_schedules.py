"""List all Cadence schedules in a domain.

Usage:
    uv run python -m schedule_samples.list_schedules
"""

import argparse
import asyncio

from cadence.client import Client


async def main(args: argparse.Namespace) -> None:
    async with Client(domain=args.domain, target=args.target) as client:
        count = 0
        async for entry in client.list_schedules():
            print(entry.schedule_id)
            count += 1
        if count == 0:
            print("(no schedules found)")
        else:
            print(f"\n{count} schedule(s) in domain {args.domain!r}")


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="List Cadence schedules")
    p.add_argument("--target", default="localhost:7833", help="Cadence frontend host:port")
    p.add_argument("--domain", default="default", help="Cadence domain")
    return p


if __name__ == "__main__":
    asyncio.run(main(build_parser().parse_args()))
