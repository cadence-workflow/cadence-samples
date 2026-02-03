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

func MarkdownQueryWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute * 60,
		StartToCloseTimeout:    time.Minute * 60,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	logger := workflow.GetLogger(ctx)
	logger.Info("MarkdownQueryWorkflow started")

	workflow.SetQueryHandler(ctx, "Signal", func() (blocks.QueryResponse, error) {
		logger := workflow.GetLogger(ctx)
		logger.Info("Responding to 'Signal' query")

		return makeMarkdownQueryResponse(ctx), nil
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
		err := workflow.ExecuteActivity(ctx, MarkdownQueryActivity, complete).Get(ctx, &result)
		if err != nil {
			return err
		}
		logger.Info("Activity result: " + result)
		if complete {
			return nil
		}
	}
}

func makeMarkdownQueryResponse(ctx workflow.Context) blocks.QueryResponse {
	type P map[string]interface{}

	markdownTemplate, err := template.New("").Parse(`
	## Markdown Query Workflow
	
	You can use markdown as your query response, which also supports starting and signaling workflows.
	
	* Use the Complete button to complete this workflow.
	* Use the Continue button just to send a signal to continue this workflow.
	* Or you can use the "Start Another" button to start another workflow of this type.
	
	{% signal 
		signalName="complete" 
		label="Complete"
		domain="cadence-samples"
		cluster="cluster0"
		workflowId="{{.workflowID}}"
		runId="{{.runID}}"
		input=true
	/%}
	{% signal
		signalName="complete" 
		label="Continue"
		domain="cadence-samples"
		cluster="cluster0"
		workflowId="{{.workflowID}}"
		runId="{{.runID}}"
		input=false
	/%}
	{% start
		workflowType="cadence_samples.MarkdownQueryWorkflow" 
		label="Start Another"
		domain="cadence-samples"
		cluster="cluster0"
		taskList="cadence-samples-worker"
		workflowId="{{.newWorkflowID}}"
		timeoutSeconds=60
	/%}
	
	{% br /%} 
	{% image src="https://cadenceworkflow.io/img/logo-white.svg" alt="Cadence Logo" height="100" /%}
		`)
	if err != nil {
		panic("Failed to parse template: " + err.Error())
	}

	var markdown bytes.Buffer
	err = markdownTemplate.Execute(&markdown, P{
		"workflowID": workflow.GetInfo(ctx).WorkflowExecution.ID,
		"runID":      workflow.GetInfo(ctx).WorkflowExecution.RunID,
		"newWorkflowID": "markdown-" + strconv.FormatInt(time.Now().UnixNano()/1000000, 10),
	})
	if err != nil {
		panic("Failed to execute template: " + err.Error())
	}

	return blocks.New(blocks.NewMarkdownSection(markdown.String()))
}

func MarkdownQueryActivity(ctx context.Context, complete bool) (string, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("MarkdownQueryActivity started, a new signal has been received", zap.Bool("complete", complete))
	if complete {
		return "Workflow will complete now", nil
	}
	return "Workflow will continue to run", nil
}
