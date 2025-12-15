package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Local Activity",
		Workflows:  []string{"LocalActivityWorkflow"},
		Activities: []string{"ProcessActivity"},
	}

	template.GenerateAll(data)
}

