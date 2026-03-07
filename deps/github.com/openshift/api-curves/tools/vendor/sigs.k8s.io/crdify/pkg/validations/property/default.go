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
	"bytes"
	"errors"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/crdify/pkg/config"
	"sigs.k8s.io/crdify/pkg/validations"
)

const defaultValidationName = "default"

var (
	_ validations.Validation                                  = (*Default)(nil)
	_ validations.Comparator[apiextensionsv1.JSONSchemaProps] = (*Default)(nil)
)

// RegisterDefault registers the Default validation
// with the provided validation registry.
func RegisterDefault(registry validations.Registry) {
	registry.Register(defaultValidationName, defaultFactory)
}

// defaultFactory is a function used to initialize a Default validation
// implementation based on the provided configuration.
func defaultFactory(_ map[string]interface{}) (validations.Validation, error) {
	return &Default{}, nil
}

// Default is a Validation that can be used to identify
// incompatible changes to the default value of CRD properties.
type Default struct {
	// enforcement is the EnforcementPolicy that this validation
	// should use when performing its validation logic
	enforcement config.EnforcementPolicy
}

// Name returns the name of the Default validation.
func (d *Default) Name() string {
	return defaultValidationName
}

// SetEnforcement sets the EnforcementPolicy for the Default validation.
func (d *Default) SetEnforcement(policy config.EnforcementPolicy) {
	d.enforcement = policy
}

// Compare compares an old and a new JSONSchemaProps, checking for changes to the default value of a property.
// In order for callers to determine if diffs to a JSONSchemaProps have been handled by this validation
// the JSONSchemaProps.Default field will be reset to 'nil' as part of this method.
// It is highly recommended that only copies of the JSONSchemaProps to compare are provided to this method
// to prevent unintentional modifications.
func (d *Default) Compare(a, b *apiextensionsv1.JSONSchemaProps) validations.ComparisonResult {
	var err error

	switch {
	case a.Default == nil && b.Default != nil:
		err = fmt.Errorf("%w : %q", ErrNetNewDefaultConstraint, string(b.Default.Raw))
	case a.Default != nil && b.Default == nil:
		err = fmt.Errorf("%w : %q", ErrRemovedDefault, string(a.Default.Raw))
	case a.Default != nil && b.Default != nil && !bytes.Equal(a.Default.Raw, b.Default.Raw):
		err = fmt.Errorf("%w : %q -> %q", ErrChangedDefault, string(a.Default.Raw), string(b.Default.Raw))
	}

	// reset values
	a.Default = nil
	b.Default = nil

	return validations.HandleErrors(d.Name(), d.enforcement, err)
}

var (
	// ErrNetNewDefaultConstraint represents an error state where a net new default was added to a property.
	ErrNetNewDefaultConstraint = errors.New("default added when there was none previously")
	// ErrRemovedDefault represents an error state where the default value was removed for a property.
	ErrRemovedDefault = errors.New("default value removed")
	// ErrChangedDefault represents an error state where the default value was changed for a property.
	ErrChangedDefault = errors.New("default value changed")
)
