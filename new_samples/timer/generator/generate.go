package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Timer",
		Workflows:  []string{"TimerWorkflow"},
		Activities: []string{"OrderProcessingActivity", "SendEmailActivity"},
	}
	template.GenerateAll(data)
}

