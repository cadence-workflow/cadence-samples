package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	// Define the data for Query samples
	data := template.TemplateData{
		SampleName: "Query",
		Workflows:  []string{"MarkdownQueryWorkflow", "LunchVoteWorkflow", "OrderFulfillmentWorkflow"},
		Activities: []string{"MarkdownQueryActivity"},
	}

	template.GenerateAll(data)
}

// Implement custom generator below