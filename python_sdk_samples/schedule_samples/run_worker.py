"""Start a worker that executes ScheduleSampleWorkflow.

Keep this running in a terminal while you run the other schedule samples.

Usage:
    uv run python -m schedule_samples.run_worker
"""

import argparse
import asyncio

from cadence.client import Client
from cadence.worker import Worker

from schedule_samples.workflow import TASK_LIST, registry


async def main(args: argparse.Namespace) -> None:
    async with Client(domain=args.domain, target=args.target) as client:
        print(f"Worker running on task list {TASK_LIST!r} — Ctrl-C to stop")
        async with Worker(client, TASK_LIST, registry):
            await asyncio.Event().wait()


def build_parser() -> argparse.ArgumentParser:
    p = argparse.ArgumentParser(description="Run the schedule sample worker")
    p.add_argument("--target", default="localhost:7833", help="Cadence frontend host:port")
    p.add_argument("--domain", default="default", help="Cadence domain")
    return p


if __name__ == "__main__":
    args = build_parser().parse_args()
    try:
        asyncio.run(main(args))
    except KeyboardInterrupt:
        pass
