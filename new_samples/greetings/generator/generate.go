package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Greetings",
		Workflows:  []string{"GreetingsWorkflow"},
		Activities: []string{"GetGreetingActivity", "GetNameActivity", "SayGreetingActivity"},
	}

	template.GenerateAll(data)
}

