package main

import "github.com/uber-common/cadence-samples/new_samples/template"

func main() {
	// Define the data for Data samples.
	// NOTE: worker.go is hand-written (not generated) because each sample
	// requires its own DataConverter in worker options. We call the individual
	// generation functions explicitly instead of template.GenerateAll so the
	// hand-written worker.go is never clobbered.
	data := template.TemplateData{
		SampleName: "Data",
		Workflows: []string{
			"CompressionDataConverterWorkflow",
			"EncryptionDataConverterWorkflow",
			"S3OffloadDataConverterWorkflow",
		},
		Activities: []string{
			"CompressionDataConverterActivity",
			"EncryptionDataConverterActivity",
			"S3OffloadDataConverterActivity",
		},
	}

	// Explicitly skip GenerateWorker — worker.go is maintained by hand.
	template.GenerateMain(data)
	template.GenerateSampleReadMe(data)
	template.GenerateGeneratorReadMe(data)
}
