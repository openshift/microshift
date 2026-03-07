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
package requiredfields

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

func initAnalyzer(rfc *RequiredFieldsConfig) (*analysis.Analyzer, error) {
	return newAnalyzer(rfc), nil
}

// validateConfig validates the configuration in the config.RequiredFieldsConfig struct.
func validateConfig(rfc *RequiredFieldsConfig, fldPath *field.Path) field.ErrorList {
	if rfc == nil {
		return field.ErrorList{}
	}

	fieldErrors := field.ErrorList{}

	fieldErrors = append(fieldErrors, validateRequiredFieldsPointers(rfc.Pointers, fldPath.Child("pointers"))...)
	fieldErrors = append(fieldErrors, validateRequiredFieldsOmitEmpty(rfc.OmitEmpty, fldPath.Child("omitempty"))...)
	fieldErrors = append(fieldErrors, validateRequiredFieldsOmitZero(rfc.OmitZero, fldPath.Child("omitzero"))...)

	return fieldErrors
}

// validateRequiredFieldsPointers is used to validate the configuration in the config.RequiredFieldsPointers struct.
func validateRequiredFieldsPointers(rfc RequiredFieldsPointers, fldPath *field.Path) field.ErrorList {
	fieldErrors := field.ErrorList{}

	switch rfc.Policy {
	case "", RequiredFieldsPointerPolicySuggestFix, RequiredFieldsPointerPolicyWarn:
	default:
		fieldErrors = append(fieldErrors, field.Invalid(fldPath.Child("policy"), rfc.Policy, fmt.Sprintf("invalid value, must be one of %q, %q or omitted", RequiredFieldsPointerPolicySuggestFix, RequiredFieldsPointerPolicyWarn)))
	}

	return fieldErrors
}

// validateOptionFieldsOmitEmpty is used to validate the configuration in the config.OptionalFieldsOmitEmpty struct.
func validateRequiredFieldsOmitEmpty(rfc RequiredFieldsOmitEmpty, fldPath *field.Path) field.ErrorList {
	fieldErrors := field.ErrorList{}

	switch rfc.Policy {
	case "", RequiredFieldsOmitEmptyPolicyIgnore, RequiredFieldsOmitEmptyPolicyWarn, RequiredFieldsOmitEmptyPolicySuggestFix:
	default:
		fieldErrors = append(fieldErrors, field.Invalid(fldPath.Child("policy"), rfc.Policy, fmt.Sprintf("invalid value, must be one of %q, %q, %q or omitted", RequiredFieldsOmitEmptyPolicyIgnore, RequiredFieldsOmitEmptyPolicyWarn, RequiredFieldsOmitEmptyPolicySuggestFix)))
	}

	return fieldErrors
}

// validateRequiredFieldsOmitZero is used to validate the configuration in the config.RequiredFieldsOmitZero struct.
func validateRequiredFieldsOmitZero(rfc RequiredFieldsOmitZero, fldPath *field.Path) field.ErrorList {
	fieldErrors := field.ErrorList{}

	switch rfc.Policy {
	case "", RequiredFieldsOmitZeroPolicySuggestFix, RequiredFieldsOmitZeroPolicyWarn, RequiredFieldsOmitZeroPolicyForbid:
	default:
		fieldErrors = append(fieldErrors, field.Invalid(fldPath.Child("policy"), rfc.Policy, fmt.Sprintf("invalid value, must be one of %q, %q, %q or omitted", RequiredFieldsOmitZeroPolicySuggestFix, RequiredFieldsOmitZeroPolicyWarn, RequiredFieldsOmitZeroPolicyForbid)))
	}

	return fieldErrors
}
