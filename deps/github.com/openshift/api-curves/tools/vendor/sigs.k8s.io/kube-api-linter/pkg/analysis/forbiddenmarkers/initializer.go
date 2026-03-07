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
package forbiddenmarkers

import (
	"golang.org/x/tools/go/analysis"
	"k8s.io/apimachinery/pkg/util/sets"
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
		false,
		validateConfig,
	)
}

func initAnalyzer(cfg *Config) (*analysis.Analyzer, error) {
	return newAnalyzer(cfg), nil
}

// validateConfig implements validation of the conditions linter config.
func validateConfig(cfg *Config, fldPath *field.Path) field.ErrorList {
	if cfg == nil {
		return field.ErrorList{field.Required(fldPath, "configuration is required for the forbiddenmarkers linter when it is enabled")}
	}

	fieldErrors := field.ErrorList{}

	fieldErrors = append(fieldErrors, validateMarkers(fldPath.Child("markers"), cfg.Markers...)...)

	return fieldErrors
}

func validateMarkers(fldPath *field.Path, markers ...Marker) field.ErrorList {
	if len(markers) == 0 {
		return field.ErrorList{field.Required(fldPath, "must contain at least one forbidden marker")}
	}

	fieldErrors := field.ErrorList{}

	knownMarkers := sets.New[string]()

	for i, marker := range markers {
		indexPath := fldPath.Index(i)
		if knownMarkers.Has(marker.Identifier) {
			fieldErrors = append(fieldErrors, field.Duplicate(indexPath.Child("identifier"), marker.Identifier))
			continue
		}

		knownMarkers.Insert(marker.Identifier)

		fieldErrors = append(fieldErrors, validateRuleSets(indexPath.Child("ruleSets"), marker.RuleSets...)...)
	}

	return fieldErrors
}

func validateRuleSets(fldPath *field.Path, ruleSets ...RuleSet) field.ErrorList {
	if len(ruleSets) == 0 {
		return field.ErrorList{}
	}

	fieldErrors := field.ErrorList{}

	for i, ruleSet := range ruleSets {
		fieldErrors = append(fieldErrors, validateAttributes(fldPath.Index(i).Child("attributes"), ruleSet.Attributes...)...)
	}

	return fieldErrors
}

func validateAttributes(fldPath *field.Path, attributes ...MarkerAttribute) field.ErrorList {
	if len(attributes) == 0 {
		return field.ErrorList{field.Required(fldPath, "must contain at least one attribute")}
	}

	fieldErrors := field.ErrorList{}

	knownAttributes := sets.New[string]()

	for i, attribute := range attributes {
		indexPath := fldPath.Index(i)
		if knownAttributes.Has(attribute.Name) {
			fieldErrors = append(fieldErrors, field.Duplicate(indexPath.Child("name"), attribute.Name))
			continue
		}

		knownAttributes.Insert(attribute.Name)

		fieldErrors = append(fieldErrors, validateValues(indexPath.Child("values"), attribute.Values...)...)
	}

	return fieldErrors
}

func validateValues(fldPath *field.Path, values ...string) field.ErrorList {
	if len(values) == 0 {
		return field.ErrorList{}
	}

	fieldErrors := field.ErrorList{}

	knownAttributes := sets.New[string]()

	for i, value := range values {
		indexPath := fldPath.Index(i)
		if knownAttributes.Has(value) {
			fieldErrors = append(fieldErrors, field.Duplicate(indexPath, value))
			continue
		}

		knownAttributes.Insert(value)
	}

	return fieldErrors
}
