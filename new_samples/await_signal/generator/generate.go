package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	// Define the data for Activities samples
	data := template.TemplateData{
		SampleName: "Await Signal",
		Workflows:  []string{"AwaitSignalWorkflow"},
		Activities: []string{"Signal1Activity", "Signal2Activity", "Signal3Activity"},
	}

	template.GenerateAll(data)
}

// Implement custom generator below
