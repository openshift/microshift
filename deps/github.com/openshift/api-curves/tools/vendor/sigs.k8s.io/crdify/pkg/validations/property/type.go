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

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/crdify/pkg/config"
	"sigs.k8s.io/crdify/pkg/validations"
)

var (
	_ validations.Validation                                  = (*Type)(nil)
	_ validations.Comparator[apiextensionsv1.JSONSchemaProps] = (*Type)(nil)
)

const typeValidationName = "type"

// RegisterType registers the Type validation
// with the provided validation registry.
func RegisterType(registry validations.Registry) {
	registry.Register(typeValidationName, typeFactory)
}

// typeFactory is a function used to initialize a Type validation
// implementation based on the provided configuration.
func typeFactory(_ map[string]interface{}) (validations.Validation, error) {
	return &Type{}, nil
}

// Type is a Validation that can be used to identify
// incompatible changes to the type constraints of CRD properties.
type Type struct {
	enforcement config.EnforcementPolicy
}

// Name returns the name of the Type validation.
func (t *Type) Name() string {
	return typeValidationName
}

// SetEnforcement sets the EnforcementPolicy for the Type validation.
func (t *Type) SetEnforcement(policy config.EnforcementPolicy) {
	t.enforcement = policy
}

// Compare compares an old and a new JSONSchemaProps, checking for incompatible changes to the type constraints of a property.
// In order for callers to determine if diffs to a JSONSchemaProps have been handled by this validation
// the JSONSchemaProps.Type field will be reset to '""' as part of this method.
// It is highly recommended that only copies of the JSONSchemaProps to compare are provided to this method
// to prevent unintentional modifications.
func (t *Type) Compare(a, b *apiextensionsv1.JSONSchemaProps) validations.ComparisonResult {
	var err error
	if a.Type != b.Type {
		err = fmt.Errorf("%w : %q -> %q", ErrTypeChanged, a.Type, b.Type)
	}

	a.Type = ""
	b.Type = ""

	return validations.HandleErrors(t.Name(), t.enforcement, err)
}

// ErrTypeChanged represents an error state when a property type changed.
var ErrTypeChanged = errors.New("type changed")
