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

package existingfieldremoval

import (
	"errors"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/crdify/pkg/config"
	"sigs.k8s.io/crdify/pkg/validations"
)

var (
	_ validations.Validation                                           = (*ExistingFieldRemoval)(nil)
	_ validations.Comparator[apiextensionsv1.CustomResourceDefinition] = (*ExistingFieldRemoval)(nil)
)

const name = "existingFieldRemoval"

// Register registers the ExistingFieldRemoval validation
// with the provided validation registry.
func Register(registry validations.Registry) {
	registry.Register(name, factory)
}

// factory is a function used to initialize an ExistingFieldRemoval validation
// implementation based on the provided configuration.
func factory(_ map[string]interface{}) (validations.Validation, error) {
	return &ExistingFieldRemoval{}, nil
}

// ExistingFieldRemoval is a validations.Validation implementation
// used to check if any existing fields have been removed from one
// CRD instance to another.
type ExistingFieldRemoval struct {
	// enforcement is the EnforcementPolicy that this validation
	// should use when performing its validation logic
	enforcement config.EnforcementPolicy
}

// Name returns the name of the ExistingFieldRemoval validation.
func (efr *ExistingFieldRemoval) Name() string {
	return name
}

// SetEnforcement sets the EnforcementPolicy for the ExistingFieldRemoval validation.
func (efr *ExistingFieldRemoval) SetEnforcement(policy config.EnforcementPolicy) {
	efr.enforcement = policy
}

// Compare compares an old and a new CustomResourceDefintion, checking for any fields that were removed
// from the old CustomResourceDefinition in the new CustomResourceDefinition.
func (efr *ExistingFieldRemoval) Compare(a, b *apiextensionsv1.CustomResourceDefinition) validations.ComparisonResult {
	errs := []error{}

	for _, newVersion := range b.Spec.Versions {
		existingVersion := validations.GetCRDVersionByName(a, newVersion.Name)
		if existingVersion == nil {
			continue
		}

		existingFields := getFields(existingVersion)
		newFields := getFields(&newVersion)

		removedFields := existingFields.Difference(newFields)
		for _, removedField := range removedFields.UnsortedList() {
			errs = append(errs, fmt.Errorf("%w : %v.%v", ErrRemovedExistingField, newVersion.Name, removedField))
		}
	}

	return validations.HandleErrors(efr.Name(), efr.enforcement, errs...)
}

// ErrRemovedExistingField represents an error state where existing fields have been removed
// from the CustomResourceDefinition.
var ErrRemovedExistingField = errors.New("removed field")

// getFields returns a set of all the fields for the provided CustomResourceDefinitionVersion.
func getFields(v *apiextensionsv1.CustomResourceDefinitionVersion) sets.Set[string] {
	fields := sets.New[string]()

	validations.SchemaHas(v.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
		func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, _ []*apiextensionsv1.JSONSchemaProps) bool {
			fields.Insert(simpleLocation.String())
			return false
		},
	)

	return fields
}
