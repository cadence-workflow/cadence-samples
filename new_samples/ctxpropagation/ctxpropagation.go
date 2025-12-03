package main

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/cadence/.gen/go/shared"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// contextKey is an unexported type used as key for items stored in the Context object
type contextKey struct{}

// propagateKey is the key used to store the value in the Context object
var propagateKey = contextKey{}

// propagationKey is the key used by the propagator to pass values through the cadence server headers
const propagationKey = "_prop"

// Values is a struct holding values to propagate
type Values struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// propagator implements the custom context propagator
type propagator struct{}

// NewContextPropagator returns a context propagator that propagates a set of
// string key-value pairs across a workflow
func NewContextPropagator() workflow.ContextPropagator {
	return &propagator{}
}

// Inject injects values from context into headers for propagation
func (s *propagator) Inject(ctx context.Context, writer workflow.HeaderWriter) error {
	value := ctx.Value(propagateKey)
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	writer.Set(propagationKey, payload)
	return nil
}

// InjectFromWorkflow injects values from workflow context into headers
func (s *propagator) InjectFromWorkflow(ctx workflow.Context, writer workflow.HeaderWriter) error {
	value := ctx.Value(propagateKey)
	payload, err := json.Marshal(value)
	if err != nil {
		return err
	}
	writer.Set(propagationKey, payload)
	return nil
}

// Extract extracts values from headers and puts them into context
func (s *propagator) Extract(ctx context.Context, reader workflow.HeaderReader) (context.Context, error) {
	if err := reader.ForEachKey(func(key string, value []byte) error {
		if key == propagationKey {
			var values Values
			if err := json.Unmarshal(value, &values); err != nil {
				return err
			}
			ctx = context.WithValue(ctx, propagateKey, values)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ctx, nil
}

// ExtractToWorkflow extracts values from headers and puts them into workflow context
func (s *propagator) ExtractToWorkflow(ctx workflow.Context, reader workflow.HeaderReader) (workflow.Context, error) {
	if err := reader.ForEachKey(func(key string, value []byte) error {
		if key == propagationKey {
			var values Values
			if err := json.Unmarshal(value, &values); err != nil {
				return err
			}
			ctx = workflow.WithValue(ctx, propagateKey, values)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return ctx, nil
}

// SetValuesInHeader places the Values container inside the header
func SetValuesInHeader(values Values, header *shared.Header) error {
	payload, err := json.Marshal(values)
	if err == nil {
		header.Fields[propagationKey] = payload
	} else {
		return err
	}
	return nil
}

// CtxPropagationWorkflow demonstrates custom context propagation.
func CtxPropagationWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Second * 5,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	// Check if custom context was propagated to workflow
	if val := ctx.Value(propagateKey); val != nil {
		vals := val.(Values)
		logger.Info("Custom context propagated to workflow", zap.String(vals.Key, vals.Value))
	}

	var values Values
	if err := workflow.ExecuteActivity(ctx, CtxPropagationActivity).Get(ctx, &values); err != nil {
		logger.Error("Workflow failed.", zap.Error(err))
		return err
	}

	logger.Info("Context propagated to activity", zap.String(values.Key, values.Value))
	logger.Info("Workflow completed.")
	return nil
}

// CtxPropagationActivity returns the propagated values from context.
func CtxPropagationActivity(ctx context.Context) (*Values, error) {
	if val := ctx.Value(propagateKey); val != nil {
		vals := val.(Values)
		return &vals, nil
	}
	return nil, nil
}

