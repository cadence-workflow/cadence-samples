package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Search Attributes",
		Workflows:  []string{"searchAttributesWorkflow"},
		Activities: []string{"listExecutions"},
	}
	template.GenerateAll(data)
}
