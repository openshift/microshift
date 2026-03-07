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

// MaxOptions is an abstraction for the common
// options for all the "Maximum" related constraints
// on CRD properties.
type MaxOptions struct {
	enforcement config.EnforcementPolicy
}

// MaxVerification is a generic helper function for comparing
// two cmp.Ordered values. It returns an error if:
// - older value is nil and newer value is not nil
// - older and newer values are not nil, newer is less than older.
func MaxVerification[T cmp.Ordered](older, newer *T) error {
	var err error

	switch {
	case older == nil && newer != nil:
		err = fmt.Errorf("%w : %v", ErrNetNewMaximumConstraint, *newer)
	case older != nil && newer != nil && *newer < *older:
		err = fmt.Errorf("%w : %v -> %v", ErrMaximumIncreased, *older, *newer)
	}

	return err
}

var (
	// ErrNetNewMaximumConstraint represents an error state where a net new maximum constraint was added to a property.
	ErrNetNewMaximumConstraint = errors.New("maximum constraint added when there was none previously")
	// ErrMaximumIncreased represents an error state where a maximum constaint on a property was decreased.
	ErrMaximumIncreased = errors.New("maximum decreased")
)

var (
	_ validations.Validation                                  = (*Maximum)(nil)
	_ validations.Comparator[apiextensionsv1.JSONSchemaProps] = (*Maximum)(nil)
)

const maximumValidationName = "maximum"

// RegisterMaximum registers the Maximum validation
// with the provided validation registry.
func RegisterMaximum(registry validations.Registry) {
	registry.Register(maximumValidationName, maximumFactory)
}

// maximumFactory is a function used to initialize a Maximum validation
// implementation based on the provided configuration.
func maximumFactory(_ map[string]interface{}) (validations.Validation, error) {
	return &Maximum{}, nil
}

// Maximum is a Validation that can be used to identify
// incompatible changes to the maximum constraints of CRD properties.
type Maximum struct {
	MaxOptions
}

// Name returns the name of the Maximum validation.
func (m *Maximum) Name() string {
	return maximumValidationName
}

// SetEnforcement sets the EnforcementPolicy for the Maximum validation.
func (m *Maximum) SetEnforcement(policy config.EnforcementPolicy) {
	m.enforcement = policy
}

// Compare compares an old and a new JSONSchemaProps, checking for incompatible changes to the maximum constraints of a property.
// In order for callers to determine if diffs to a JSONSchemaProps have been handled by this validation
// the JSONSchemaProps.Maximum field will be reset to 'nil' as part of this method.
// It is highly recommended that only copies of the JSONSchemaProps to compare are provided to this method
// to prevent unintentional modifications.
func (m *Maximum) Compare(a, b *apiextensionsv1.JSONSchemaProps) validations.ComparisonResult {
	err := MaxVerification(a.Maximum, b.Maximum)

	a.Maximum = nil
	b.Maximum = nil

	return validations.HandleErrors(m.Name(), m.enforcement, err)
}

var (
	_ validations.Validation                                  = (*MaxItems)(nil)
	_ validations.Comparator[apiextensionsv1.JSONSchemaProps] = (*MaxItems)(nil)
)

const maxItemsValidationName = "maxItems"

// RegisterMaxItems registers the MaxItems validation
// with the provided validation registry.
func RegisterMaxItems(registry validations.Registry) {
	registry.Register(maxItemsValidationName, maxItemsFactory)
}

// maxItemsFactory is a function used to initialize a MaxItems validation
// implementation based on the provided configuration.
func maxItemsFactory(_ map[string]interface{}) (validations.Validation, error) {
	return &MaxItems{}, nil
}

// MaxItems is a Validation that can be used to identify
// incompatible changes to the maxItems constraints of CRD properties.
type MaxItems struct {
	MaxOptions
}

// Name returns the name of the MaxItems validation.
func (m *MaxItems) Name() string {
	return maxItemsValidationName
}

// SetEnforcement sets the EnforcementPolicy for the MaxItems validation.
func (m *MaxItems) SetEnforcement(policy config.EnforcementPolicy) {
	m.enforcement = policy
}

