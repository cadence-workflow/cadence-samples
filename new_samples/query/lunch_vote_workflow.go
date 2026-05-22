package main

import (
	"bytes"
	"text/template"
	"time"

	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

// lunchVoteFormattedResponse is the JSON shape Cadence Web expects for markdown query results (formattedData, text/markdown, data).
type lunchVoteFormattedResponse struct {
	CadenceResponseType string `json:"cadenceResponseType"`
	Format              string `json:"format"`
	Data                string `json:"data"`
}

// LunchVoteWorkflow demonstrates using MarkDoc query responses for interactive voting.
// Users can vote for lunch options via signal buttons rendered in the query response.
func LunchVoteWorkflow(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("LunchVoteWorkflow started")

	votes := []map[string]string{}

	workflow.SetQueryHandler(ctx, "options", func() (lunchVoteFormattedResponse, error) {
		logger := workflow.GetLogger(ctx)
		logger.Info("Responding to 'options' query")

		return makeLunchVoteResponse(ctx, votes), nil
	})

	votesChan := workflow.GetSignalChannel(ctx, "lunch_order")
	workflow.Go(ctx, func(ctx workflow.Context) {
		for {
			var vote map[string]string
			votesChan.Receive(ctx, &vote)
			votes = append(votes, vote)
			logger.Info("Vote received", zap.Any("vote", vote))
		}
	})

	// Voting period - reduced from 30 minutes for sample purposes
	err := workflow.Sleep(ctx, 10*time.Minute)
	if err != nil {
		logger.Error("Sleep failed", zap.Error(err))
		return err
	}

	logger.Info("LunchVoteWorkflow completed.", zap.Any("votes", votes))
	return nil
}

// makeLunchVoteResponse creates the MarkDoc query response for lunch voting
func makeLunchVoteResponse(ctx workflow.Context, votes []map[string]string) lunchVoteFormattedResponse {
	type P map[string]interface{}

	markdownTemplate, err := template.New("").Parse(`
## Lunch Options

We're voting on where to order lunch today. Select the option you want to vote for.

---

### Current Votes

{{.voteTable}}

### Menu Options

{{.menuTable}}

---

### Cast Your Vote

{% signal 
	signalName="lunch_order" 
	label="Farmhouse - Red Thai Curry"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="{{.workflowID}}"
	runId="{{.runID}}"
	input={"location":"Farmhouse","meal":"Red Thai Curry","requests":"spicy"}
/%}
{% signal 
	signalName="lunch_order" 
	label="Ethiopian Wat"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="{{.workflowID}}"
	runId="{{.runID}}"
	input={"location":"Ethiopian","meal":"Wat with Injera","requests":""}
/%}
{% signal 
	signalName="lunch_order" 
	label="Ler Ros - Tofu Bahn Mi"
	domain="cadence-samples"
	cluster="cluster0"
	workflowId="{{.workflowID}}"
	runId="{{.runID}}"
	input={"location":"Ler Ros","meal":"Tofu Bahn Mi","requests":""}
/%}

{% br /%}

*Vote closes when workflow times out (10 minutes)*
	`)
	if err != nil {
		panic("Failed to parse template: " + err.Error())
	}

	var markdown bytes.Buffer
	err = markdownTemplate.Execute(&markdown, P{
		"workflowID": workflow.GetInfo(ctx).WorkflowExecution.ID,
		"runID":      workflow.GetInfo(ctx).WorkflowExecution.RunID,
		"voteTable":  makeLunchVoteTable(votes),
		"menuTable":  makeLunchMenu(),
	})
	if err != nil {
		panic("Failed to execute template: " + err.Error())
	}

	return lunchVoteFormattedResponse{
		CadenceResponseType: "formattedData",
		Format:              "text/markdown",
		Data:                markdown.String(),
	}
}

// makeLunchVoteTable generates a markdown table of current votes
func makeLunchVoteTable(votes []map[string]string) string {
	if len(votes) == 0 {
		return "| Location | Meal | Requests |\n|----------|------|----------|\n| *No votes yet* | | |\n"
	}
	table := "| Location | Meal | Requests |\n|----------|------|----------|\n"
	for _, vote := range votes {
		loc := vote["location"]
		meal := vote["meal"]
		requests := vote["requests"]
		table += "| " + loc + " | " + meal + " | " + requests + " |\n"
	}
	return table
}

// makeLunchMenu generates a markdown table with menu options and images
func makeLunchMenu() string {
	options := []struct {
		image string
		desc  string
	}{
		{
			image: "https://upload.wikimedia.org/wikipedia/commons/thumb/e/e2/Red_roast_duck_curry.jpg/200px-Red_roast_duck_curry.jpg",
			desc:  "**Farmhouse - Red Thai Curry**: A dish in Thai cuisine made from curry paste, coconut milk, meat, seafood, vegetables, and herbs.",
		},
		{
			image: "https://upload.wikimedia.org/wikipedia/commons/thumb/0/0c/B%C3%A1nh_m%C3%AC_th%E1%BB%8Bt_n%C6%B0%E1%BB%9Bng.png/200px-B%C3%A1nh_m%C3%AC_th%E1%BB%8Bt_n%C6%B0%E1%BB%9Bng.png",
			desc:  "**Ler Ros - Tofu Bahn Mi**: A Vietnamese sandwich with a baguette filled with lemongrass tofu, vegetables, and fresh herbs.",
		},
		{
			image: "https://upload.wikimedia.org/wikipedia/commons/thumb/5/54/Ethiopian_wat.jpg/960px-Ethiopian_wat.jpg",
			desc:  "**Ethiopian Wat**: A traditional Ethiopian stew made from spices, vegetables, and legumes, served with injera flatbread.",
		},
	}

	table := "| Picture | Description |\n|---------|-------------|\n"
	for _, option := range options {
		table += "| ![food](" + option.image + ") | " + option.desc + " |\n"
	}
	table += "\n*(source: wikipedia)*"

	return table
}
