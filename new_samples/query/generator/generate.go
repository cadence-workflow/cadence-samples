package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	// Define the data for HelloWorld
	data := template.TemplateData{
		SampleName: "Markdown Query Workflow",
		Workflows:  []string{"MarkdownQueryWorkflow"},
		Activities: []string{"MarkdownQueryActivity"},
	}

	template.GenerateAll(data)
}

// Implement custom generator below