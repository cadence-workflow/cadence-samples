package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Mutex",
		Workflows:  []string{"mutexWorkflow", "sampleWorkflowWithMutex"},
		Activities: []string{"signalWithStartMutexWorkflowActivity"},
	}
	template.GenerateAll(data)
}
