package main

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// ExpenseReport is the workflow input
type ExpenseReport struct {
	Employee   string        `json:"employee"`
	Department string        `json:"department"`
	Currency   string        `json:"currency"`
	Items      []ExpenseItem `json:"items"`
}

// ExpenseItem is one line of an expense report
type ExpenseItem struct {
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Amount      float64 `json:"amount"`
}

// ExpenseDecision is the workflow result
type ExpenseDecision struct {
	Status          string    `json:"status"`
	RequestedAmount float64   `json:"requestedAmount"`
	ApprovedAmount  float64   `json:"approvedAmount"`
	Reviewer        string    `json:"reviewer"`
	Comment         string    `json:"comment"`
	DecidedAt       time.Time `json:"decidedAt"`
}

// ExpenseReviewSignal is the payload for the approve, approve_partial and reject signals
type ExpenseReviewSignal struct {
	Reviewer string  `json:"reviewer"`
	Comment  string  `json:"comment"`
	Amount   float64 `json:"amount"`
}

const (
	ExpenseStatusPending  = "pending"
	ExpenseStatusApproved = "approved"
	ExpenseStatusPartial  = "partially_approved"
	ExpenseStatusRejected = "rejected"
	ExpenseStatusExpired  = "expired"
	ExpenseStatusCanceled = "canceled"

	// Cadence Web renders this query as the workflow's Home tab
	cadenceWebHomeQuery = "home"

	expenseReviewTimeout = time.Hour
)

// expenseHomeFormattedResponse is the JSON shape Cadence Web expects for home page query results.
type expenseHomeFormattedResponse struct {
	CadenceResponseType string `json:"cadenceResponseType"`
	Format              string `json:"format"`
	Data                string `json:"data"`
	Title               string `json:"title,omitempty"`
}

// ExpenseReportWorkflow waits for a reviewer to approve or reject an expense report.
// Its Home tab in Cadence Web shows the report, the review actions,
// the decision, and a list of other expense reports.
func ExpenseReportWorkflow(ctx workflow.Context, report ExpenseReport) (ExpenseDecision, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("ExpenseReportWorkflow started", zap.String("employee", report.Employee))

	requested := 0.0
	for _, item := range report.Items {
		if item.Amount < 0 {
			return ExpenseDecision{}, fmt.Errorf("item %q has a negative amount", item.Description)
		}
		requested += item.Amount
	}

	err := workflow.UpsertSearchAttributes(ctx, map[string]interface{}{
		"CustomKeywordField": ExpenseStatusPending,
		"CustomDoubleField":  requested,
	})
	if err != nil {
		return ExpenseDecision{}, err
	}

	decision := ExpenseDecision{
		Status:          ExpenseStatusPending,
		RequestedAmount: requested,
	}
	submittedAt := workflow.Now(ctx)
	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID

	err = workflow.SetQueryHandler(ctx, cadenceWebHomeQuery, func() (expenseHomeFormattedResponse, error) {
		return makeExpenseHome(report, decision, submittedAt, workflowID)
	})
	if err != nil {
		return decision, err
	}

	decide := func(status string, approved float64, signal ExpenseReviewSignal) {
		decision.Status = status
		decision.ApprovedAmount = approved
		decision.Reviewer = signal.Reviewer
		decision.Comment = signal.Comment
		decision.DecidedAt = workflow.Now(ctx)
	}

	selector := workflow.NewSelector(ctx)
	selector.AddReceive(workflow.GetSignalChannel(ctx, "approve"), func(ch workflow.Channel, _ bool) {
		var signal ExpenseReviewSignal
		ch.Receive(ctx, &signal)
		decide(ExpenseStatusApproved, requested, signal)
	})
	selector.AddReceive(workflow.GetSignalChannel(ctx, "approve_partial"), func(ch workflow.Channel, _ bool) {
		var signal ExpenseReviewSignal
		ch.Receive(ctx, &signal)
		// Clamp to [0, requested]. The full amount is an approval and zero is a rejection.
		amount := max(0, min(signal.Amount, requested))
		switch {
		case amount >= requested:
			decide(ExpenseStatusApproved, requested, signal)
		case amount == 0:
			decide(ExpenseStatusRejected, 0, signal)
		default:
			decide(ExpenseStatusPartial, amount, signal)
		}
	})
	selector.AddReceive(workflow.GetSignalChannel(ctx, "reject"), func(ch workflow.Channel, _ bool) {
		var signal ExpenseReviewSignal
		ch.Receive(ctx, &signal)
		decide(ExpenseStatusRejected, 0, signal)
	})
	var timerErr error
	selector.AddFuture(workflow.NewTimer(ctx, expenseReviewTimeout), func(f workflow.Future) {
		// The timer is also ready when the workflow is canceled.
		if timerErr = f.Get(ctx, nil); timerErr == nil {
			decide(ExpenseStatusExpired, 0, ExpenseReviewSignal{Reviewer: "System", Comment: "No review within 1 hour"})
		}
	})
	selector.Select(ctx)
	if timerErr != nil {
		decide(ExpenseStatusCanceled, 0, ExpenseReviewSignal{Reviewer: "System", Comment: "Workflow canceled"})
	}

	err = workflow.UpsertSearchAttributes(ctx, map[string]interface{}{
		"CustomKeywordField": decision.Status,
	})
	if err != nil {
		return decision, err
	}
	if timerErr != nil {
		return decision, timerErr
	}

	logger.Info("ExpenseReportWorkflow completed", zap.String("status", decision.Status))
	return decision, nil
}

