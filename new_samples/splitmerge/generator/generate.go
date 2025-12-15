package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	data := template.TemplateData{
		SampleName: "Split Merge",
		Workflows:  []string{"SplitMergeWorkflow"},
		Activities: []string{"ChunkProcessingActivity"},
	}
	template.GenerateAll(data)
}

