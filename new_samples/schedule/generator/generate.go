package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Schedule",
		Workflows:  []string{"scheduledWorkflow"},
		Activities: []string{"scheduledActivity"},
	}

	template.GenerateAll(data)
}
