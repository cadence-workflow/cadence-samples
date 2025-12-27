package main

import (
	"context"
	"bytes"
	"strconv"
	"text/template"
	"time"

	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/workflow"
	"go.uber.org/cadence/x/blocks"
	"go.uber.org/zap"
)

const (
	CompleteSignalChan = "complete"
)

func MDXQueryWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute * 60,
		StartToCloseTimeout:    time.Minute * 60,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("MDXQueryWorkflow started")

	workflow.SetQueryHandler(ctx, "Signal", func() (blocks.QueryResponse, error) {
		logger := workflow.GetLogger(ctx)
		logger.Info("Responding to 'Signal' query")

		return makeMDXQueryResponse(ctx), nil
	})

	var complete bool
	completeChan := workflow.GetSignalChannel(ctx, CompleteSignalChan)
	for {
		s := workflow.NewSelector(ctx)
		s.AddReceive(completeChan, func(ch workflow.Channel, ok bool) {
			if ok {
				ch.Receive(ctx, &complete)
			}
			logger.Info("Signal input: " + strconv.FormatBool(complete))
		})
		s.Select(ctx)

		var result string
		err := workflow.ExecuteActivity(ctx, MDXQueryActivity, complete).Get(ctx, &result)
		if err != nil {
			return err
		}
		logger.Info("Activity result: " + result)
		if complete {
			return nil
		}
	}
}

func makeMDXQueryResponse(ctx workflow.Context) blocks.QueryResponse {
	type P map[string]interface{}

	mdxTemplate, err := template.New("").Parse(`
	## MDX Query Workflow
	
	You can use markdown as your query response, which also supports starting and signaling workflows.
	
	* Use the Complete button to complete this workflow.
	* Use the Continue button just to send a signal to continue this workflow.
	* Or you can use the "Start Another" button to start another workflow of this type.
	
	<Signal
		domain="cadence-samples"
		cluster="cluster0"
		wf-id="{{.workflowID}}"
		run-id="{{.runID}}"
		name="complete"
		input={true}
	>
		Complete
	</Signal>

	<Signal
	  domain="cadence-samples"
	  cluster="cluster0"
	  wf-id="{{.workflowID}}"
	  run-id="{{.runID}}"
	  name="complete"
	  input={false}
	>
		Continue
	</Signal>
	
	<Start
	  domain="cadence-samples"
	  workflow-type="cadence_samples.MDXQueryWorkflow"
	  task-list="cadence-samples-worker"
	  wf-id={"mdx-" + Math.floor(Math.random() * 10000000)}
	  input={undefined}
	  timeout-seconds={60}
	>
		Start Another
	</Start>
		`)
	if err != nil {
		panic("Failed to parse template: " + err.Error())
	}

	var mdx bytes.Buffer
	err = mdxTemplate.Execute(&mdx, P{
		"workflowID": workflow.GetInfo(ctx).WorkflowExecution.ID,
		"runID":      workflow.GetInfo(ctx).WorkflowExecution.RunID,
	})
	if err != nil {
		panic("Failed to execute template: " + err.Error())
	}

	return blocks.New(blocks.NewMarkdownSection(mdx.String()))
}

func MDXQueryActivity(ctx context.Context, complete bool) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("MDXQueryActivity started, a new signal has been received", zap.Bool("complete", complete))
	if complete {
		return "Workflow will complete now", nil
	}
	return "Workflow will continue to run", nil
}
