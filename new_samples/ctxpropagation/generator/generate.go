package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Context Propagation",
		Workflows:  []string{"CtxPropagationWorkflow"},
		Activities: []string{"CtxPropagationActivity"},
	}
	template.GenerateAll(data)
}

