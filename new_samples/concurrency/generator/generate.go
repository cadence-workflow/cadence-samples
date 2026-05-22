package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	// Define the data for Concurrency samples
	data := template.TemplateData{
		SampleName: "Concurrency",
		Workflows:  []string{"BatchWorkflow"},
		Activities: []string{"BatchActivity"},
	}

	template.GenerateAll(data)
}

// Implement custom generator below
