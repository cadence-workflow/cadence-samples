package main

import (
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/cadence/x/blocks"
	"go.uber.org/zap"
)

/**
 * This is the blocks workflow sample that demonstrates JSON blob queries.
 */

// ApplicationName is the task list for this sample
const ApplicationName = "blocksGroup"

const blocksWorkflowName = "blocksWorkflow"

// This is an example of using the 'blocks' query response in a cadence query, in this example,
// to select the lunch option.
func blocksWorkflow(ctx workflow.Context, name string) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		HeartbeatTimeout:       time.Second * 20,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)

	votes := []map[string]string{}

	workflow.SetQueryHandler(ctx, "options", func() (blocks.QueryResponse, error) {
		logger := workflow.GetLogger(ctx)
		logger.Info("Responding to 'options' query")

		return makeResponse(votes), nil
	})

	votesChan := workflow.GetSignalChannel(ctx, "lunch_order")
	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			var vote map[string]string
			votesChan.Receive(ctx, &vote)
			votes = append(votes, vote)
		}
	})
	defer func() {
		votesChan.Close()
	}()

	err := workflow.Sleep(ctx, 30*time.Minute)
	if err != nil {
		logger.Error("Sleep failed", zap.Error(err))
		return err
	}

	logger.Info("Workflow completed.", zap.Any("Result", votes))
	return nil
}

// makeResponse creates the lunch query response payload based on the current votes
func makeResponse(votes []map[string]string) blocks.QueryResponse {
	return blocks.New(
		blocks.NewMarkdownSection("## Lunch options\nWe're voting on where to order lunch today. Select the option you want to vote for."),
		blocks.NewDivider(),
		blocks.NewMarkdownSection(makeVoteTable(votes)),
		blocks.NewMarkdownSection(makeMenu()),
		blocks.NewSignalActions(
			blocks.NewSignalButton("Farmhouse", "lunch_order", map[string]string{"location": "farmhouse - red thai curry", "requests": "spicy"}),
			blocks.NewSignalButtonWithExternalWorkflow("Ethiopian", "no_lunch_order_walk_in_person", nil, "in-person-order-workflow", ""),
			blocks.NewSignalButton("Ler Ros", "lunch_order", map[string]string{"location": "Ler Ros", "meal": "tofo Bahn Mi"}),
		),
	)
}

func makeMenu() string {

	options := []struct {
		image string
		desc  string
	}{
		{
			image: "https://upload.wikimedia.org/wikipedia/commons/thumb/e/e2/Red_roast_duck_curry.jpg/200px-Red_roast_duck_curry.jpg",
			desc:  "Farmhouse - Red Thai Curry: (Thai: แกง, romanized: kaeng, pronounced [kɛ̄ːŋ]) is a dish in Thai cuisine made from curry paste, coconut milk or water, meat, seafood, vegetables or fruit, and herbs. Curries in Thailand mainly differ from the Indian subcontinent in their use of ingredients such as fresh rhizomes, herbs, and aromatic leaves rather than a mix of dried spices.",
		},
		{
			image: "https://upload.wikimedia.org/wikipedia/commons/thumb/0/0c/B%C3%A1nh_m%C3%AC_th%E1%BB%8Bt_n%C6%B0%E1%BB%9Bng.png/200px-B%C3%A1nh_m%C3%AC_th%E1%BB%8Bt_n%C6%B0%E1%BB%9Bng.png",
			desc:  "Ler Ros: Lemongrass Tofu Bahn Mi: In Vietnamese cuisine, bánh mì, bánh mỳ or banh mi is a sandwich consisting of a baguette filled with various ingredients, most commonly including a protein such as pâté, chicken, or pork, and vegetables such as lettuce, cilantro, and cucumber.",
		},
		{
			image: "https://upload.wikimedia.org/wikipedia/commons/thumb/5/54/Ethiopian_wat.jpg/960px-Ethiopian_wat.jpg",
			desc:  "Ethiopian Wat: Wat is a traditional Ethiopian dish made from a mixture of spices, vegetables, and legumes. It is typically served with injera, a sourdough flatbread that is used to scoop up the food.",
		},
	}

	table := "|  Picture |  Description  |\n|---|----|\n"
	for _, option := range options {
		table += "| ![food](" + option.image + ") | " + option.desc + " |\n"
	}

	table += "\n\n\n(source wikipedia)"

	return table
}

func makeVoteTable(votes []map[string]string) string {
	if len(votes) == 0 {
		return "| lunch order vote | meal | requests |\n|-------|-------|-------|\n| No votes yet |\n"
	}
	table := "| lunch order vote | meal | requests |\n|-------|-------|-------|\n"
	for _, vote := range votes {

		loc := vote["location"]
		meal := vote["meal"]
		requests := vote["requests"]

		table += "| " + loc + " | " + meal + " | " + requests + " |\n"
	}

	return table
}
