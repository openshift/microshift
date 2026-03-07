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

package property

import (
	"errors"
	"fmt"
	"slices"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/crdify/pkg/config"
	"sigs.k8s.io/crdify/pkg/validations"
)

const enumValidationName = "enum"

var (
	_ validations.Validation                                  = (*Enum)(nil)
	_ validations.Comparator[apiextensionsv1.JSONSchemaProps] = (*Enum)(nil)
)

// RegisterEnum registers the Enum validation
// with the provided validation registry.
func RegisterEnum(registry validations.Registry) {
	registry.Register(enumValidationName, enumFactory)
}

// enumFactory is a function used to initialize an Enum validation
// implementation based on the provided configuration.
func enumFactory(cfg map[string]interface{}) (validations.Validation, error) {
	enumCfg := &EnumConfig{}

	err := ConfigToType(cfg, enumCfg)
	if err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	err = ValidateEnumConfig(enumCfg)
	if err != nil {
		return nil, fmt.Errorf("validating enum config: %w", err)
	}

	return &Enum{EnumConfig: *enumCfg}, nil
}

// ValidateEnumConfig validates the provided EnumConfig
// setting default values where appropriate.
// Currently the defaulting behavior defaults the
// EnumConfig.AdditionPolicy to AdditionPolicyDisallow
// if it is set to the empty string ("").
func ValidateEnumConfig(in *EnumConfig) error {
	if in == nil {
		// nothing to validate
		return nil
	}

	switch in.AdditionPolicy {
	case AdditionPolicyAllow, AdditionPolicyDisallow:
		// do nothing, valid case
	case AdditionPolicy(""):
		// default to disallow
		in.AdditionPolicy = AdditionPolicyDisallow
	default:
		return fmt.Errorf("%w : %q", errUnknownAdditionPolicy, in.AdditionPolicy)
	}

	return nil
}

var errUnknownAdditionPolicy = errors.New("unknown addition policy")

// AdditionPolicy is used to represent how the Enum validation
// should determine compatibility of adding new enum values to an
// existing enum constraint.
type AdditionPolicy string

const (
	// AdditionPolicyAllow signals that adding new enum values to
	// an existing enum constraint should be considered a compatible change.
	AdditionPolicyAllow AdditionPolicy = "Allow"

	// AdditionPolicyDisallow signals that adding new enum values to
	// an existing enum constraint should be considered an incompatible change.
	AdditionPolicyDisallow AdditionPolicy = "Disallow"
)

// EnumConfig contains additional configurations for the Enum validation.
type EnumConfig struct {
	// additionPolicy is how adding enums to an existing set of
	// enums should be treated.
	// Allowed values are Allow and Disallow.
	// When set to Allow, adding new values to an existing set
	// of enums will not be flagged.
	// When set to Disallow, adding new values to an existing
	// set of enums will be flagged.
	// Defaults to Disallow.
	AdditionPolicy AdditionPolicy `json:"additionPolicy,omitempty"`
}

// Enum is a Validation that can be used to identify
// incompatible changes to the enum values of CRD properties.
type Enum struct {
	// EnumConfig is the set of additional configuration options
	EnumConfig

	// enforcement is the EnforcementPolicy that this validation
	// should use when performing its validation logic
	enforcement config.EnforcementPolicy
}

// Name returns the name of the Enum validation.
func (e *Enum) Name() string {
	return enumValidationName
}

// SetEnforcement sets the EnforcementPolicy for the Enum validation.
func (e *Enum) SetEnforcement(policy config.EnforcementPolicy) {
	e.enforcement = policy
}

// Compare compares an old and a new JSONSchemaProps, checking for incompatible changes to the enum constraints of a property.
// In order for callers to determine if diffs to a JSONSchemaProps have been handled by this validation
// the JSONSchemaProps.Enum field will be reset to 'nil' as part of this method.
// It is highly recommended that only copies of the JSONSchemaProps to compare are provided to this method
// to prevent unintentional modifications.
func (e *Enum) Compare(a, b *apiextensionsv1.JSONSchemaProps) validations.ComparisonResult {
	oldEnums := sets.New[string]()
	for _, json := range a.Enum {
		oldEnums.Insert(string(json.Raw))
	}

	newEnums := sets.New[string]()
	for _, json := range b.Enum {
		newEnums.Insert(string(json.Raw))
	}

	removedEnums := oldEnums.Difference(newEnums)
	addedEnums := newEnums.Difference(oldEnums)

	var err error

	switch {
	case oldEnums.Len() == 0 && newEnums.Len() > 0:
		newEnumSlice := newEnums.UnsortedList()
		slices.Sort(newEnumSlice)
		err = fmt.Errorf("%w : %v", ErrNetNewEnumConstraint, newEnumSlice)
	case removedEnums.Len() > 0:
		removedEnumSlice := removedEnums.UnsortedList()
		slices.Sort(removedEnumSlice)
		err = fmt.Errorf("%w : %v", ErrRemovedEnums, removedEnumSlice)
	case addedEnums.Len() > 0 && e.AdditionPolicy != AdditionPolicyAllow:
		addedEnumSlice := addedEnums.UnsortedList()
		slices.Sort(addedEnumSlice)
		err = fmt.Errorf("%w : %v", ErrAddedEnums, addedEnumSlice)
	}

	a.Enum = nil
	b.Enum = nil

	return validations.HandleErrors(e.Name(), e.enforcement, err)
}

var (
	// ErrNetNewEnumConstraint represents an error state where a net new enum constraint was added to a property.
	ErrNetNewEnumConstraint = errors.New("enum constraint added when there was none previously")
	// ErrRemovedEnums represents an error state where at least one previously allowed enum value was removed
	// from the enum constraint on a property.
	ErrRemovedEnums = errors.New("allowed enum values removed")
	// ErrAddedEnums represents an error state where at least one enum value, that was not previously allowed,
	// was added to the enum constraint on a property.
	ErrAddedEnums = errors.New("allowed enum values added")
)
