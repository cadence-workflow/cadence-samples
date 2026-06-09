package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	// Define the data for the Signal sample
	data := template.TemplateData{
		SampleName: "Signal Workflow",
		Workflows:  []string{"SimpleSignalWorkflow", "AwaitSignalWorkflow"},
		Activities: []string{"SimpleSignalActivity", "Signal1Activity", "Signal2Activity", "Signal3Activity"},
	}

	template.GenerateAll(data)
}

// Implement custom generator below
