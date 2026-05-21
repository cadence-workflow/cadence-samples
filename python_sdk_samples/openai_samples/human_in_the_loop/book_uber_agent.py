from dataclasses import dataclass
from typing import Any

from agents import Agent, RunConfig, Runner, ToolApprovalItem, function_tool
import cadence


agent_registry = cadence.Registry()

@dataclass
class UberTrip:
    from_address: str
    to_address: str
    passengers: int
    price: float
    driver_name: str
    driver_phone: str
    driver_car: str
    driver_car_plate: str
    driver_car_color: str

@agent_registry.activity(name="book_uber")
async def book_uber(from_address: str, to_address: str, passengers: int) -> UberTrip:
    """
    Book a Uber ride from start address to the destination address. default passengers is 1.
    """
    return UberTrip(from_address=from_address, to_address=to_address, passengers=passengers, price=100, driver_name="John Doe", driver_phone="1234567890", driver_car="Toyota", driver_car_plate="1234567890", driver_car_color="Red")

@agent_registry.workflow(name="BookUberAgentWorkflow")
class BookUberAgentWorkflow:
    def __init__(self) -> None:
        # Tool calls awaiting a decision, keyed by call_id.
        self._pending: dict[str, ToolApprovalItem] = {}
        # Decisions delivered via signals: call_id -> True (approve) / False (reject).
        self._decisions: dict[str, bool] = {}

    @cadence.workflow.query(name="get_interruptions")
    def get_interruptions(self) -> list[dict[str, Any]]:
        return [
            {
                "call_id": item.call_id,
                "tool_name": item.tool_name,
                "arguments": item.arguments,
            }
            for item in self._pending.values()
        ]

    @cadence.workflow.signal(name="approve_interruption")
    def approve_interruption(self, call_id: str, approved: bool = True) -> None:
        if call_id in self._pending:
            self._decisions[call_id] = approved

    @cadence.workflow.run
    async def run(self, input: str) -> str:
        agent = Agent(
            name="Book Uber Agent",
            instructions="You can book a uber ride from start address to destination address.",
            model="gpt-4o-mini",
            tools=[
                function_tool(book_uber, needs_approval=True),
            ],
        )

        run_config = RunConfig(tracing_disabled=True)
        run_input: Any = input

        while True:
            result = await Runner.run(agent, run_input, run_config=run_config)

            if not result.interruptions:
                return result.final_output

            self._pending = {
                item.call_id: item for item in result.interruptions if item.call_id
            }
            self._decisions = {}

            # Block until every pending tool call has an approve/reject signal.
            await cadence.workflow.wait_condition(
                lambda: all(call_id in self._decisions for call_id in self._pending)
            )

            # Resume from the existing run state with each decision applied.
            state = result.to_state()
            for call_id, item in self._pending.items():
                if self._decisions[call_id]:
                    state.approve(item)
                else:
                    state.reject(item, rejection_message="User rejected the tool call.")

            run_input = state
            self._pending = {}
            self._decisions = {}
