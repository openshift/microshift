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
package nonpointerstructs

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

func initAnalyzer(cfg *Config) (*analysis.Analyzer, error) {
	return newAnalyzer(cfg), nil
}

func validateConfig(cfg *Config, fldPath *field.Path) field.ErrorList {
	if cfg == nil {
		return field.ErrorList{}
	}

	fieldErrors := field.ErrorList{}

	switch cfg.PreferredRequiredMarker {
	case "", markers.RequiredMarker, markers.KubebuilderRequiredMarker, markers.K8sRequiredMarker:
	default:
		fieldErrors = append(fieldErrors, field.Invalid(fldPath.Child("preferredRequiredMarker"), cfg.PreferredRequiredMarker, fmt.Sprintf("invalid value, must be one of %q, %q, %q or omitted", markers.RequiredMarker, markers.KubebuilderRequiredMarker, markers.K8sRequiredMarker)))
	}

	switch cfg.PreferredOptionalMarker {
	case "", markers.OptionalMarker, markers.KubebuilderOptionalMarker, markers.K8sOptionalMarker:
	default:
		fieldErrors = append(fieldErrors, field.Invalid(fldPath.Child("preferredOptionalMarker"), cfg.PreferredOptionalMarker, fmt.Sprintf("invalid value, must be one of %q, %q, %q or omitted", markers.OptionalMarker, markers.KubebuilderOptionalMarker, markers.K8sOptionalMarker)))
	}

	return fieldErrors
}
