package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Choice",
		Workflows:  []string{"ChoiceWorkflow"},
		Activities: []string{"GetOrderActivity", "ProcessAppleActivity", "ProcessBananaActivity", "ProcessOrangeActivity"},
	}
	template.GenerateAll(data)
}

