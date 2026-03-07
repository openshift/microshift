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

//nolint:dupl
package property

import (
	"cmp"
	"errors"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/crdify/pkg/config"
	"sigs.k8s.io/crdify/pkg/validations"
)

// MinOptions is an abstraction for the common
// options for all the "Minimum" related constraints
// on CRD properties.
type MinOptions struct {
	enforcement config.EnforcementPolicy
}

// MinVerification is a generic helper function for comparing
// two cmp.Ordered values. It returns an error if:
// - older value is nil and newer value is not nil
// - older and newer values are not nil, newer is greater than older.
func MinVerification[T cmp.Ordered](older, newer *T) error {
	var err error

	switch {
	case older == nil && newer != nil:
		err = fmt.Errorf("%w : %v", ErrNetNewMinimumConstraint, *newer)
	case older != nil && newer != nil && *newer > *older:
		err = fmt.Errorf("%w : %v -> %v", ErrMinimumIncreased, *older, *newer)
	}

	return err
}

var (
	// ErrNetNewMinimumConstraint represents an error state where a net new minimum constraint was added to a property.
	ErrNetNewMinimumConstraint = errors.New("minimum constraint added when there was none previously")
	// ErrMinimumIncreased represents an error state where a minimum constaint on a property was increased.
	ErrMinimumIncreased = errors.New("minimum increased")
)

var (
	_ validations.Validation                                  = (*Minimum)(nil)
	_ validations.Comparator[apiextensionsv1.JSONSchemaProps] = (*Minimum)(nil)
)

const minimumValidationName = "minimum"

// RegisterMinimum registers the Minimum validation
// with the provided validation registry.
func RegisterMinimum(registry validations.Registry) {
	registry.Register(minimumValidationName, minimumFactory)
}

// minimumFactory is a function used to initialize a Minimum validation
// implementation based on the provided configuration.
func minimumFactory(_ map[string]interface{}) (validations.Validation, error) {
	return &Minimum{}, nil
}

// Minimum is a Validation that can be used to identify
// incompatible changes to the minimum constraints of CRD properties.
type Minimum struct {
	MinOptions
}

// Name returns the name of the Minimum validation.
func (m *Minimum) Name() string {
	return minimumValidationName
}

// SetEnforcement sets the EnforcementPolicy for the Minimum validation.
func (m *Minimum) SetEnforcement(policy config.EnforcementPolicy) {
	m.enforcement = policy
}

// Compare compares an old and a new JSONSchemaProps, checking for incompatible changes to the minimum constraints of a property.
// In order for callers to determine if diffs to a JSONSchemaProps have been handled by this validation
// the JSONSchemaProps.Minimum field will be reset to 'nil' as part of this method.
// It is highly recommended that only copies of the JSONSchemaProps to compare are provided to this method
// to prevent unintentional modifications.
func (m *Minimum) Compare(a, b *apiextensionsv1.JSONSchemaProps) validations.ComparisonResult {
	err := MinVerification(a.Minimum, b.Minimum)

	a.Minimum = nil
	b.Minimum = nil

	return validations.HandleErrors(m.Name(), m.enforcement, err)
}

var (
	_ validations.Validation                                  = (*MinItems)(nil)
	_ validations.Comparator[apiextensionsv1.JSONSchemaProps] = (*MinItems)(nil)
)

const minItemsValidationName = "minItems"

// RegisterMinItems registers the MinItems validation
// with the provided validation registry.
func RegisterMinItems(registry validations.Registry) {
	registry.Register(minItemsValidationName, minItemsFactory)
}

// minItemsFactory is a function used to initialize a MinItems validation
// implementation based on the provided configuration.
func minItemsFactory(_ map[string]interface{}) (validations.Validation, error) {
	return &MinItems{}, nil
}

// MinItems is a Validation that can be used to identify
// incompatible changes to the minItems constraints of CRD properties.
type MinItems struct {
	MinOptions
}

// Name returns the name of the MinItems validation.
func (m *MinItems) Name() string {
	return minItemsValidationName
}

// SetEnforcement sets the EnforcementPolicy for the MinItems validation.
func (m *MinItems) SetEnforcement(policy config.EnforcementPolicy) {
	m.enforcement = policy
}

