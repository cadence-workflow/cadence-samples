package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Query",
		Workflows:  []string{"QueryWorkflow", "LunchVoteWorkflow", "MarkdownQueryWorkflow", "OrderFulfillmentWorkflow", "ExpenseReportWorkflow"},
		Activities: []string{},
	}

	template.GenerateAll(data)
}

