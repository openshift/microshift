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
package ssatags

import (
	"fmt"

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

// Init returns the initialized Analyzer.
func initAnalyzer(stc *SSATagsConfig) (*analysis.Analyzer, error) {
	return newAnalyzer(stc), nil
}

// validateConfig implements validation of the ssa tags linter config.
func validateConfig(stc *SSATagsConfig, fldPath *field.Path) field.ErrorList {
	if stc == nil {
		return field.ErrorList{}
	}

	fieldErrors := field.ErrorList{}

	switch stc.ListTypeSetUsage {
	case "", SSATagsListTypeSetUsageWarn, SSATagsListTypeSetUsageIgnore:
	default:
		fieldErrors = append(fieldErrors, field.Invalid(fldPath.Child("listTypeSetUsage"), stc.ListTypeSetUsage, fmt.Sprintf("invalid value, must be one of %q, %q or omitted", SSATagsListTypeSetUsageWarn, SSATagsListTypeSetUsageIgnore)))
	}

	return fieldErrors
}
