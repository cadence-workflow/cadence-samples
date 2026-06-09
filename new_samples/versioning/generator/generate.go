package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Versioning",
		Workflows:  []string{"VersionedWorkflow"},
		Activities: []string{"OldActivity", "NewActivity"},
	}
	template.GenerateAll(data)
}

