// Copyright 2025 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package served

import (
	"fmt"
	"slices"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	versionhelper "k8s.io/apimachinery/pkg/version"
	"sigs.k8s.io/crdify/pkg/config"
	"sigs.k8s.io/crdify/pkg/validations"
)

// Validator validates Kubernetes CustomResourceDefinitions using the configured validations.
type Validator struct {
	comparators          []validations.Comparator[apiextensionsv1.JSONSchemaProps]
	conversionPolicy     config.ConversionPolicy
	unhandledEnforcement config.EnforcementPolicy
}

// ValidatorOption configures a Validator.
type ValidatorOption func(*Validator)

// WithComparators configures a Validator with the provided JSONSchemaProps Comparators.
// Each call to WithComparators is a replacement, not additive.
func WithComparators(comparators ...validations.Comparator[apiextensionsv1.JSONSchemaProps]) ValidatorOption {
	return func(v *Validator) {
		v.comparators = comparators
	}
}

// WithUnhandledEnforcementPolicy sets the unhandled enforcement policy for the validator.
func WithUnhandledEnforcementPolicy(policy config.EnforcementPolicy) ValidatorOption {
	return func(v *Validator) {
		if policy == "" {
			policy = config.EnforcementPolicyError
		}

		v.unhandledEnforcement = policy
	}
}

// WithConversionPolicy sets the conversion policy for the validator.
func WithConversionPolicy(policy config.ConversionPolicy) ValidatorOption {
	return func(v *Validator) {
		if policy == "" {
			policy = config.ConversionPolicyNone
		}

		v.conversionPolicy = policy
	}
}

// New creates a new Validator to validate the served versions of an old and new CustomResourceDefinition
// configured with the provided ValidatorOptions.
func New(opts ...ValidatorOption) *Validator {
	validator := &Validator{
		comparators:          []validations.Comparator[apiextensionsv1.JSONSchemaProps]{},
		conversionPolicy:     config.ConversionPolicyNone,
		unhandledEnforcement: config.EnforcementPolicyError,
	}

	for _, opt := range opts {
		opt(validator)
	}

	return validator
}

// Validate runs the validations configured in the Validator.
func (v *Validator) Validate(a, b *apiextensionsv1.CustomResourceDefinition) map[string]map[string][]validations.ComparisonResult {
	result := map[string]map[string][]validations.ComparisonResult{}

	// If conversion webhook is specified and conversion policy is ignore, pass check
	if v.conversionPolicy == config.ConversionPolicyIgnore && b.Spec.Conversion != nil && b.Spec.Conversion.Strategy == apiextensionsv1.WebhookConverter {
		return result
	}

	aResults := v.compareVersionPairs(a)
	bResults := v.compareVersionPairs(b)
	subtractExistingIssues(bResults, aResults)

	return bResults
}

func (v *Validator) compareVersionPairs(crd *apiextensionsv1.CustomResourceDefinition) map[string]map[string][]validations.ComparisonResult {
	result := map[string]map[string][]validations.ComparisonResult{}

	for resultVersionPair, versions := range makeVersionPairs(crd) {
		result[resultVersionPair] = validations.CompareVersions(versions[0], versions[1], v.unhandledEnforcement, v.comparators...)
	}

	return result
}

func makeVersionPairs(crd *apiextensionsv1.CustomResourceDefinition) map[string][2]apiextensionsv1.CustomResourceDefinitionVersion {
	servedVersions := make([]apiextensionsv1.CustomResourceDefinitionVersion, 0, len(crd.Spec.Versions))

	for _, version := range crd.Spec.Versions {
		if version.Served {
			servedVersions = append(servedVersions, version)
		}
	}

	if len(servedVersions) < 2 {
		return nil
	}

	slices.SortFunc(servedVersions, func(a, b apiextensionsv1.CustomResourceDefinitionVersion) int {
		return versionhelper.CompareKubeAwareVersionStrings(a.Name, b.Name)
	})

	pairs := make(map[string][2]apiextensionsv1.CustomResourceDefinitionVersion, numUnidirectionalPermutations(servedVersions))

	for i, iVersion := range servedVersions[:len(servedVersions)-1] {
		for _, jVersion := range servedVersions[i+1:] {
			resultVersionPair := fmt.Sprintf("%s -> %s", iVersion.Name, jVersion.Name)
			pairs[resultVersionPair] = [2]apiextensionsv1.CustomResourceDefinitionVersion{iVersion, jVersion}
		}
	}

	return pairs
}

func numUnidirectionalPermutations[T any](in []T) int {
	n := len(in)

	return (n * (n - 1)) / 2
}

// subtractExistingIssues removes errors and warnings from b's results that are also found in a's results.
func subtractExistingIssues(b, a map[string]map[string][]validations.ComparisonResult) {
	sliceToMapByName := func(in []validations.ComparisonResult) map[string]*validations.ComparisonResult {
		out := make(map[string]*validations.ComparisonResult, len(in))

		for i := range in {
			v := &in[i]
			out[v.Name] = v
		}

		return out
	}

	for versionPair, bVersionPairResults := range b {
		aVersionPairResults, ok := a[versionPair]
		if !ok {
			// If the version pair is not found in a, that means
			// b introduced a new version, so we'll keep _all_
			// of the comparison results for this pair
			continue
		}

		for fieldPath, bFieldPathResults := range bVersionPairResults {
			aFieldPathResults, ok := aVersionPairResults[fieldPath]
			if !ok {
				// If this field path is not found in a's results
				// for this version pair, that means b introduced a new field
				// in an existing schema, so we'll keep _all_ of the comparison
				// results for this field path.
				continue
			}

			aResultMap := sliceToMapByName(aFieldPathResults)
			bResultMap := sliceToMapByName(bFieldPathResults)

			for validationName, bValidationResult := range bResultMap {
				aValidationResult, ok := aResultMap[validationName]
				if !ok {
					// If a's results do not include results for this validation,
					// that means we ran a new validation for b that we did not
					// run for a. We never intend to do that, so if that is somehow
					// the case, let's panic and say what our programmer intent was.
					panic(fmt.Sprintf("Validation %q not found in a's results for version pair %q. This should never happen because this validator uses the same validation configuration for CRDs a and b.", validationName, versionPair))
				}

				bValidationResult.Errors = slices.DeleteFunc(bValidationResult.Errors, func(bErr string) bool {
					return slices.Contains(aValidationResult.Errors, bErr)
				})
				if len(bValidationResult.Errors) == 0 {
					bValidationResult.Errors = nil
				}

				bValidationResult.Warnings = slices.DeleteFunc(bValidationResult.Warnings, func(bWarn string) bool {
					return slices.Contains(aValidationResult.Warnings, bWarn)
				})
				if len(bValidationResult.Warnings) == 0 {
					bValidationResult.Warnings = nil
				}
			}
		}
	}
}
