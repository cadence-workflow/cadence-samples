# What

This demo shows how to run OpenAI agents in a durable way (retries, reset to a check point). Specifically, we shows how to deal with human-in-the-loop.

# Setup OpenAI API keys

Make sure the OPENAI_API_KEY environment variable is set. See details for best practices.
https://help.openai.com/en/articles/5112595-best-practices-for-api-key-safety

# Setup Cadence Server

Refer to step 3 of the [Quick Start](../../README.md) in `python_sdk_samples/README.md` for instructions on starting the Cadence Server:

```bash
curl -LO https://raw.githubusercontent.com/cadence-workflow/cadence/refs/heads/master/docker/docker-compose.yml && docker-compose up --wait
```

# Start Agent workers

```
cd python_sdk_samples
uv sync
uv run python -m openai_samples.human_in_the_loop.main
```

# Trigger Agent Run

Run Cadence CLI command

```
cadence --domain default workflow start \
--workflow_type BookUberAgentWorkflow \
--tasklist agent-task-list \
--execution_timeout 30 \
--input '"Book a trip for me (1 passenger) from Uber Seattle Office to Seattle Airport at 10:00 AM"'
```

# Agent Run Stopped

Because we require approval on the tool call, the workflow run stopped.
![Agent Waiting](images/agent_waiting.png)

# View Interruptions

Click `get_interruptions` button in `Queries` tab after selecting the workflow run in [cadence-web](http://localhost:8088/domains/default/cluster0/workflows)
![Workflow Quries Screenshot](images/query_get_interruptions.png)

Make decisions on whether to approve or reject the tool calls.

# View Agent Run Result

After clicking on the button, workflow received a signal and decides the next step. In this example, we see completed with approval.

![Agent Complete](images/agent_complete.png)
