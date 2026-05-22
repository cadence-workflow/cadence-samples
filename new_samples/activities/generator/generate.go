package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	// Define the data for Activities samples
	data := template.TemplateData{
		SampleName: "Activities",
		Workflows:  []string{"DynamicWorkflow", "ParallelBranchPickFirstWorkflow"},
		Activities: []string{"DynamicGreetingActivity", "ParallelActivity"},
	}

	template.GenerateAll(data)
}

// Implement custom generator below
