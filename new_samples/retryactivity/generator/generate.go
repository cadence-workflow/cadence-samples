package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Retry Activity",
		Workflows:  []string{"RetryWorkflow"},
		Activities: []string{"BatchProcessingActivity"},
	}

	template.GenerateAll(data)
}

