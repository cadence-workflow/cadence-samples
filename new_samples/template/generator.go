package template

import (
	"os"
	"text/template"
)

type TemplateData struct {
	SampleName string
	Workflows  []string
	Activities []string
}

func GenerateAll(data TemplateData) {
	GenerateWorker(data)
	GenerateMain(data)
	GenerateSampleReadMe(data)
	GenerateGeneratorReadMe(data)
}

func GenerateWorker(data TemplateData) {
	GenerateFile("../../template/worker.tmpl", "../worker.go", data)
	println("Generated worker.go")
}

func GenerateMain(data TemplateData) {
	GenerateFile("../../template/main.tmpl", "../main.go", data)
	println("Generated main.go")
}

func GenerateSampleReadMe(data TemplateData) {
	inputs := []string{"../../template/README.tmpl", "README_specific.md", "../../template/README_references.tmpl"}
	GenerateREADME(inputs, "../README.md", data)
}

func GenerateGeneratorReadMe(data TemplateData) {
	GenerateFile("../../template/README_generator.tmpl", "README.md", data)
	println("Generated README.md")
}

func GenerateFile(templatePath, outputPath string, data TemplateData) {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		panic("Failed to parse template " + templatePath + ": " + err.Error())
	}

	f, err := os.Create(outputPath)
	if err != nil {
		panic("Failed to create output file " + outputPath + ": " + err.Error())
	}
	defer f.Close()

	err = tmpl.Execute(f, data)
	if err != nil {
		panic("Failed to execute template: " + err.Error())
	}
}

func GenerateREADME(inputs []string, outputPath string, data TemplateData) {
	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		panic("Failed to create README file: " + err.Error())
	}
	defer f.Close()

	for _, input := range inputs {
		tmpl, err := template.ParseFiles(input)
		if err != nil {
			panic("Failed to parse README template: " + err.Error())
		}

		err = tmpl.Execute(f, data)
		if err != nil {
			panic(input + ": Failed to append README content: " + err.Error())
		}
	}

}
