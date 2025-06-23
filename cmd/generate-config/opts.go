package main

import (
	"io"
	"os"
	"text/template"

	"github.com/spf13/pflag"

	_ "embed"
)

var (
	//go:embed templates/config-file.template.yaml
	defaultTemplateString string
	//go:embed templates/custom-templates.tpl
	customTemplateBlocks string
)

type configGenOpts struct {
	fileOutput        string
	openApiFileOutput string
	fileInput         string
	templateFile      string
	templateText      string
}

func (opt *configGenOpts) BindFlags(f *pflag.FlagSet) {
	f.StringVarP(&opt.fileOutput, "output", "o", "", "output path, default is stdout")
	f.StringVarP(&opt.openApiFileOutput, "api-output", "a", "", "output path for openapi spec if desired")
	f.StringVarP(&opt.fileInput, "file", "f", "", "default is stdin")
	f.StringVarP(&opt.templateFile, "template", "t", "", "template file to use")
}

func (opt *configGenOpts) Options() error {
	opt.templateText = defaultTemplateString
	if opt.templateFile != "" {
		data, err := os.ReadFile(opt.templateFile)
		if err != nil {
			return err
		}
		opt.templateText = string(data)
	}
	return nil
}

func (opt configGenOpts) Run() error {
	yamlTemplate, err := template.New("yamlTemplate").Funcs(defaultTemplateFuncs).Parse(customTemplateBlocks)
	if err != nil {
		return err
	}

	yamlTemplate, err = yamlTemplate.Parse(opt.templateText)
	if err != nil {
		return err
	}

	var dataReader io.ReadCloser
	switch {
	case opt.fileInput != "":
		f, err := os.Open(opt.fileInput)
		if err != nil {
			return err
		}
		dataReader = f
		defer func() { _ = dataReader.Close() }()
	default:
		dataReader = os.Stdin
	}

	crdRawData, err := io.ReadAll(dataReader)
	if err != nil {
		return err
	}

	if opt.openApiFileOutput != "" {
		parser := crdParser{}
		parsedOpenApiSchema, err := parser.parseToJsonOpenAPI(crdRawData)
		if err != nil {
			return err
		}

		file, err := os.OpenFile(opt.openApiFileOutput, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}
		_, err = file.Write(parsedOpenApiSchema)
		if err != nil {
			return err
		}
		_ = file.Close()
	}

	var dataWriter io.WriteCloser
	if opt.fileOutput != "" {
		file, err := os.OpenFile(opt.fileOutput, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
		if err != nil {
			return err
		}
		dataWriter = file
		defer func() { _ = dataWriter.Close() }()
	} else {
		dataWriter = os.Stdout
	}
	return yamlTemplate.Execute(dataWriter, crdRawData)
}
