from dataclasses import dataclass
from textwrap import dedent
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


# Shape that Cadence Web expects to render a query result as markdown.
@dataclass
class MarkdownQueryResponse:
    cadenceResponseType: str
    format: str
    data: str


@agent_registry.workflow(name="BookUberAgentWorkflow")
class BookUberAgentWorkflow:
    def __init__(self) -> None:
        # Tool calls awaiting a decision, keyed by call_id.
        self._pending: dict[str, ToolApprovalItem] = {}
        # Decisions delivered via signals: call_id -> True (approve) / False (reject).
        self._decisions: dict[str, bool] = {}

    @cadence.workflow.query(name="get_interruptions")
    def get_interruptions(self) -> MarkdownQueryResponse:
        info = cadence.workflow.WorkflowContext.get().info()
        return MarkdownQueryResponse(
            cadenceResponseType="formattedData",
            format="text/markdown",
            data=_render_interruptions_markdown(
                domain=info.workflow_domain,
                workflow_id=info.workflow_id,
                run_id=info.workflow_run_id,
                pending=self._pending,
            ),
        )

    @cadence.workflow.signal(name="approve_tool_call")
    def approve_tool_call(self, call_id: str) -> None:
        if call_id in self._pending:
            self._decisions[call_id] = True

    @cadence.workflow.signal(name="reject_tool_call")
    def reject_tool_call(self, call_id: str) -> None:
        if call_id in self._pending:
            self._decisions[call_id] = False

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

            for item in result.interruptions:
                if not item.call_id:
                    raise RuntimeError("Tool call ID is required for interruption %s", item.qualified_name)
                self._pending[item.call_id] = item

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


def _render_interruptions_markdown(
    *,
    domain: str,
    workflow_id: str,
    run_id: str,
    pending: dict[str, ToolApprovalItem],
) -> str:
    if not pending:
        return dedent(
            """\
            ## Tool Approvals

            _No tool calls are awaiting approval._
            """
        )

    sections: list[str] = [
        "## Tool Approvals",
        "",
        "The agent is paused. Approve or reject each pending tool call below.",
        "",
        "---",
        "",
    ]
    for item in pending.values():
        sections.append(f"### `{item.tool_name}`")
        sections.append("")
        sections.append(f"- **Call ID:** `{item.call_id}`")
        if item.arguments:
            sections.append("- **Arguments:**")
            sections.append("")
            sections.append("```json")
            sections.append(item.arguments)
            sections.append("```")
        sections.append("")
        sections.append(_render_decision_buttons(domain, workflow_id, run_id, item.call_id))
        sections.append("")
        sections.append("---")
        sections.append("")

    return "\n".join(sections)


def _render_decision_buttons(
    domain: str, workflow_id: str, run_id: str, call_id: str
) -> str:
    return dedent(
        f"""\
        {{% signal
            signalName="approve_tool_call"
            label="✓ Approve"
            domain="{domain}"
            cluster="cluster0"
            workflowId="{workflow_id}"
            runId="{run_id}"
            input="{call_id}"
        /%}}
        {{% signal
            signalName="reject_tool_call"
            label="✗ Reject"
            domain="{domain}"
            cluster="cluster0"
            workflowId="{workflow_id}"
            runId="{run_id}"
            input="{call_id}"
        /%}}"""
    )
