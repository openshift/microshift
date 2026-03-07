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
	"k8s.io/apimachinery/pkg/util/sets"
	"sigs.k8s.io/crdify/pkg/config"
	"sigs.k8s.io/crdify/pkg/validations"
)

var (
	_ validations.Validation                                  = (*Required)(nil)
	_ validations.Comparator[apiextensionsv1.JSONSchemaProps] = (*Required)(nil)
)

const requiredValidationName = "required"

// RegisterRequired registers the Required validation
// with the provided validation registry.
func RegisterRequired(registry validations.Registry) {
	registry.Register(requiredValidationName, requiredFactory)
}

// requiredFactory is a function used to initialize a Required validation
// implementation based on the provided configuration.
func requiredFactory(_ map[string]interface{}) (validations.Validation, error) {
	return &Required{}, nil
}

// Required is a Validation that can be used to identify
// incompatible changes to the required constraints of CRD properties.
type Required struct {
	enforcement config.EnforcementPolicy
}

// Name returns the name of the Required validation.
func (r *Required) Name() string {
	return requiredValidationName
}

// SetEnforcement sets the EnforcementPolicy for the Required validation.
func (r *Required) SetEnforcement(policy config.EnforcementPolicy) {
	r.enforcement = policy
}

// Compare compares an old and a new JSONSchemaProps, checking for incompatible changes to the required constraints of a property.
// In order for callers to determine if diffs to a JSONSchemaProps have been handled by this validation
// the JSONSchemaProps.Required field will be reset to 'nil' as part of this method.
// It is highly recommended that only copies of the JSONSchemaProps to compare are provided to this method
// to prevent unintentional modifications.
func (r *Required) Compare(a, b *apiextensionsv1.JSONSchemaProps) validations.ComparisonResult {
	oldRequired := sets.New(a.Required...)
	newRequired := sets.New(b.Required...)
	diffRequired := newRequired.Difference(oldRequired)

	var err error

	if diffRequired.Len() > 0 {
		err = fmt.Errorf("%w: %v", ErrNewRequiredFields, diffRequired.UnsortedList())
	}

	a.Required = nil
	b.Required = nil

	return validations.HandleErrors(r.Name(), r.enforcement, err)
}

// ErrNewRequiredFields represents an error state where a property has new required fields.
var ErrNewRequiredFields = errors.New("new required fields")
