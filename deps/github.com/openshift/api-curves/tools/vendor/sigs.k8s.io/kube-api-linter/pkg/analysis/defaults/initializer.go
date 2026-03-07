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
package defaults

import (
	"fmt"

	"golang.org/x/tools/go/analysis"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/initializer"
	"sigs.k8s.io/kube-api-linter/pkg/analysis/registry"
	"sigs.k8s.io/kube-api-linter/pkg/markers"
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

func initAnalyzer(cfg *DefaultsConfig) (*analysis.Analyzer, error) {
	return newAnalyzer(cfg), nil
}

// validateConfig is used to validate the configuration in the DefaultsConfig struct.
func validateConfig(cfg *DefaultsConfig, fldPath *field.Path) field.ErrorList {
	if cfg == nil {
		return field.ErrorList{}
	}

	fieldErrors := field.ErrorList{}

	switch cfg.PreferredDefaultMarker {
	case "", markers.DefaultMarker, markers.KubebuilderDefaultMarker:
	default:
		fieldErrors = append(fieldErrors, field.Invalid(fldPath.Child("preferredDefaultMarker"), cfg.PreferredDefaultMarker, fmt.Sprintf("invalid value, must be one of %q, %q or omitted", markers.DefaultMarker, markers.KubebuilderDefaultMarker)))
	}

	fieldErrors = append(fieldErrors, validateOmitEmpty(cfg.OmitEmpty, fldPath.Child("omitempty"))...)
	fieldErrors = append(fieldErrors, validateOmitZero(cfg.OmitZero, fldPath.Child("omitzero"))...)

	return fieldErrors
}

// validateOmitEmpty is used to validate the configuration in the DefaultsOmitEmpty struct.
func validateOmitEmpty(oec DefaultsOmitEmpty, fldPath *field.Path) field.ErrorList {
	fieldErrors := field.ErrorList{}

	switch oec.Policy {
	case "", OmitEmptyPolicyIgnore, OmitEmptyPolicyWarn, OmitEmptyPolicySuggestFix:
	default:
		fieldErrors = append(fieldErrors, field.Invalid(fldPath.Child("policy"), oec.Policy, fmt.Sprintf("invalid value, must be one of %q, %q, %q or omitted", OmitEmptyPolicyIgnore, OmitEmptyPolicyWarn, OmitEmptyPolicySuggestFix)))
	}

	return fieldErrors
}

// validateOmitZero is used to validate the configuration in the DefaultsOmitZero struct.
func validateOmitZero(ozc DefaultsOmitZero, fldPath *field.Path) field.ErrorList {
	fieldErrors := field.ErrorList{}

	switch ozc.Policy {
	case "", OmitZeroPolicyForbid, OmitZeroPolicyWarn, OmitZeroPolicySuggestFix:
	default:
		fieldErrors = append(fieldErrors, field.Invalid(fldPath.Child("policy"), ozc.Policy, fmt.Sprintf("invalid value, must be one of %q, %q, %q or omitted", OmitZeroPolicyForbid, OmitZeroPolicyWarn, OmitZeroPolicySuggestFix)))
	}

	return fieldErrors
}
