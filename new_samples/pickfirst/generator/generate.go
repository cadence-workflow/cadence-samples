package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Pick First",
		Workflows:  []string{"PickFirstWorkflow"},
		Activities: []string{"RaceActivity"},
	}

	template.GenerateAll(data)
}

