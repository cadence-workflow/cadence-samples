package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	// Define the data for HelloWorld
	data := template.TemplateData{
		SampleName: "Signal Workflow",
		Workflows:  []string{"SimpleSignalWorkflow"},
		Activities: []string{"SimpleSignalActivity"},
	}

	template.GenerateAll(data)
}

// Implement custom generator below