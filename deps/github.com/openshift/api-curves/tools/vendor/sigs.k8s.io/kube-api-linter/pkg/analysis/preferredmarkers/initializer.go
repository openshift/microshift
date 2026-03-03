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
package preferredmarkers

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

// validateConfig implements validation of the preferredmarkers linter config.
func validateConfig(cfg *Config, fldPath *field.Path) field.ErrorList {
	if cfg == nil {
		return field.ErrorList{field.Required(fldPath, "configuration is required for the preferredmarkers linter when it is enabled")}
	}

	return validateMarkers(fldPath.Child("markers"), cfg.Markers...)
}

// validateEquivalentIdentifiers validates a single marker's equivalent identifiers.
func validateEquivalentIdentifiers(
	equivalents []EquivalentIdentifier,
	fldPath *field.Path,
	knownPreferredMarkers,
	knownEquivalentMarkers sets.Set[string],
) field.ErrorList {
	if len(equivalents) == 0 {
		return field.ErrorList{field.Required(fldPath, "must contain at least one equivalent identifier")}
	}

	var (
		errs             field.ErrorList
		localEquivalents = sets.New[string]()
	)

	for i, equivalent := range equivalents {
		equivalentPath := fldPath.Index(i)
		identifier := equivalent.Identifier

		// Check for duplicates within this marker's equivalent identifiers
		if localEquivalents.Has(identifier) {
			errs = append(errs, field.Duplicate(equivalentPath.Child("identifier"), identifier))
			continue
		}

		localEquivalents.Insert(identifier)

		// Check if this equivalent identifier is already used as a preferred identifier
		if knownPreferredMarkers.Has(identifier) {
			errs = append(errs, field.Invalid(equivalentPath.Child("identifier"), identifier, "equivalent identifier cannot be the same as a preferred identifier"))
			continue
		}

		// Check if this equivalent identifier was already used in another marker's equivalent list
		if knownEquivalentMarkers.Has(identifier) {
			errs = append(errs, field.Duplicate(equivalentPath.Child("identifier"), identifier))
			continue
		}

		knownEquivalentMarkers.Insert(identifier)
	}

	return errs
}

func validateMarkers(fldPath *field.Path, markers ...Marker) field.ErrorList {
	if len(markers) == 0 {
		return field.ErrorList{field.Required(fldPath, "must contain at least one preferred marker")}
	}

	var (
		errs                   field.ErrorList
		knownPreferredMarkers  = sets.New[string]()
		knownEquivalentMarkers = sets.New[string]()
	)

	for i, marker := range markers {
		indexPath := fldPath.Index(i)

		// Check for duplicate preferred identifiers
		if knownPreferredMarkers.Has(marker.PreferredIdentifier) {
			errs = append(errs, field.Duplicate(indexPath.Child("preferredIdentifier"), marker.PreferredIdentifier))
			continue // Skip equivalent validation for duplicate preferred identifiers
		}

		knownPreferredMarkers.Insert(marker.PreferredIdentifier)

		// Validate equivalent identifiers
		errs = append(errs, validateEquivalentIdentifiers(
			marker.EquivalentIdentifiers,
			indexPath.Child("equivalentIdentifiers"),
			knownPreferredMarkers,
			knownEquivalentMarkers,
		)...)
	}

	return errs
}
