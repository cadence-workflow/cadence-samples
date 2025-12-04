package main

import (
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/cadence/workflow"
)

var orderChoices = []string{"apple", "banana", "orange"}

// ChoiceWorkflow demonstrates conditional execution.
func ChoiceWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("ChoiceWorkflow started")

	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var order string
	if err := workflow.ExecuteActivity(ctx, GetOrderActivity).Get(ctx, &order); err != nil {
		return err
	}

	switch order {
	case "apple":
		workflow.ExecuteActivity(ctx, ProcessAppleActivity, order).Get(ctx, nil)
	case "banana":
		workflow.ExecuteActivity(ctx, ProcessBananaActivity, order).Get(ctx, nil)
	case "orange":
		workflow.ExecuteActivity(ctx, ProcessOrangeActivity, order).Get(ctx, nil)
	}

	logger.Info("ChoiceWorkflow completed")
	return nil
}

func GetOrderActivity() (string, error) {
	return orderChoices[rand.Intn(len(orderChoices))], nil
}
func ProcessAppleActivity(order string) error  { fmt.Println("Processing apple"); return nil }
func ProcessBananaActivity(order string) error { fmt.Println("Processing banana"); return nil }
func ProcessOrangeActivity(order string) error { fmt.Println("Processing orange"); return nil }

