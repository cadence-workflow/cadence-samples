from cadence import Registry, workflow

SCHEDULE_ID = "my-cadence-schedule"
TASK_LIST = "schedule-sample-tl"
WORKFLOW_TYPE = "ScheduleSampleWorkflow"

registry = Registry()


@registry.workflow()
class ScheduleSampleWorkflow:
    @workflow.run
    async def run(self) -> str:
        return "ok"
