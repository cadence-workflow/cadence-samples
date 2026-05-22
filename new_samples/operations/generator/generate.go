package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	// Define the data for Operations samples
	data := template.TemplateData{
		SampleName: "Operations",
		Workflows:  []string{"CancelWorkflow"},
		Activities: []string{"ActivityToBeCanceled", "CleanupActivity", "ActivityToBeSkipped"},
	}

	template.GenerateAll(data)
}

// Implement custom generator below
