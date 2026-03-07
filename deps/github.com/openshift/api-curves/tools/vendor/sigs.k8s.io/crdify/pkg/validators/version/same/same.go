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

package same

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/crdify/pkg/config"
	"sigs.k8s.io/crdify/pkg/validations"
)

// Validator validates Kubernetes CustomResourceDefinitions using the configured validations.
type Validator struct {
	comparators          []validations.Comparator[apiextensionsv1.JSONSchemaProps]
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

// New creates a new Validator to validate the same versions of an old and new CustomResourceDefinition
// configured with the provided ValidatorOptions.
func New(opts ...ValidatorOption) *Validator {
	validator := &Validator{
		comparators:          []validations.Comparator[apiextensionsv1.JSONSchemaProps]{},
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

	for _, oldVersion := range a.Spec.Versions {
		newVersion := validations.GetCRDVersionByName(b, oldVersion.Name)
		// in this case, there is nothing to compare. Generally, the removal
		// of an existing version is a breaking change. It may be considered safe
		// if there are no CRs stored at that version or migration has successfully
		// occurred. Since the safety of this varies and we don't have explicit
		// knowledge of this we assume a separate check will be in place to capture
		// this as a breaking change.
		if newVersion == nil {
			continue
		}

		result[oldVersion.Name] = validations.CompareVersions(*oldVersion.DeepCopy(), *newVersion.DeepCopy(), v.unhandledEnforcement, v.comparators...)
	}

	return result
}
