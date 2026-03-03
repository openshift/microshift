/*
Copyright 2025 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package jsontags

import (
	"fmt"
	"regexp"

	"golang.org/x/tools/go/analysis"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/initializer"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/registry"
)

func init() {
	registry.DefaultRegistry().RegisterLinter(Initializer())
}

// Initializer returns the AnalyzerInitializer for this
// Analyzer so that it can be added to the registry.
func Initializer() initializer.AnalyzerInitializer {
	return initializer.NewConfigurableInitializer(
		name,
		initAnalyzer,
		true,
		validateConfig,
	)
}

func initAnalyzer(jtc *JSONTagsConfig) (*analysis.Analyzer, error) {
	return newAnalyzer(jtc)
}

// validateConfig is used to validate the configuration in the config.JSONTagsConfig struct.
func validateConfig(jtc *JSONTagsConfig, fldPath *field.Path) field.ErrorList {
	if jtc == nil {
		return field.ErrorList{}
	}

	fieldErrors := field.ErrorList{}

	if jtc.JSONTagRegex != "" {
		if _, err := regexp.Compile(jtc.JSONTagRegex); err != nil {
			fieldErrors = append(fieldErrors, field.Invalid(fldPath.Child("jsonTagRegex"), jtc.JSONTagRegex, fmt.Sprintf("invalid regex: %v", err)))
		}
	}

	return fieldErrors
}
