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

const descriptionValidationName = "description"

var (
	_ validations.Validation                                  = (*Description)(nil)
	_ validations.Comparator[apiextensionsv1.JSONSchemaProps] = (*Description)(nil)
)

// RegisterDescription registers the Description validation
// with the provided validation registry.
func RegisterDescription(registry validations.Registry) {
	registry.Register(descriptionValidationName, descriptionFactory)
}

// descriptionFactory is a function used to initialize a Description validation
// implementation based on the provided configuration.
func descriptionFactory(_ map[string]interface{}) (validations.Validation, error) {
	return &Description{}, nil
}

// Description is a Validation that can be used to identify
// incompatible changes to the description of CRD properties.
type Description struct {
	// enforcement is the EnforcementPolicy that this validation
	// should use when performing its validation logic
	enforcement config.EnforcementPolicy
}

// Name returns the name of the Description validation.
func (d *Description) Name() string {
	return descriptionValidationName
}

// SetEnforcement sets the EnforcementPolicy for the Description validation.
func (d *Description) SetEnforcement(policy config.EnforcementPolicy) {
	d.enforcement = policy
}

// Compare compares an old and a new JSONSchemaProps, checking for changes to the description of a property.
// In order for callers to determine if diffs to a JSONSchemaProps have been handled by this validation
// the JSONSchemaProps.Description field will be reset to '""' as part of this method.
// It is highly recommended that only copies of the JSONSchemaProps to compare are provided to this method
// to prevent unintentional modifications.
func (d *Description) Compare(a, b *apiextensionsv1.JSONSchemaProps) validations.ComparisonResult {
	var err error

	if a.Description != b.Description {
		err = fmt.Errorf("%w : %q -> %q", ErrChangedDescription, a.Description, b.Description)
	}

	// reset values
	a.Description = ""
	b.Description = ""

	return validations.HandleErrors(d.Name(), d.enforcement, err)
}

// ErrChangedDescription represents an error state where the description was changed for a property.
var ErrChangedDescription = errors.New("description changed")
