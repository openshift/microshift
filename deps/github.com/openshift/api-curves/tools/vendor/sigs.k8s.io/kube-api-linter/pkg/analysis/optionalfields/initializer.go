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
package optionalfields

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
func Initializer() initializer.AnalyzerInitializer {
	return initializer.NewConfigurableInitializer(
		name,
		initAnalyzer,
		true,
		validateConfig,
	)
}

// Init returns the initialized Analyzer.
func initAnalyzer(ofc *OptionalFieldsConfig) (*analysis.Analyzer, error) {
	return newAnalyzer(ofc), nil
}

// validateConfig validates the configuration in the config.OptionalFieldsConfig struct.
func validateConfig(ofc *OptionalFieldsConfig, fldPath *field.Path) field.ErrorList {
	if ofc == nil {
		return field.ErrorList{}
	}

	fieldErrors := field.ErrorList{}

	fieldErrors = append(fieldErrors, validateOptionFieldsPointers(ofc.Pointers, fldPath.Child("pointers"))...)
	fieldErrors = append(fieldErrors, validateOptionFieldsOmitEmpty(ofc.OmitEmpty, fldPath.Child("omitEmpty"))...)

	return fieldErrors
}

// validateOptionFieldsPointers is used to validate the configuration in the config.OptionalFieldsPointers struct.
func validateOptionFieldsPointers(opc OptionalFieldsPointers, fldPath *field.Path) field.ErrorList {
	fieldErrors := field.ErrorList{}

	switch opc.Preference {
	case "", OptionalFieldsPointerPreferenceAlways, OptionalFieldsPointerPreferenceWhenRequired:
	default:
		fieldErrors = append(fieldErrors, field.Invalid(fldPath.Child("preference"), opc.Preference, fmt.Sprintf("invalid value, must be one of %q, %q or omitted", OptionalFieldsPointerPreferenceAlways, OptionalFieldsPointerPreferenceWhenRequired)))
	}

	switch opc.Policy {
	case "", OptionalFieldsPointerPolicySuggestFix, OptionalFieldsPointerPolicyWarn:
	default:
		fieldErrors = append(fieldErrors, field.Invalid(fldPath.Child("policy"), opc.Policy, fmt.Sprintf("invalid value, must be one of %q, %q or omitted", OptionalFieldsPointerPolicySuggestFix, OptionalFieldsPointerPolicyWarn)))
	}

	return fieldErrors
}

// validateOptionFieldsOmitEmpty is used to validate the configuration in the config.OptionalFieldsOmitEmpty struct.
func validateOptionFieldsOmitEmpty(oec OptionalFieldsOmitEmpty, fldPath *field.Path) field.ErrorList {
	fieldErrors := field.ErrorList{}

	switch oec.Policy {
	case "", OptionalFieldsOmitEmptyPolicyIgnore, OptionalFieldsOmitEmptyPolicyWarn, OptionalFieldsOmitEmptyPolicySuggestFix:
	default:
		fieldErrors = append(fieldErrors, field.Invalid(fldPath.Child("policy"), oec.Policy, fmt.Sprintf("invalid value, must be one of %q, %q, %q or omitted", OptionalFieldsOmitEmptyPolicyIgnore, OptionalFieldsOmitEmptyPolicyWarn, OptionalFieldsOmitEmptyPolicySuggestFix)))
	}

	return fieldErrors
}
