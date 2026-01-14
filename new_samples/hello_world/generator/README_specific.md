### Start your workflow

This workflow takes an input message and greet you as response. Try the following CLI

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --workflow_type cadence_samples.HelloWorldWorkflow \
  --tl cadence-samples-worker \
  --et 60 \
  --input '{"message":"Cadence"}'
```

You should see output like this:

![Trigger command output](images/02-trigger-command-started-workflow.png)

And the worker will log the completed workflow:

![Worker output showing workflow completed](images/01-worker-output-workflow-completed.png)

Here are the details to this command:

* `--domain` option describes under which domain to run this workflow
* `--env development` calls the "local" cadence server
* `--workflow_type` option describes which workflow to execute
* `-tl` (or `--tasklist`) tells cadence-server which tasklist to schedule tasks with. This is the same tasklist the worker polls tasks from. See worker.go
* `--et` (or `--execution_timeout`) tells cadence server how long to wait until timing out the workflow
* `--input` is the input to your workflow

To see more options run `cadence --help`

### View your workflow

#### Cadence UI (cadence-web)

Click on `cadence-samples` domain in cadence-web to view your workflow.

![Workflow list showing completed workflow](images/03-web-ui-workflow-list-completed.png)

Click on the workflow to see details:

* In Summary tab, you will see the input and output to your workflow

![Summary tab](images/04-web-ui-summary-tab.png)

* Click on History tab to see individual steps. Expand an activity to see its result:

![History tab with activity result](images/05-web-ui-history-activity-result.png)

* In Summary tab, you will see the input and output to your workflow
* Click on History tab to see individual steps.

#### CLI

List workflows using the following command:

```bash
cadence --env development --domain cadence-samples workflow list
```

You can view an individual workflow by using the following command:

```bash
cadence --env development \
  --domain cadence-samples \
  workflow describe \
  --wid <workflow_id>
```

* `workflow` is the noun to run commands within workflow scope
* `describe` is the verb to return the summary of the workflow
* `--wid` (or `--workflow_id`) is the option to pass the workflow id. If there are multiple "run"s, it will return the latest one.
* (optional) `--rid` (or `--run_id`) is the option to pass the run id to describe a specific run, instead of the latest.

To view the entire history of the workflow, use the following command:

```bash
cadence --env development \
  --domain cadence-samples \
  workflow show \
  --wid <workflow_id>
```

## Troubleshooting

If you see port conflicts when starting Docker, use `lsof` to find what's using the port:

![Docker port conflict troubleshooting](images/06-docker-port-conflict-troubleshooting.png)

See the main [README](../../README.md#docker-troubleshooting) for detailed Docker troubleshooting steps.