package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Signal Counter",
		Workflows:  []string{"SignalCounterWorkflow"},
		Activities: []string{},
	}
	template.GenerateAll(data)
}

