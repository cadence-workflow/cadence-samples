package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Dynamic Invocation",
		Workflows:  []string{"DynamicGreetingsWorkflow"},
		Activities: []string{"GetNameActivity", "GetGreetingActivity", "SayGreetingActivity"},
	}
	template.GenerateAll(data)
}

