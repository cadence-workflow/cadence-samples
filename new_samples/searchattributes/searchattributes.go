package main

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/activity"
	"go.uber.org/cadence/client"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

/**
 * This sample shows how to use search attributes with Cadence.
 *
 * NOTE: Custom search attributes (e.g. CustomIntField) require ElasticSearch to be
 * enabled in the Cadence server. See https://cadenceworkflow.io/docs/concepts/search-attributes/
 */

// searchAttributesWorkflow reads, upserts, and queries search attributes on a running workflow.
func searchAttributesWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("SearchAttributes workflow started")

	// Read the search attribute that was set when the workflow was started.
	info := workflow.GetInfo(ctx)
	val := info.SearchAttributes.IndexedFields["CustomIntField"]
	var currentIntValue int
	if err := client.NewValue(val).Get(&currentIntValue); err != nil {
		logger.Error("Failed to read CustomIntField", zap.Error(err))
		return err
	}
	logger.Info("Initial search attributes", zap.String("CustomIntField", strconv.Itoa(currentIntValue)))

	// Upsert: update CustomIntField and add several new attributes.
	if err := workflow.UpsertSearchAttributes(ctx, map[string]interface{}{
		"CustomIntField":      2,
		"CustomKeywordField":  "Update1",
		"CustomBoolField":     true,
		"CustomDoubleField":   3.14,
		"CustomDatetimeField": time.Date(2019, 1, 1, 0, 0, 0, 0, time.Local),
		"CustomStringField":   "String field is for full-text search. It is tokenized and cannot be used in ORDER BY.",
	}); err != nil {
		return err
	}

	if err := printSearchAttributes(workflow.GetInfo(ctx).SearchAttributes, logger); err != nil {
		return err
	}

	// Second upsert: update a single attribute.
	if err := workflow.UpsertSearchAttributes(ctx, map[string]interface{}{
		"CustomKeywordField": "Update2",
	}); err != nil {
		return err
	}

	if err := printSearchAttributes(workflow.GetInfo(ctx).SearchAttributes, logger); err != nil {
		return err
	}

	// Wait for the upsert to be indexed by ElasticSearch before querying.
	workflow.Sleep(ctx, 2*time.Second)

	// Use an activity to list workflows that match the updated search attributes.
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: 2 * time.Minute,
		StartToCloseTimeout:    2 * time.Minute,
		HeartbeatTimeout:       20 * time.Second,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)
	query := "CustomIntField=2 and CustomKeywordField='Update2' order by CustomDatetimeField DESC"
	var listResults []*shared.WorkflowExecutionInfo
	if err := workflow.ExecuteActivity(ctx, listExecutions, query).Get(ctx, &listResults); err != nil {
		logger.Error("Failed to list workflow executions", zap.Error(err))
		return err
	}

	logger.Info("Workflow completed", zap.String("matchedExecution", listResults[0].String()))
	return nil
}

func printSearchAttributes(searchAttributes *shared.SearchAttributes, logger *zap.Logger) error {
	buf := new(bytes.Buffer)
	for k, v := range searchAttributes.IndexedFields {
		var currentVal interface{}
		if err := client.NewValue(v).Get(&currentVal); err != nil {
			logger.Error("Failed to decode search attribute", zap.String("key", k), zap.Error(err))
			return err
		}
		fmt.Fprintf(buf, "%s=%v\n", k, currentVal)
	}
	logger.Info(fmt.Sprintf("Current search attributes:\n%s", buf.String()))
	return nil
}

// listExecutions is an activity that queries Cadence visibility for workflows matching the given query.
// Requires ElasticSearch to be configured on the Cadence server.
func listExecutions(ctx context.Context, query string) ([]*shared.WorkflowExecutionInfo, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Listing workflow executions", zap.String("query", query))

	cadenceClient := client.NewClient(BuildCadenceClient(), Domain, nil)

	var executions []*shared.WorkflowExecutionInfo
	var nextPageToken []byte
	for {
		resp, err := cadenceClient.ListWorkflow(ctx, &shared.ListWorkflowExecutionsRequest{
			Domain:        strPtr(Domain),
			PageSize:      int32Ptr(10),
			NextPageToken: nextPageToken,
			Query:         strPtr(query),
		})
		if err != nil {
			return nil, err
		}
		executions = append(executions, resp.Executions...)
		nextPageToken = resp.NextPageToken
		activity.RecordHeartbeat(ctx, nextPageToken)
		if len(nextPageToken) == 0 {
			break
		}
	}
	return executions, nil
}

// getSearchAttributesForStart returns the search attributes to set when starting the workflow.
func getSearchAttributesForStart() map[string]interface{} {
	return map[string]interface{}{
		"CustomIntField": 1,
	}
}

func strPtr(s string) *string   { return &s }
func int32Ptr(i int32) *int32   { return &i }
