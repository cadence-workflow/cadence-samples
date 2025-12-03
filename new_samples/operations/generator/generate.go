package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	// Define the data for Cancel Workflow sample
	data := template.TemplateData{
		SampleName: "Cancel Workflow",
		Workflows:  []string{"CancelWorkflow"},
		Activities: []string{"ActivityToBeCanceled", "ActivityToBeSkipped", "CleanupActivity"},
	}

	template.GenerateAll(data)
}

// Implement custom generator below

