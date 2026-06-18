import asyncio
from datetime import timedelta

from cadence import Registry, workflow
from cadence.workflow import ActivityOptions, execute_activity

SCHEDULE_ID = "my-cadence-schedule"
BUFFER_SCHEDULE_ID = "my-buffer-schedule"
TASK_LIST = "schedule-sample-tl"
WORKFLOW_TYPE = "ScheduleSampleWorkflow"
SLOW_WORKFLOW_TYPE = "SlowScheduleSampleWorkflow"

registry = Registry()


@registry.workflow()
class ScheduleSampleWorkflow:
    @workflow.run
    async def run(self) -> str:
        return "ok"


@registry.activity(name="slow_sleep_activity")
async def slow_sleep_activity() -> str:
    await asyncio.sleep(25)
    return "done"


@registry.workflow(name=SLOW_WORKFLOW_TYPE)
class SlowScheduleSampleWorkflow:
    @workflow.run
    async def run(self) -> str:
        await execute_activity(
            slow_sleep_activity,
            options=ActivityOptions(
                schedule_to_close_timeout=timedelta(minutes=5),
                start_to_close_timeout=timedelta(minutes=5),
            ),
        )
        return "slow-ok"
