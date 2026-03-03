package assets

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"strings"
	"text/template"
)

var templateFuncs = map[string]interface{}{
	"base64": base64encode,
	"indent": indent,
	"load":   load,
}

func AddTemplateFunc(name string, fn interface{}) error {
	if _, ok := templateFuncs[name]; ok {
		return fmt.Errorf("%q already registered as template func", name)
	}
	templateFuncs[name] = fn
	return nil
}

func AddTemplateFuncOrDie(name string, fn interface{}) {
	err := AddTemplateFunc(name, fn)
	if err != nil {
		panic(err)
	}
}

func indent(indention int, v []byte) string {
	newline := "\n" + strings.Repeat(" ", indention)
	return strings.Replace(string(v), "\n", newline, -1)
}

func base64encode(v []byte) string {
	return base64.StdEncoding.EncodeToString(v)
}

func load(n string, assets map[string][]byte) []byte {
	return assets[n]
}

func renderFile(name string, tb []byte, data interface{}) ([]byte, error) {
	tmpl, err := template.New(name).Funcs(templateFuncs).Parse(string(tb))
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
