package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	// Define the data for HelloWorld
	data := template.TemplateData{
		SampleName: "Hello World",
		Workflows:  []string{"HelloWorldWorkflow"},
		Activities: []string{"HelloWorldActivity"},
	}

	template.GenerateAll(data)
}

// Implement custom generator below