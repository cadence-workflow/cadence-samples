import asyncio
import logging
import signal
import cadence
from cadence.contrib.openai import PydanticDataConverter, cadence_registry

from .tools import tools_registry
from .book_trip_agent import agent_registry

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)



async def main():
    # start Cadence worker
    worker = cadence.worker.Worker(
        cadence.Client(
            domain="default",
            target="localhost:7833",
            data_converter=PydanticDataConverter(),
        ),
        "agent-task-list",
        cadence.Registry.of(
            cadence_registry.cadence_registry,
            tools_registry,
            agent_registry),
    )

    # start BookFlightAgentWorkflow
    async with worker:
        logger.info("Worker started. Go to http://localhost:8088/domains/default/cluster0/workflows to start an agent run.")
        logger.info("Sample input: Book a trip for me from Uber Seattle Office to Uber San Francisco Office tomorrow at 10:00 AM")
        shutdown_event = asyncio.Event()
        loop = asyncio.get_running_loop()
        for sig in (signal.SIGTERM, signal.SIGINT):
            loop.add_signal_handler(sig, shutdown_event.set)
        logger.info("Press Ctrl+C to stop the worker.")

        await shutdown_event.wait()

if __name__ == "__main__":
    asyncio.run(main())
