package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

const (
	orderChoiceApple  = "apple"
	orderChoiceBanana = "banana"
	orderChoiceCherry = "cherry"
)

var orderChoices = []string{orderChoiceApple, orderChoiceBanana, orderChoiceCherry}

// ChoiceWorkflow demonstrates conditional execution based on activity results.
// It executes different activities depending on the order type returned.
func ChoiceWorkflow(ctx workflow.Context) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("ChoiceWorkflow started")

	// Get the order type
	var orderChoice string
	err := workflow.ExecuteActivity(ctx, GetOrderActivity).Get(ctx, &orderChoice)
	if err != nil {
		return err
	}
	logger.Info("Order received", zap.String("choice", orderChoice))

	// Execute different activity based on order type
	switch orderChoice {
	case orderChoiceApple:
		err = workflow.ExecuteActivity(ctx, ProcessAppleActivity, orderChoice).Get(ctx, nil)
	case orderChoiceBanana:
		err = workflow.ExecuteActivity(ctx, ProcessBananaActivity, orderChoice).Get(ctx, nil)
	case orderChoiceCherry:
		err = workflow.ExecuteActivity(ctx, ProcessCherryActivity, orderChoice).Get(ctx, nil)
	default:
		logger.Error("Unexpected order", zap.String("choice", orderChoice))
		return errors.New("unknown order type: " + orderChoice)
	}

	if err != nil {
		return err
	}

	logger.Info("ChoiceWorkflow completed")
	return nil
}

// GetOrderActivity returns a random order type.
func GetOrderActivity() (string, error) {
	idx := rand.Intn(len(orderChoices))
	order := orderChoices[idx]
	fmt.Printf("GetOrderActivity: Order is for %s\n", order)
	return order, nil
}

// ProcessAppleActivity handles apple orders.
func ProcessAppleActivity(choice string) error {
	fmt.Printf("ProcessAppleActivity: Processing %s order\n", choice)
	return nil
}

// ProcessBananaActivity handles banana orders.
func ProcessBananaActivity(choice string) error {
	fmt.Printf("ProcessBananaActivity: Processing %s order\n", choice)
	return nil
}

// ProcessCherryActivity handles cherry orders.
func ProcessCherryActivity(choice string) error {
	fmt.Printf("ProcessCherryActivity: Processing %s order\n", choice)
	return nil
}