// Compare compares an old and a new JSONSchemaProps, checking for incompatible changes to the maxItems constraints of a property.
// In order for callers to determine if diffs to a JSONSchemaProps have been handled by this validation
// the JSONSchemaProps.MaxItems field will be reset to 'nil' as part of this method.
// It is highly recommended that only copies of the JSONSchemaProps to compare are provided to this method
// to prevent unintentional modifications.
func (m *MaxItems) Compare(a, b *apiextensionsv1.JSONSchemaProps) validations.ComparisonResult {
	err := MaxVerification(a.MaxItems, b.MaxItems)

	a.MaxItems = nil
	b.MaxItems = nil

	return validations.HandleErrors(m.Name(), m.enforcement, err)
}

var (
	_ validations.Validation                                  = (*MaxLength)(nil)
	_ validations.Comparator[apiextensionsv1.JSONSchemaProps] = (*MaxLength)(nil)
)

const maxLengthValidationName = "maxLength"

// RegisterMaxLength registers the MaxLength validation
// with the provided validation registry.
func RegisterMaxLength(registry validations.Registry) {
	registry.Register(maxLengthValidationName, maxLengthFactory)
}

// maxLengthFactory is a function used to initialize a MaxLength validation
// implementation based on the provided configuration.
func maxLengthFactory(_ map[string]interface{}) (validations.Validation, error) {
	return &MaxLength{}, nil
}

// MaxLength is a Validation that can be used to identify
// incompatible changes to the maxLength constraints of CRD properties.
type MaxLength struct {
	MaxOptions
}

// Name returns the name of the MaxLength validation.
func (m *MaxLength) Name() string {
	return maxLengthValidationName
}

// SetEnforcement sets the EnforcementPolicy for the MaxLength validation.
func (m *MaxLength) SetEnforcement(policy config.EnforcementPolicy) {
	m.enforcement = policy
}

// Compare compares an old and a new JSONSchemaProps, checking for incompatible changes to the maxLength constraints of a property.
// In order for callers to determine if diffs to a JSONSchemaProps have been handled by this validation
// the JSONSchemaProps.MaxLength field will be reset to 'nil' as part of this method.
// It is highly recommended that only copies of the JSONSchemaProps to compare are provided to this method
// to prevent unintentional modifications.
func (m *MaxLength) Compare(a, b *apiextensionsv1.JSONSchemaProps) validations.ComparisonResult {
	err := MaxVerification(a.MaxLength, b.MaxLength)

	a.MaxLength = nil
	b.MaxLength = nil

	return validations.HandleErrors(m.Name(), m.enforcement, err)
}

var (
	_ validations.Validation                                  = (*MaxProperties)(nil)
	_ validations.Comparator[apiextensionsv1.JSONSchemaProps] = (*MaxProperties)(nil)
)

const maxPropertiesValidationName = "maxProperties"

// RegisterMaxProperties registers the MaxProperties validation
// with the provided validation registry.
func RegisterMaxProperties(registry validations.Registry) {
	registry.Register(maxPropertiesValidationName, maxPropertiesFactory)
}

// maxPropertiesFactory is a function used to initialize a MaxProperties validation
// implementation based on the provided configuration.
func maxPropertiesFactory(_ map[string]interface{}) (validations.Validation, error) {
	return &MaxProperties{}, nil
}

// MaxProperties is a Validation that can be used to identify
// incompatible changes to the maxProperties constraints of CRD properties.
type MaxProperties struct {
	MaxOptions
}

// Name returns the name of the MaxProperties validation.
func (m *MaxProperties) Name() string {
	return maxPropertiesValidationName
}

// SetEnforcement sets the EnforcementPolicy for the MaxProperties validation.
func (m *MaxProperties) SetEnforcement(policy config.EnforcementPolicy) {
	m.enforcement = policy
}

// Compare compares an old and a new JSONSchemaProps, checking for incompatible changes to the maxProperties constraints of a property.
// In order for callers to determine if diffs to a JSONSchemaProps have been handled by this validation
// the JSONSchemaProps.MaxProperties field will be reset to 'nil' as part of this method.
// It is highly recommended that only copies of the JSONSchemaProps to compare are provided to this method
// to prevent unintentional modifications.
func (m *MaxProperties) Compare(a, b *apiextensionsv1.JSONSchemaProps) validations.ComparisonResult {
	err := MaxVerification(a.MaxProperties, b.MaxProperties)

	a.MaxProperties = nil
	b.MaxProperties = nil

	return validations.HandleErrors(m.Name(), m.enforcement, err)
}
