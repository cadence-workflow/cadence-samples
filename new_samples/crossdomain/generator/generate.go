package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Cross Domain",
		Workflows:  []string{"CrossDomainWorkflow", "ChildDomainWorkflow"},
		Activities: []string{"ChildDomainActivity"},
	}
	template.GenerateAll(data)
}

