package main

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// orderDashboardFormattedResponse is the JSON shape Cadence Web expects for markdown query results (formattedData, text/markdown, data).
type orderDashboardFormattedResponse struct {
	CadenceResponseType string `json:"cadenceResponseType"`
	Format              string `json:"format"`
	Data                string `json:"data"`
}

// Order represents an e-commerce order being fulfilled
type Order struct {
	OrderID       string
	CustomerName  string
	CustomerEmail string
	Items         []OrderItem
	TotalAmount   float64
	Status        string
	TrackingNum   string
	Carrier       string
	RefundAmount  float64
	RefundReason  string
	CreatedAt     time.Time
}

// OrderItem represents a line item in an order
type OrderItem struct {
	Name     string
	Quantity int
	Price    float64
}

// ActionLogEntry represents an ops action taken on the order
type ActionLogEntry struct {
	Timestamp time.Time
	Action    string
	Operator  string
	Details   string
}

// Order status constants
const (
	StatusPendingPayment  = "pending_payment"
	StatusPaymentApproved = "payment_approved"
	StatusReadyToShip     = "ready_to_ship"
	StatusShipped         = "shipped"
	StatusDelivered       = "delivered"
	StatusCancelled       = "cancelled"
	StatusRefunded        = "refunded"
)

// Signal payloads
type RejectPaymentSignal struct {
	Reason   string `json:"reason"`
	Operator string `json:"operator"`
}

type ApprovePaymentSignal struct {
	Operator string `json:"operator"`
}

type ShipOrderSignal struct {
	TrackingNumber string `json:"trackingNumber"`
	Carrier        string `json:"carrier"`
	Operator       string `json:"operator"`
}

type RefundSignal struct {
	Amount   float64 `json:"amount"`
	Reason   string  `json:"reason"`
	Operator string  `json:"operator"`
}

type CancelOrderSignal struct {
	Reason   string `json:"reason"`
	Operator string `json:"operator"`
}

type SimpleSignal struct {
	Operator string `json:"operator"`
}

