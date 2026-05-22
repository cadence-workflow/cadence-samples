package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Consistent Query",
		Workflows:  []string{"ConsistentQueryWorkflow"},
		Activities: []string{},
	}
	template.GenerateAll(data)
}

