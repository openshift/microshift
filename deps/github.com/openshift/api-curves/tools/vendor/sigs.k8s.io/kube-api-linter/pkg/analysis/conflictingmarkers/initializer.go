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
package conflictingmarkers

import (
	"fmt"

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

// initAnalyzer returns the initialized Analyzer.
func initAnalyzer(cfg *ConflictingMarkersConfig) (*analysis.Analyzer, error) {
	return newAnalyzer(cfg), nil
}

// validateConfig validates the configuration in the config.ConflictingMarkersConfig struct.
func validateConfig(cfg *ConflictingMarkersConfig, fldPath *field.Path) field.ErrorList {
	if cfg == nil {
		return field.ErrorList{}
	}

	fieldErrors := field.ErrorList{}

	// Validate that at least one conflict set is defined
	if len(cfg.Conflicts) == 0 {
		fieldErrors = append(fieldErrors, field.Required(fldPath.Child("conflicts"), "at least one conflict set is required"))
		return fieldErrors
	}

	nameSet := sets.New[string]()

	for i, conflictSet := range cfg.Conflicts {
		if nameSet.Has(conflictSet.Name) {
			fieldErrors = append(fieldErrors, field.Duplicate(fldPath.Child("conflicts").Index(i).Child("name"), conflictSet.Name))
			continue
		}

		fieldErrors = append(fieldErrors, validateConflictSet(conflictSet, fldPath.Child("conflicts").Index(i))...)

		nameSet.Insert(conflictSet.Name)
	}

	return fieldErrors
}

func validateConflictSet(conflictSet ConflictSet, fldPath *field.Path) field.ErrorList {
	fieldErrors := field.ErrorList{}

	if conflictSet.Name == "" {
		fieldErrors = append(fieldErrors, field.Required(fldPath.Child("name"), "name is required"))
	}

	if conflictSet.Description == "" {
		fieldErrors = append(fieldErrors, field.Required(fldPath.Child("description"), "description is required"))
	}

	if len(conflictSet.Sets) < 2 {
		fieldErrors = append(fieldErrors, field.Required(fldPath.Child("sets"), "at least 2 sets are required"))
		return fieldErrors
	}

	// Validate each set is non-empty
	for i, set := range conflictSet.Sets {
		if len(set) == 0 {
			fieldErrors = append(fieldErrors, field.Required(fldPath.Child("sets").Index(i), "set cannot be empty"))
		}
	}

	// Check for overlapping markers between any sets
	for i := range conflictSet.Sets {
		for j := i + 1; j < len(conflictSet.Sets); j++ {
			setI := sets.New(conflictSet.Sets[i]...)
			setJ := sets.New(conflictSet.Sets[j]...)

			if intersection := setI.Intersection(setJ); intersection.Len() > 0 {
				fieldErrors = append(fieldErrors, field.Invalid(
					fldPath.Child("sets"),
					conflictSet,
					fmt.Sprintf("sets %d and %d cannot contain overlapping markers: %v", i+1, j+1, sets.List(intersection))))
			}
		}
	}

	return fieldErrors
}