// Domain, cluster, workflow ID and run ID are left out of the tags below:
// Cadence Web fills them in from the workflow page.
var expenseHomeTemplate = template.Must(template.New("expenseHome").Funcs(template.FuncMap{"escapeTags": escapeTags}).Parse(`**{{escapeTags .report.Employee}}** ({{escapeTags .report.Department}}) submitted this report at {{.submittedAt}}.

| Status | Requested | Approved |
|--------|-----------|----------|
| {{.statusBadge}} | {{escapeTags .requested}} | {{escapeTags .approved}} |

### Line items

| Description | Category | Amount |
|-------------|----------|--------|
{{range .report.Items}}| {{escapeTags .Description}} | {{escapeTags .Category}} | {{printf "%.2f" .Amount}} |
{{end}}
{{if .isPending}}
### Review

Approve or reject within 1 hour, or the report expires.

{% signal signalName="approve" label="✓ Approve {{.requested}}" input={"reviewer":"finance-ops","comment":"Looks good"} /%}
{% input-field name="approvedAmount" label="Approved amount" type="number" default={{.halfRaw}} required=true /%}
{% signal signalName="approve_partial" label="◐ Approve amount" input={"reviewer":"finance-ops","comment":"Partly in policy","amount":$approvedAmount} /%}
{% signal signalName="reject" label="✗ Reject" input={"reviewer":"finance-ops","comment":"Missing receipts"} /%}
{{else}}
### Decision

**{{escapeTags .decision.Reviewer}}** decided at {{.decidedAt}}: *{{escapeTags .decision.Comment}}*
{{end}}
---

### Other expense reports

{% workflows query="WorkflowType = 'cadence_samples.ExpenseReportWorkflow' AND WorkflowID != '{{.workflowID}}'" columns="WorkflowID,CloseStatus,CustomKeywordField,StartTime,CloseTime" /%}

---

### New report

{% input-field name="newAmount" label="Amount for the new receipt ({{.report.Currency}})" type="number" required=true /%}

{% start
	workflowType="cadence_samples.ExpenseReportWorkflow"
	label="Submit another report"
	taskList="cadence-samples-worker"
	timeoutSeconds=7200
	input={"employee":"{{.report.Employee}}","department":"{{.report.Department}}","currency":"{{.report.Currency}}","items":[{"description":"Team lunch","category":"Meals","amount":$newAmount}]}
/%}
`))

func makeExpenseHome(report ExpenseReport, decision ExpenseDecision, submittedAt time.Time, workflowID string) (expenseHomeFormattedResponse, error) {
	money := func(amount float64) string {
		return fmt.Sprintf("%.2f %s", amount, report.Currency)
	}
	half := decision.RequestedAmount / 2

	var markdown bytes.Buffer
	err := expenseHomeTemplate.Execute(&markdown, map[string]interface{}{
		"report":      report,
		"decision":    decision,
		"workflowID":  workflowID,
		"isPending":   decision.Status == ExpenseStatusPending,
		"statusBadge": getExpenseStatusBadge(decision.Status),
		"requested":   money(decision.RequestedAmount),
		"approved":    money(decision.ApprovedAmount),
		"halfRaw":     fmt.Sprintf("%.2f", half),
		"submittedAt": submittedAt.Format("2006-01-02 15:04:05"),
		"decidedAt":   decision.DecidedAt.Format("2006-01-02 15:04:05"),
	})
	if err != nil {
		return expenseHomeFormattedResponse{}, err
	}

	return expenseHomeFormattedResponse{
		CadenceResponseType: "formattedData",
		Format:              "text/markdown",
		Data:                markdown.String(),
		Title:               "Expense report from " + report.Employee,
	}, nil
}

// escapeTags stops user text from adding Markdoc tags. Markdown shows &#123; as "{".
func escapeTags(s string) string {
	return strings.ReplaceAll(s, "{%", "&#123;%")
}

func getExpenseStatusBadge(status string) string {
	switch status {
	case ExpenseStatusPending:
		return "🟡 **Pending review**"
	case ExpenseStatusApproved:
		return "🟢 **Approved**"
	case ExpenseStatusPartial:
		return "🟠 **Partially approved**"
	case ExpenseStatusRejected:
		return "🔴 **Rejected**"
	case ExpenseStatusExpired:
		return "⚪ **Expired**"
	case ExpenseStatusCanceled:
		return "⚫ **Canceled**"
	default:
		return status
	}
}
