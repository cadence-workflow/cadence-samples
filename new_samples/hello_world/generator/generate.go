package main

import (
	"io"
	"os"
	"text/template"
)

type TemplateData struct {
	SampleName string
	Workflows  []string
	Activities []string
}

func main() {
	// Define the data for HelloWorld
	data := TemplateData{
		SampleName: "Hello World",
		Workflows:  []string{"HelloWorldWorkflow"},
		Activities: []string{"HelloWorldActivity"},
	}

	// Generate worker.go
	generateFile("../../template/worker.tmpl", "../worker.go", data)
	println("Generated worker.go")

	// Generate main.go
	generateFile("../../template/main.tmpl", "../main.go", data)
	println("Generated main.go")

	// Generate README.md (combine template + specific + references)
	generateREADME("../../template/README.tmpl", "README_specific.md", "../../template/README_references.md", "../README.md", data)
	println("Generated README.md")
}

func generateFile(templatePath, outputPath string, data TemplateData) {
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

func generateREADME(templatePath, specificPath, referencesPath, outputPath string, data TemplateData) {
	// Create output file
	f, err := os.Create(outputPath)
	if err != nil {
		panic("Failed to create README file: " + err.Error())
	}
	defer f.Close()

	// First, write the generic template part
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		panic("Failed to parse README template: " + err.Error())
	}

	err = tmpl.Execute(f, data)
	if err != nil {
		panic("Failed to execute README template: " + err.Error())
	}

	// Then, append the specific content
	specific, err := os.Open(specificPath)
	if err != nil {
		panic("Failed to open specific README content: " + err.Error())
	}
	defer specific.Close()

	_, err = io.Copy(f, specific)
	if err != nil {
		panic("Failed to append specific README content: " + err.Error())
	}

	// Finally, append the references
	references, err := os.Open(referencesPath)
	if err != nil {
		panic("Failed to open references content: " + err.Error())
	}
	defer references.Close()

	_, err = io.Copy(f, references)
	if err != nil {
		panic("Failed to append references content: " + err.Error())
	}
}


