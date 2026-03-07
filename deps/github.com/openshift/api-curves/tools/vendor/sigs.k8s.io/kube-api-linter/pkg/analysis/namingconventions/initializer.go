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
package namingconventions

import (
	"fmt"
	"regexp"

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
		return field.ErrorList{field.Required(fldPath, "configuration is required for the namingconventions linter when it is enabled")}
	}

	if len(cfg.Conventions) == 0 {
		return field.ErrorList{field.Required(fldPath.Child("conventions"), "at least one naming convention must be specified when the namingconventions linter is enabled")}
	}

	return validateConventions(fldPath.Child("conventions"), cfg.Conventions...)
}

func validateConventions(fldPath *field.Path, conventions ...Convention) field.ErrorList {
	fieldErrs := field.ErrorList{}

	seenConventions := sets.New[string]()

	for i, convention := range conventions {
		if seenConventions.Has(convention.Name) {
			fieldErrs = append(fieldErrs, field.Duplicate(fldPath.Index(i).Child("name"), convention.Name))
			continue
		}

		fieldErrs = append(fieldErrs, validateConvention(fldPath.Index(i), convention)...)
		seenConventions.Insert(convention.Name)
	}

	return fieldErrs
}

func validateConvention(fldPath *field.Path, convention Convention) field.ErrorList {
	fieldErrs := field.ErrorList{}

	if len(convention.Name) == 0 {
		fieldErrs = append(fieldErrs, field.Required(fldPath.Child("name"), "name is required"))
	}

	if len(convention.ViolationMatcher) == 0 {
		fieldErrs = append(fieldErrs, field.Required(fldPath.Child("violationMatcher"), "violationMatcher is required"))
	}

	matcher, err := regexp.Compile(convention.ViolationMatcher)
	if err != nil {
		fieldErrs = append(fieldErrs, field.Invalid(fldPath.Child("violationMatcher"), convention.ViolationMatcher, fmt.Sprintf("violationMatcher regular expression failed to compile: %s", err.Error())))
	}

	if convention.Message == "" {
		fieldErrs = append(fieldErrs, field.Required(fldPath.Child("message"), "message is required"))
	}

	fieldErrs = append(fieldErrs, validateOperation(fldPath.Child("operation"), convention.Operation)...)
	fieldErrs = append(fieldErrs, validateReplace(fldPath.Child("replacement"), convention.Operation, matcher, convention.Replacement)...)

	return fieldErrs
}

func validateOperation(fldPath *field.Path, operation Operation) field.ErrorList {
	allowedOperations := sets.New(OperationDrop, OperationInform, OperationReplacement, OperationDropField)

	if len(operation) == 0 {
		return field.ErrorList{field.Required(fldPath, "operation is required")}
	}

	if len(operation) > 0 && !allowedOperations.Has(operation) {
		return field.ErrorList{field.Invalid(fldPath, operation, fmt.Sprintf("operation must be one of %q, %q, %q, or %q", OperationInform, OperationDrop, OperationDropField, OperationReplacement))}
	}

	return field.ErrorList{}
}

func validateReplace(fldPath *field.Path, operation Operation, matcher *regexp.Regexp, replace string) field.ErrorList {
	if (len(replace) > 0 && operation != OperationReplacement) || (len(replace) == 0 && operation == OperationReplacement) {
		return field.ErrorList{field.Invalid(fldPath, replace, "replacement must be specified when operation is 'Replacement' and is forbidden otherwise")}
	}

	if len(replace) > 0 && operation == OperationReplacement && matcher.MatchString(replace) {
		return field.ErrorList{field.Invalid(fldPath, replace, "replacement must not be matched by violationMatcher")}
	}

	return field.ErrorList{}
}
