package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Tracing",
		Workflows:  []string{"TracingWorkflow"},
		Activities: []string{"TracingActivity"},
	}
	template.GenerateAll(data)
}