// Compare compares an old and a new JSONSchemaProps, checking for incompatible changes to the minItems constraints of a property.
// In order for callers to determine if diffs to a JSONSchemaProps have been handled by this validation
// the JSONSchemaProps.MinItems field will be reset to 'nil' as part of this method.
// It is highly recommended that only copies of the JSONSchemaProps to compare are provided to this method
// to prevent unintentional modifications.
func (m *MinItems) Compare(a, b *apiextensionsv1.JSONSchemaProps) validations.ComparisonResult {
	err := MinVerification(a.MinItems, b.MinItems)

	a.MinItems = nil
	b.MinItems = nil

	return validations.HandleErrors(m.Name(), m.enforcement, err)
}

var (
	_ validations.Validation                                  = (*MinLength)(nil)
	_ validations.Comparator[apiextensionsv1.JSONSchemaProps] = (*MinLength)(nil)
)

const minLengthValidationName = "minLength"

// RegisterMinLength registers the MinLength validation
// with the provided validation registry.
func RegisterMinLength(registry validations.Registry) {
	registry.Register(minLengthValidationName, minLengthFactory)
}

// minLengthFactory is a function used to initialize a MinLength validation
// implementation based on the provided configuration.
func minLengthFactory(_ map[string]interface{}) (validations.Validation, error) {
	return &MinLength{}, nil
}

// MinLength is a Validation that can be used to identify
// incompatible changes to the minLength constraints of CRD properties.
type MinLength struct {
	MinOptions
}

// Name returns the name of the MinLength validation.
func (m *MinLength) Name() string {
	return minLengthValidationName
}

// SetEnforcement sets the EnforcementPolicy for the MinLength validation.
func (m *MinLength) SetEnforcement(policy config.EnforcementPolicy) {
	m.enforcement = policy
}

// Compare compares an old and a new JSONSchemaProps, checking for incompatible changes to the minLength constraints of a property.
// In order for callers to determine if diffs to a JSONSchemaProps have been handled by this validation
// the JSONSchemaProps.MinLength field will be reset to 'nil' as part of this method.
// It is highly recommended that only copies of the JSONSchemaProps to compare are provided to this method
// to prevent unintentional modifications.
func (m *MinLength) Compare(a, b *apiextensionsv1.JSONSchemaProps) validations.ComparisonResult {
	err := MinVerification(a.MinLength, b.MinLength)

	a.MinLength = nil
	b.MinLength = nil

	return validations.HandleErrors(m.Name(), m.enforcement, err)
}

var (
	_ validations.Validation                                  = (*MinProperties)(nil)
	_ validations.Comparator[apiextensionsv1.JSONSchemaProps] = (*MinProperties)(nil)
)

const minPropertiesValidationName = "minProperties"

// RegisterMinProperties registers the MinProperties validation
// with the provided validation registry.
func RegisterMinProperties(registry validations.Registry) {
	registry.Register(minPropertiesValidationName, minPropertiesFactory)
}

// minPropertiesFactory is a function used to initialize a MinProperties validation
// implementation based on the provided configuration.
func minPropertiesFactory(_ map[string]interface{}) (validations.Validation, error) {
	return &MinProperties{}, nil
}

// MinProperties is a Validation that can be used to identify
// incompatible changes to the minProperties constraints of CRD properties.
type MinProperties struct {
	MinOptions
}

// Name returns the name of the MinProperties validation.
func (m *MinProperties) Name() string {
	return minPropertiesValidationName
}

// SetEnforcement sets the EnforcementPolicy for the MinProperties validation.
func (m *MinProperties) SetEnforcement(policy config.EnforcementPolicy) {
	m.enforcement = policy
}

// Compare compares an old and a new JSONSchemaProps, checking for incompatible changes to the minProperties constraints of a property.
// In order for callers to determine if diffs to a JSONSchemaProps have been handled by this validation
// the JSONSchemaProps.MinProperties field will be reset to 'nil' as part of this method.
// It is highly recommended that only copies of the JSONSchemaProps to compare are provided to this method
// to prevent unintentional modifications.
func (m *MinProperties) Compare(a, b *apiextensionsv1.JSONSchemaProps) validations.ComparisonResult {
	err := MinVerification(a.MinProperties, b.MinProperties)

	a.MinProperties = nil
	b.MinProperties = nil

	return validations.HandleErrors(m.Name(), m.enforcement, err)
}
