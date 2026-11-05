package main

import (
	"errors"

	"go.uber.org/cadence/workflow"
)

// Saga collects compensations and runs them in reverse order to undo
// the steps that already succeeded.
type Saga struct {
	compensations []func(ctx workflow.Context) error
}

// AddCompensation registers a function that undoes a completed step.
func (s *Saga) AddCompensation(compensation func(ctx workflow.Context) error) {
	s.compensations = append(s.compensations, compensation)
}

// Compensate runs the compensations, last registered first, and returns
// the errors of any that failed.
func (s *Saga) Compensate(ctx workflow.Context) error {
	// A disconnected context lets compensations run even if the workflow was cancelled.
	disconnectedCtx, _ := workflow.NewDisconnectedContext(ctx)

	var errs []error
	for i := len(s.compensations) - 1; i >= 0; i-- {
		// Keep going on failure so every completed step gets a chance to be undone.
		if err := s.compensations[i](disconnectedCtx); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