// OrderFulfillmentWorkflow demonstrates a state-driven MarkDoc UI for ops teams.
// This workflow shows how Cadence Web queries can replace custom admin panels.
func OrderFulfillmentWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("OrderFulfillmentWorkflow started")

	// Initialize sample order
	order := Order{
		OrderID:       "ORD-2024-001234",
		CustomerName:  "Alice Johnson",
		CustomerEmail: "alice.johnson@example.com",
		Items: []OrderItem{
			{Name: "Wireless Headphones", Quantity: 2, Price: 79.99},
			{Name: "Phone Case", Quantity: 1, Price: 19.99},
		},
		TotalAmount: 179.97,
		Status:      StatusPendingPayment,
		CreatedAt:   workflow.Now(ctx),
	}

	// Action log for audit trail
	actionLog := []ActionLogEntry{
		{
			Timestamp: workflow.Now(ctx),
			Action:    "Order Created",
			Operator:  "System",
			Details:   fmt.Sprintf("Order %s created for %s", order.OrderID, order.CustomerName),
		},
	}

	// Register query handler for the ops dashboard
	workflow.SetQueryHandler(ctx, "dashboard", func() (orderDashboardFormattedResponse, error) {
		logger.Info("Responding to 'dashboard' query")
		return makeOrderDashboard(ctx, order, actionLog), nil
	})

	// Set up signal channels
	approvePaymentChan := workflow.GetSignalChannel(ctx, "approve_payment")
	rejectPaymentChan := workflow.GetSignalChannel(ctx, "reject_payment")
	markReadyChan := workflow.GetSignalChannel(ctx, "mark_ready_to_ship")
	shipOrderChan := workflow.GetSignalChannel(ctx, "ship_order")
	refundChan := workflow.GetSignalChannel(ctx, "issue_refund")
	cancelChan := workflow.GetSignalChannel(ctx, "cancel_order")
	deliveredChan := workflow.GetSignalChannel(ctx, "mark_delivered")

	// Main workflow loop - process signals until terminal state
	for !isTerminalState(order.Status) {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(approvePaymentChan, func(ch workflow.Channel, ok bool) {
			var signal ApprovePaymentSignal
			ch.Receive(ctx, &signal)
			if order.Status == StatusPendingPayment {
				order.Status = StatusPaymentApproved
				actionLog = append(actionLog, ActionLogEntry{
					Timestamp: workflow.Now(ctx),
					Action:    "Payment Approved",
					Operator:  getOperator(signal.Operator),
					Details:   fmt.Sprintf("Payment of $%.2f approved", order.TotalAmount),
				})
				logger.Info("Payment approved", zap.String("operator", signal.Operator))
			}
		})

		selector.AddReceive(rejectPaymentChan, func(ch workflow.Channel, ok bool) {
			var signal RejectPaymentSignal
			ch.Receive(ctx, &signal)
			if order.Status == StatusPendingPayment {
				order.Status = StatusCancelled
				actionLog = append(actionLog, ActionLogEntry{
					Timestamp: workflow.Now(ctx),
					Action:    "Payment Rejected",
					Operator:  getOperator(signal.Operator),
					Details:   fmt.Sprintf("Reason: %s", signal.Reason),
				})
				logger.Info("Payment rejected", zap.String("reason", signal.Reason))
			}
		})

		selector.AddReceive(markReadyChan, func(ch workflow.Channel, ok bool) {
			var signal SimpleSignal
			ch.Receive(ctx, &signal)
			if order.Status == StatusPaymentApproved {
				order.Status = StatusReadyToShip
				actionLog = append(actionLog, ActionLogEntry{
					Timestamp: workflow.Now(ctx),
					Action:    "Marked Ready to Ship",
					Operator:  getOperator(signal.Operator),
					Details:   "Order prepared and ready for shipping",
				})
				logger.Info("Order marked ready to ship")
			}
		})

		selector.AddReceive(shipOrderChan, func(ch workflow.Channel, ok bool) {
			var signal ShipOrderSignal
			ch.Receive(ctx, &signal)
			if order.Status == StatusReadyToShip {
				order.Status = StatusShipped
				order.TrackingNum = signal.TrackingNumber
				order.Carrier = signal.Carrier
				actionLog = append(actionLog, ActionLogEntry{
					Timestamp: workflow.Now(ctx),
					Action:    "Order Shipped",
					Operator:  getOperator(signal.Operator),
					Details:   fmt.Sprintf("Carrier: %s, Tracking: %s", signal.Carrier, signal.TrackingNumber),
				})
				logger.Info("Order shipped", zap.String("tracking", signal.TrackingNumber))
			}
		})

		selector.AddReceive(refundChan, func(ch workflow.Channel, ok bool) {
			var signal RefundSignal
			ch.Receive(ctx, &signal)
			if order.Status == StatusPaymentApproved {
				order.Status = StatusRefunded
				order.RefundAmount = signal.Amount
				order.RefundReason = signal.Reason
				actionLog = append(actionLog, ActionLogEntry{
					Timestamp: workflow.Now(ctx),
					Action:    "Refund Issued",
					Operator:  getOperator(signal.Operator),
					Details:   fmt.Sprintf("Amount: $%.2f, Reason: %s", signal.Amount, signal.Reason),
				})
				logger.Info("Refund issued", zap.Float64("amount", signal.Amount))
			}
		})

		selector.AddReceive(cancelChan, func(ch workflow.Channel, ok bool) {
			var signal CancelOrderSignal
			ch.Receive(ctx, &signal)
			if order.Status == StatusReadyToShip {
				order.Status = StatusCancelled
				actionLog = append(actionLog, ActionLogEntry{
					Timestamp: workflow.Now(ctx),
					Action:    "Order Cancelled",
					Operator:  getOperator(signal.Operator),
					Details:   fmt.Sprintf("Reason: %s", signal.Reason),
				})
				logger.Info("Order cancelled", zap.String("reason", signal.Reason))
			}
		})

		selector.AddReceive(deliveredChan, func(ch workflow.Channel, ok bool) {
			var signal SimpleSignal
			ch.Receive(ctx, &signal)
			if order.Status == StatusShipped {
				order.Status = StatusDelivered
				actionLog = append(actionLog, ActionLogEntry{
					Timestamp: workflow.Now(ctx),
					Action:    "Order Delivered",
					Operator:  getOperator(signal.Operator),
					Details:   "Package confirmed delivered to customer",
				})
				logger.Info("Order marked as delivered")
			}
		})

		selector.Select(ctx)
	}

	logger.Info("OrderFulfillmentWorkflow completed", zap.String("finalStatus", order.Status))
	return nil
}

