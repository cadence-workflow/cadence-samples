package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Branch",
		Workflows:  []string{"BranchWorkflow", "ParallelWorkflow"},
		Activities: []string{"BranchActivity"},
	}

	template.GenerateAll(data)
}

