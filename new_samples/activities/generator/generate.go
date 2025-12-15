package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	// Define the data for Dynamic Activity sample
	data := template.TemplateData{
		SampleName: "Dynamic Activity",
		Workflows:  []string{"DynamicWorkflow"},
		Activities: []string{"DynamicGreetingActivity"},
	}

	template.GenerateAll(data)
}

// Implement custom generator below