func isTerminalState(status string) bool {
	return status == StatusDelivered || status == StatusCancelled || status == StatusRefunded
}

func getOperator(operator string) string {
	if operator == "" {
		return "ops-user"
	}
	return operator
}

func makeOrderDashboard(ctx workflow.Context, order Order, actionLog []ActionLogEntry) orderDashboardFormattedResponse {
	type P map[string]interface{}

	markdownTemplate, err := template.New("").Parse(`
## 🛒 Order Dashboard

> **Your admin panel** - manage orders directly from Cadence Web.

---

### ⚡ Available Actions
{{.actionButtons}}

---

### 📋 Order Details

| Field | Value |
|-------|-------|
| **Order ID** | {{.orderID}} |
| **Customer** | {{.customerName}} |
| **Email** | {{.customerEmail}} |
| **Created** | {{.createdAt}} |
| **Status** | {{.statusBadge}} |
{{if .trackingNum}} **Tracking: {{.carrier}} - {{.trackingNum}}** {{end}}
{{if .refundAmount}}**Refund: ${{.refundAmount}}** {{end}}

### 📦 Order Items

| Item | Qty | Price | Subtotal |
|------|-----|-------|----------|
{{.itemsTable}}

**Total: ${{.totalAmount}}**

---

### ⏱️ Action History

| Timestamp | Action | Operator | Details |
|-----------|--------|----------|---------|
{{.actionHistory}}

---

*Click query "Run" button again to see updated status after taking an action.*
	`)
	if err != nil {
		panic("Failed to parse template: " + err.Error())
	}

	var markdown bytes.Buffer
	err = markdownTemplate.Execute(&markdown, P{
		"orderID":       order.OrderID,
		"customerName":  order.CustomerName,
		"customerEmail": order.CustomerEmail,
		"createdAt":     order.CreatedAt.Format("2006-01-02 15:04:05"),
		"statusBadge":   getStatusBadge(order.Status),
		"trackingNum":   order.TrackingNum,
		"carrier":       order.Carrier,
		"refundAmount":  fmt.Sprintf("%.2f", order.RefundAmount),
		"refundReason":  order.RefundReason,
		"totalAmount":   fmt.Sprintf("%.2f", order.TotalAmount),
		"itemsTable":    makeItemsTable(order.Items),
		"actionButtons": makeActionButtons(ctx, order),
		"actionHistory": makeActionHistory(actionLog),
	})
	if err != nil {
		panic("Failed to execute template: " + err.Error())
	}

	return orderDashboardFormattedResponse{
		CadenceResponseType: "formattedData",
		Format:              "text/markdown",
		Data:                markdown.String(),
	}
}

func getStatusBadge(status string) string {
	badges := map[string]string{
		StatusPendingPayment:  "🟡 **Pending Payment**",
		StatusPaymentApproved: "🟢 **Payment Approved**",
		StatusReadyToShip:     "📦 **Ready to Ship**",
		StatusShipped:         "🚚 **Shipped**",
		StatusDelivered:       "✅ **Delivered**",
		StatusCancelled:       "❌ **Cancelled**",
		StatusRefunded:        "💰 **Refunded**",
	}
	if badge, ok := badges[status]; ok {
		return badge
	}
	return status
}

func makeItemsTable(items []OrderItem) string {
	table := ""
	for _, item := range items {
		subtotal := float64(item.Quantity) * item.Price
		table += fmt.Sprintf("| %s | %d | $%.2f | $%.2f |\n", item.Name, item.Quantity, item.Price, subtotal)
	}
	return table
}

