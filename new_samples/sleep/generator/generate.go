package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Sleep",
		Workflows:  []string{"SleepWorkflow"},
		Activities: []string{"MainSleepActivity"},
	}

	template.GenerateAll(data)
}

