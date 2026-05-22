package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Child Workflow",
		Workflows:  []string{"ParentWorkflow", "ChildWorkflow"},
		Activities: []string{},
	}

	template.GenerateAll(data)
}

