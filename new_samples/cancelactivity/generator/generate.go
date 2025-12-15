package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Cancel Activity",
		Workflows:  []string{"CancelWorkflow"},
		Activities: []string{"LongRunningActivity", "CleanupActivity", "SkippedActivity"},
	}
	template.GenerateAll(data)
}
