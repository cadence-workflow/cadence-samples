import asyncio
import logging
import signal

import cadence
from cadence.contrib.openai import PydanticDataConverter, cadence_registry

from research_tools import resolve_output_dir, tools_registry
from research_workflow import workflow_registry


TASK_LIST = "auto-research-task-list"

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)


async def main() -> None:
    resolve_output_dir()

    worker = cadence.worker.Worker(
        cadence.Client(
            domain="default",
            target="localhost:7833",
            data_converter=PydanticDataConverter(),
        ),
        TASK_LIST,
        cadence.Registry.of(
            cadence_registry.cadence_registry,
            tools_registry,
            workflow_registry,
        ),
    )

    async with worker:
        logger.info(
            "Auto Research worker started. Open "
            "http://localhost:8088/domains/default/cluster0/workflows"
        )
        logger.info("Workflow type: AutoResearchWorkflow")
        logger.info("Task list: %s", TASK_LIST)
        logger.info("Sample topic: The impact of AI on drug discovery")

        shutdown_event = asyncio.Event()
        loop = asyncio.get_running_loop()
        for sig in (signal.SIGTERM, signal.SIGINT):
            loop.add_signal_handler(sig, shutdown_event.set)

        logger.info("Press Ctrl+C to stop the worker.")
        await shutdown_event.wait()


if __name__ == "__main__":
    asyncio.run(main())
