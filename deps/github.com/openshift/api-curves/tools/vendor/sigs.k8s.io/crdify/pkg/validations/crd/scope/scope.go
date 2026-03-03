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

package scope

import (
	"errors"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"sigs.k8s.io/crdify/pkg/config"
	"sigs.k8s.io/crdify/pkg/validations"
)

var (
	_ validations.Validation                                           = (*Scope)(nil)
	_ validations.Comparator[apiextensionsv1.CustomResourceDefinition] = (*Scope)(nil)
)

const name = "scope"

// Register registers the Scope validation
// with the provided validation registry.
func Register(registry validations.Registry) {
	registry.Register(name, factory)
}

// factory is a function used to initialize a Scope validation
// implementation based on the provided configuration.
func factory(_ map[string]interface{}) (validations.Validation, error) {
	return &Scope{}, nil
}

// Scope is a validations.Validation implementation
// used to check if the scope has changed from one
// CRD instance to another.
type Scope struct {
	// enforcement is the EnforcementPolicy that this validation
	// should use when performing its validation logic
	enforcement config.EnforcementPolicy
}

// Name returns the name of the Scope validation.
func (s *Scope) Name() string {
	return name
}

// SetEnforcement sets the EnforcementPolicy for the Scope validation.
func (s *Scope) SetEnforcement(enforcement config.EnforcementPolicy) {
	s.enforcement = enforcement
}

// Compare compares an old and a new CustomResourceDefintion, checking for any change to the scope from the
// old CustomResourceDefinition to the new CustomResourceDefinition.
func (s *Scope) Compare(a, b *apiextensionsv1.CustomResourceDefinition) validations.ComparisonResult {
	var err error
	if a.Spec.Scope != b.Spec.Scope {
		err = fmt.Errorf("%w : %q -> %q", ErrChangedScope, a.Spec.Scope, b.Spec.Scope)
	}

	return validations.HandleErrors(s.Name(), s.enforcement, err)
}

// ErrChangedScope represents an error state where the scope of the CustomResourceDefinition has changed.
var ErrChangedScope = errors.New("scope changed")
