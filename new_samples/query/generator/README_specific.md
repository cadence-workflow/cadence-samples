## MDX Query Workflow

This workflow is very similar to the signal sample, except shows how to use queries in MDX format. Try the following CLI

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 1000 \
  --workflow_type cadence_samples.MDXQueryWorkflow
```

Verify that your workflow started. Your can find your worklow by looking at the "Workflow type" column.

If this is your first sample, please refer to [HelloWorkflow sample](https://github.com/cadence-workflow/cadence-samples/tree/master/new_samples/hello_world) about how to view your workflows.


### Signal your workflow using the MDX query

This workflow will need a signal to complete successfully. In this sample, instead of using CLI, API or Web, we will use an MDX query which has signal buttons. 

* Go to the `cadence-samples` domain in cadence-web and click on this workflow. 
* Click on the "Query" tab.
