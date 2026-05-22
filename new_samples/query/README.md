<!-- THIS IS A GENERATED FILE -->
<!-- PLEASE DO NOT EDIT -->

# Query Sample

## Prerequisites

0. Install Cadence CLI. See instruction [here](https://cadenceworkflow.io/docs/cli/).
1. Run the Cadence server:
    1. Clone the [Cadence](https://github.com/cadence-workflow/cadence) repository if you haven't done already: `git clone https://github.com/cadence-workflow/cadence.git`
    2. Run `docker compose -f docker/docker-compose.yml up` to start Cadence server
    3. See more details at https://github.com/uber/cadence/blob/master/README.md
2. Once everything is up and running in Docker, open [localhost:8088](localhost:8088) to view Cadence UI.
3. Register the `cadence-samples` domain:

```bash
cadence --domain cadence-samples domain register
```

Refresh the [domains page](http://localhost:8088/domains) from step 2 to verify `cadence-samples` is registered.

## Steps to run sample

Inside the folder this sample is defined, run the following command:

```bash
go run .
```

This will call the main function in main.go which starts the worker, which will be execute the sample workflow code

## Query Samples

This folder contains samples demonstrating how to use Cadence queries with **MarkDoc-formatted responses**. MarkDoc allows you to create interactive query responses with buttons that can signal workflows or start new workflows.

### Why This Matters for Ops Teams

Many teams build custom admin panels (using Retool, React, etc.) to manage long-running workflows because:
- The CLI requires manually formatting JSON payloads
- The generic Web UI doesn't provide context-specific actions
- Support staff need simple buttons, not JSON knowledge

**MarkDoc solves this.** Your workflow query becomes your admin panel:
- State-appropriate buttons that change based on workflow status
- Structured payloads sent with a single click
- Built-in audit trail in workflow history
- Zero additional infrastructure required

---

### Markdown Query Workflow

A basic example demonstrating MarkDoc query usage with signal buttons.

```bash
cadence --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 1000 \
  --workflow_type cadence_samples.MarkdownQueryWorkflow
```

#### How to interact

1. Go to the `cadence-samples` domain in cadence-web and click on this workflow
2. Click the **"Query"** tab
3. Select **"Signal"** from the query dropdown
4. Use the rendered buttons to control the workflow

---

### Lunch Vote Workflow

An interactive voting system demonstrating dynamic query responses.

```bash
cadence --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 600 \
  --workflow_type cadence_samples.LunchVoteWorkflow
```

#### How to vote

1. Navigate to the workflow in cadence-web
2. Click the **"Query"** tab, select **"options"**
3. Click any vote button
4. Refresh the query to see updated vote counts

---

### Order Fulfillment Workflow (Admin Panel Demo)

**This is the flagship sample.** It demonstrates how MarkDoc can replace custom admin panels for ops teams.

```bash
cadence --domain cadence-samples \
  workflow start \
  --tl cadence-samples-worker \
  --et 3600 \
  --workflow_type cadence_samples.OrderFulfillmentWorkflow
```

#### The Scenario

You're an ops team member managing e-commerce orders. Instead of building a Retool dashboard or custom React app, you use the Cadence Web query feature as your admin panel.

#### Order State Machine

```
pending_payment → payment_approved → ready_to_ship → shipped → delivered
       ↓                 ↓                 ↓
   cancelled          refunded         cancelled
```

#### How to Use

1. **Start the workflow** using the CLI command above
2. **Open Cadence Web** at `localhost:8088`
3. Navigate to `cadence-samples` domain → find your workflow
4. Click the **"Query"** tab
5. Select **"dashboard"** from the dropdown
6. **You'll see:**
   - Order details (customer, items, total)
   - Current status with visual indicator
   - State-appropriate action buttons
   - Complete action history (audit trail)

#### Walking Through the Flow

**Step 1: Payment Review**
- Status shows "🟡 Pending Payment"
- Available actions: "Approve Payment" or "Reject" (with reason options)
- Click **"✓ Approve Payment"**

**Step 2: Fulfillment**
- Refresh query - status now shows "🟢 Payment Approved"
- Available actions: "Mark Ready to Ship" or "Issue Refund"
- Click **"📦 Mark Ready to Ship"**

**Step 3: Shipping**
- Refresh query - status shows "📦 Ready to Ship"
- Available actions: Ship via UPS/FedEx/USPS, or Cancel Order
- Click **"🚚 Ship via UPS"**

**Step 4: Delivery**
- Refresh query - status shows "🚚 Shipped" with tracking number
- Available action: "Mark as Delivered"
- Click **"✅ Mark as Delivered"**

**Step 5: Complete**
- Status shows "✅ Delivered"
- No more actions available
- Full audit trail visible in Action History table

#### Key Features Demonstrated

| Feature | What It Shows |
|---------|---------------|
| **State-Driven UI** | Buttons change based on order status - you can't ship before payment approval |
| **Structured Payloads** | Shipping sends `{trackingNumber, carrier}`, refunds send `{amount, reason}` |
| **Multiple Choice via Buttons** | Rejection reasons as separate buttons - no JSON formatting needed |
| **Audit Trail** | Every action recorded with timestamp, operator, and details |
| **Business Context** | Order details, items, amounts displayed alongside actions |

#### The Value Proposition

> **"Your workflow IS your admin panel."**

Instead of:
- Building a Retool dashboard
- Maintaining a separate React app
- Teaching ops to format JSON

You get:
- Interactive UI generated from workflow state
- Actions that enforce valid state transitions
- Automatic audit logging in workflow history
- Zero additional infrastructure

---

### MarkDoc Syntax Reference

MarkDoc uses special tags for interactive elements:

**Signal Button:**
```
{% signal 
    signalName="approve_payment" 
    label="Approve"
    domain="cadence-samples"
    workflowId="your-workflow-id"
    runId="your-run-id"
    input={"key":"value"}
/%}
```

**Start Workflow Button:**
```
{% start
    workflowType="cadence_samples.MyWorkflow" 
    label="Start New"
    domain="cadence-samples"
    taskList="cadence-samples-worker"
    workflowId="new-workflow-id"
    timeoutSeconds=60
/%}
```

**Other Tags:**
- `{% br /%}` - Line break
- `{% image src="url" alt="text" /%}` - Image

## References

* The website: https://cadenceworkflow.io
* Cadence's server: https://github.com/uber/cadence
* Cadence's Go client: https://github.com/uber-go/cadence-client