func makeActionButtons(ctx workflow.Context, order Order) string {
	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
	runID := workflow.GetInfo(ctx).WorkflowExecution.RunID

	var buttons string

	switch order.Status {
	case StatusPendingPayment:
		buttons = fmt.Sprintf(`
**Payment Review:**

{%% signal 
	signalName="approve_payment" 
	label="✓ Approve Payment"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="%s"
	runId="%s"
	input={"operator":"ops-user"}
/%%}
{%% signal 
	signalName="reject_payment" 
	label="✗ Reject: Policy Violation"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="%s"
	runId="%s"
	input={"reason":"Policy Violation","operator":"ops-user"}
/%%}
{%% signal 
	signalName="reject_payment" 
	label="✗ Reject: Fraud Suspected"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="%s"
	runId="%s"
	input={"reason":"Fraud Suspected","operator":"ops-user"}
/%%}
{%% signal 
	signalName="reject_payment" 
	label="✗ Reject: Customer Request"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="%s"
	runId="%s"
	input={"reason":"Customer Request","operator":"ops-user"}
/%%}
`, workflowID, runID, workflowID, runID, workflowID, runID, workflowID, runID)

	case StatusPaymentApproved:
		buttons = fmt.Sprintf(`
**Fulfillment Actions:**

{%% signal 
	signalName="mark_ready_to_ship" 
	label="📦 Mark Ready to Ship"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="%s"
	runId="%s"
	input={"operator":"ops-user"}
/%%}

**Refund Options:**

{%% signal 
	signalName="issue_refund" 
	label="💰 Full Refund ($%.2f)"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="%s"
	runId="%s"
	input={"amount":%.2f,"reason":"Full refund requested","operator":"ops-user"}
/%%}
{%% signal 
	signalName="issue_refund" 
	label="💰 Partial Refund (50%%)"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="%s"
	runId="%s"
	input={"amount":%.2f,"reason":"Partial refund - customer goodwill","operator":"ops-user"}
/%%}
`, workflowID, runID, order.TotalAmount, workflowID, runID, order.TotalAmount, workflowID, runID, order.TotalAmount/2)

	case StatusReadyToShip:
		buttons = fmt.Sprintf(`
**Shipping Options:**

{%% signal 
	signalName="ship_order" 
	label="🚚 Ship via UPS"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="%s"
	runId="%s"
	input={"trackingNumber":"1Z999AA10123456784","carrier":"UPS","operator":"ops-user"}
/%%}
{%% signal 
	signalName="ship_order" 
	label="🚚 Ship via FedEx"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="%s"
	runId="%s"
	input={"trackingNumber":"794644790126","carrier":"FedEx","operator":"ops-user"}
/%%}
{%% signal 
	signalName="ship_order" 
	label="🚚 Ship via USPS"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="%s"
	runId="%s"
	input={"trackingNumber":"9400111899223456789012","carrier":"USPS","operator":"ops-user"}
/%%}

**Cancel Order:**

{%% signal 
	signalName="cancel_order" 
	label="❌ Cancel Order"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="%s"
	runId="%s"
	input={"reason":"Cancelled before shipping","operator":"ops-user"}
/%%}
`, workflowID, runID, workflowID, runID, workflowID, runID, workflowID, runID)

	case StatusShipped:
		buttons = fmt.Sprintf(`
**Delivery Confirmation:**

{%% signal 
	signalName="mark_delivered" 
	label="✅ Mark as Delivered"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="%s"
	runId="%s"
	input={"operator":"ops-user"}
/%%}
`, workflowID, runID)

	default:
		buttons = `
*No actions available - order has been completed.*
`
	}

	return buttons
}

func makeActionHistory(actionLog []ActionLogEntry) string {
	history := ""
	for _, entry := range actionLog {
		history += fmt.Sprintf("| %s | %s | %s | %s |\n",
			entry.Timestamp.Format("15:04:05"),
			entry.Action,
			entry.Operator,
			entry.Details)
	}
	return history
}
