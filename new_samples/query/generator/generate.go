package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	// Define the data for HelloWorld
	data := template.TemplateData{
		SampleName: "MDX Query Workflow",
		Workflows:  []string{"MDXQueryWorkflow"},
		Activities: []string{"MDXQueryActivity"},
	}

	template.GenerateAll(data)
}

// Implement custom generator below