package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Delay Start",
		Workflows:  []string{"DelayStartWorkflow"},
		Activities: []string{"DelayStartActivity"},
	}
	template.GenerateAll(data)
}
