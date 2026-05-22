package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Side Effect",
		Workflows:  []string{"SideEffectWorkflow"},
		Activities: []string{},
	}
	template.GenerateAll(data)
}

