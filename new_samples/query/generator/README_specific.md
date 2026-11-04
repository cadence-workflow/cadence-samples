## Query Workflow Sample

This sample demonstrates **workflow queries** - inspecting workflow state without affecting execution.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 180 \
  --workflow_type cadence_samples.QueryWorkflow
```

### Query the Workflow

While the workflow is running, query its state:

```bash
cadence --env development \
  --domain cadence-samples \
  workflow query \
  --wid <workflow_id> \
  --qt state
```

### What Happens

The workflow goes through states that you can query:

```
Time 0:   state = "started"
Time 1s:  state = "waiting on timer"
Time 2m:  state = "done" (workflow completes)
```

### Key Concept: Query Handler

```go
func QueryWorkflow(ctx workflow.Context) error {
    currentState := "started"
    
    // Register query handler for "state" query type
    workflow.SetQueryHandler(ctx, "state", func() (string, error) {
        return currentState, nil
    })
    
    currentState = "waiting on timer"
    workflow.NewTimer(ctx, 2*time.Minute).Get(ctx, nil)
    
    currentState = "done"
    return nil
}
```

### Use Cases

- Progress monitoring dashboards
- Debugging running workflows
- Health checks without affecting execution


## Expense Report Sample (workflow Home page)

This sample shows a custom **Home** page in Cadence Web. The workflow takes an expense report, waits for a reviewer, and returns the decision. Its `home` query returns Markdoc that renders the report, review buttons, and a list of other expense reports.

Start Cadence Web with `CADENCE_WORKFLOW_HOME_ENABLED=true`.

### Start the Workflow

```bash
cadence --env development \
  --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 7200 \
  --workflow_type cadence_samples.ExpenseReportWorkflow \
  --input '{"employee":"Sam Lee","department":"Engineering","currency":"USD","items":[{"description":"Hotel, 2 nights","category":"Travel","amount":412.8},{"description":"Client dinner","category":"Meals","amount":96.25},{"description":"Conference pass","category":"Training","amount":350}]}'
```

### What Happens

Open the workflow in Cadence Web. The Home tab shows the report and these review actions:

- **Approve**: approves the full amount.
- **Approve amount**: approves the amount in the **Approved amount** field. It starts at half the requested amount.
- **Reject**: rejects the report.

With no review within 1 hour, the report expires. After a decision, the workflow completes, and the Home page shows the decision.

Below the review actions, the page shows a list of other expense reports and a **Submit another report** button that starts a new workflow with the amount you type.

### Key Concept: Home Page Tags

```go
workflow.SetQueryHandler(ctx, "home", func() (expenseHomeFormattedResponse, error) {
    return makeExpenseHome(report, decision, submittedAt, workflowID)
})
```

The returned Markdoc uses these tags. Each one gets domain, cluster, workflow ID and run ID from the page:

| Tag | Shows |
|-----|-------|
| `{% workflows query="..." /%}` | A list of workflows that match a visibility query. Needs advanced visibility. |
| `{% signal /%}` and `{% start /%}` | Buttons that signal this workflow or start a new one. |
| `{% input-field /%}` | A text or number input. A button sends its value as `$name`, for example `input={"amount":$approvedAmount}`. |
