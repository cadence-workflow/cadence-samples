import cadence
from agents import Agent, function_tool, Runner, RunConfig

from .tools import book_flight, book_uber

agent_registry = cadence.Registry()

@agent_registry.workflow(name="BookTripAgentWorkflow")
class BookTripAgentWorkflow:

    @cadence.workflow.run
    async def run(self, input: str) -> str:

        long_trip_agent =Agent(
            name = "Plan Long Trip Agent",
            model = "gpt-4o-mini",
            instructions= """
            Book a flight from start address to destination address.
            Use Uber to connect local address to airport or vice versa.
            """,
            tools = [
                function_tool(book_flight),
                function_tool(book_uber),
            ],
        )

        short_trip_agent =Agent(
            name = "Plan Short Trip Agent",
            instructions= """
            Book a Uber ride from start address to destination address.
            """,
            model = "gpt-4o-mini",
            tools = [
                function_tool(book_uber),
            ],
        )

        # define agent using OpenAI SDK as usual
        agent =Agent(
            name = "Book Trip Agent",
            instructions = """
            You are a trip planner. You can plan short or long trips.
            """,
            model = "gpt-4o-mini",
            handoffs = [
                short_trip_agent,
                long_trip_agent,
            ],
        )
        result = await Runner.run(agent, input, run_config=RunConfig(
                tracing_disabled=True,
            ))
        return result.final_output
