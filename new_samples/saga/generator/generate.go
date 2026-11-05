package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Saga",
		Workflows:  []string{"SagaWorkflow"},
		Activities: []string{"StepActivity", "CompensateStepActivity"},
	}

	template.GenerateAll(data)
}
