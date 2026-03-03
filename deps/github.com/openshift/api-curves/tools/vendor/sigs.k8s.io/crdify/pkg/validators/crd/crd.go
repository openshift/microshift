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

package crd

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/crdify/pkg/validations"
)

// Validator validates Kubernetes CustomResourceDefinitions using the configured validations.
type Validator struct {
	comparators []validations.Comparator[apiextensionsv1.CustomResourceDefinition]
}

// ValidatorOption configures a Validator.
type ValidatorOption func(*Validator)

// WithComparators configures a Validator with the provided CustomResourceDefinition Comparators.
// Each call to WithComparators is a replacement, not additive.
func WithComparators(comparators ...validations.Comparator[apiextensionsv1.CustomResourceDefinition]) ValidatorOption {
	return func(v *Validator) {
		v.comparators = comparators
	}
}

// New returns a new Validator for validating an old and new CustomResourceDefinition
// configured with the provided ValidatorOptions.
func New(opts ...ValidatorOption) *Validator {
	validator := &Validator{
		comparators: []validations.Comparator[apiextensionsv1.CustomResourceDefinition]{},
	}

	for _, opt := range opts {
		opt(validator)
	}

	return validator
}

// Validate runs the validations configured in the Validator.
func (v *Validator) Validate(a, b *apiextensionsv1.CustomResourceDefinition) []validations.ComparisonResult {
	result := []validations.ComparisonResult{}

	for _, comparator := range v.comparators {
		compResult := comparator.Compare(a, b)
		result = append(result, compResult)
	}

	return result
}
