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

package validations

import (
	"errors"
	"fmt"
	"sync"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/util/validation/field"
	"sigs.k8s.io/crdify/pkg/config"
	"sigs.k8s.io/crdify/pkg/slices"
)

// GetCRDVersionByName returns a CustomResourceDefinitionVersion with the provided name from the provided CustomResourceDefinition.
func GetCRDVersionByName(crd *apiextensionsv1.CustomResourceDefinition, name string) *apiextensionsv1.CustomResourceDefinitionVersion {
	if crd == nil {
		return nil
	}

	for _, version := range crd.Spec.Versions {
		if version.Name == name {
			return &version
		}
	}

	return nil
}

// FlattenCRDVersion flattens the provided CustomResourceDefinition into a mapping of
// property path (i.e ^.spec.foo.bar) to its JSONSchemaProps.
func FlattenCRDVersion(crdVersion apiextensionsv1.CustomResourceDefinitionVersion) map[string]*apiextensionsv1.JSONSchemaProps {
	flatMap := map[string]*apiextensionsv1.JSONSchemaProps{}

	SchemaHas(crdVersion.Schema.OpenAPIV3Schema,
		field.NewPath("^"),
		field.NewPath("^"),
		nil,
		func(s *apiextensionsv1.JSONSchemaProps, _, simpleLocation *field.Path, _ []*apiextensionsv1.JSONSchemaProps) bool {
			flatMap[simpleLocation.String()] = s.DeepCopy()
			return false
		},
	)

	return flatMap
}

// Diff is a utility struct for holding an old and new JSONSchemaProps.
type Diff struct {
	Old *apiextensionsv1.JSONSchemaProps
	New *apiextensionsv1.JSONSchemaProps
}

// FlattenedCRDVersionDiff calculates differences between flattened CRD versions.
// Returns the set of differing properties as a map of the property path (i.e ^.spec.foo.bar)
// to the Diff (old and new JSONSchemaProps).
func FlattenedCRDVersionDiff(a, b map[string]*apiextensionsv1.JSONSchemaProps) map[string]Diff {
	diffMap := map[string]Diff{}

	for prop, oldSchema := range a {
		// Create a copy of the old schema and set the properties to nil.
		// In theory this will make it so we don't provide a diff for a parent property
		// based on changes to the children properties. The changes to the children
		// properties should still be evaluated since we are looping through a flattened
		// map of all the properties for the CRD version
		oldSchemaCopy := DropChildrenPropertiesFromJSONSchema(oldSchema)
		newSchema, ok := b[prop]

		// In the event the property no longer exists on the new version
		// create a diff entry with the new value being empty
		if !ok {
			diffMap[prop] = Diff{Old: oldSchemaCopy, New: &apiextensionsv1.JSONSchemaProps{}}
			// Continue as there is no newSchema to copy and evaluate for this prop.
			continue
		}

		// Do the same copy and unset logic for the new schema properties
		// before comparison to ensure we are only comparing the individual properties
		newSchemaCopy := DropChildrenPropertiesFromJSONSchema(newSchema)

		if !equality.Semantic.DeepEqual(oldSchemaCopy, newSchemaCopy) {
			diffMap[prop] = Diff{Old: oldSchemaCopy, New: newSchemaCopy}
		}
	}

	return diffMap
}

// DropChildrenPropertiesFromJSONSchema sets properties on a schema
// associated with children schemas to `nil`. Useful when calculating
// differences between a before and after of a given schema
// without the changes to its children schemas influencing the
// diff calculation.
// Returns a copy of the provided apiextensionsv1.JSONSchemaProps with children schemas dropped.
func DropChildrenPropertiesFromJSONSchema(schema *apiextensionsv1.JSONSchemaProps) *apiextensionsv1.JSONSchemaProps {
	schemaCopy := schema.DeepCopy()
	schemaCopy.Properties = nil
	schemaCopy.Items = nil

	return schemaCopy
}

// SchemaWalkerFunc is a function that walks a JSONSchemaProps.
// ancestry is an order list of ancestors of s, where index 0 is the root and index len-1 is the direct parent.
type SchemaWalkerFunc func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, ancestry []*apiextensionsv1.JSONSchemaProps) bool

// SchemaHas recursively traverses the Schema and calls the `pred`
// predicate to see if the schema contains specific values.
//
// The predicate MUST NOT keep a copy of the json schema NOR modify the
// schema.
//
//nolint:gocognit,cyclop
func SchemaHas(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, ancestry []*apiextensionsv1.JSONSchemaProps, pred SchemaWalkerFunc) bool {
	if s == nil {
		return false
	}

	if pred(s, fldPath, simpleLocation, ancestry) {
		return true
	}

	//nolint:gocritic
	nextAncestry := append(ancestry, s)

	if s.Items != nil {
		if s.Items != nil && schemaHasRecurse(s.Items.Schema, fldPath.Child("items"), simpleLocation.Key("*"), nextAncestry, pred) {
			return true
		}

		for i := range s.Items.JSONSchemas {
			if schemaHasRecurse(&s.Items.JSONSchemas[i], fldPath.Child("items", "jsonSchemas").Index(i), simpleLocation.Index(i), nextAncestry, pred) {
				return true
			}
		}
	}

	for i := range s.AllOf {
		if schemaHasRecurse(&s.AllOf[i], fldPath.Child("allOf").Index(i), simpleLocation, nextAncestry, pred) {
			return true
		}
	}

	for i := range s.AnyOf {
		if schemaHasRecurse(&s.AnyOf[i], fldPath.Child("anyOf").Index(i), simpleLocation, nextAncestry, pred) {
			return true
		}
	}

	for i := range s.OneOf {
		if schemaHasRecurse(&s.OneOf[i], fldPath.Child("oneOf").Index(i), simpleLocation, nextAncestry, pred) {
			return true
		}
	}

	if schemaHasRecurse(s.Not, fldPath.Child("not"), simpleLocation, nextAncestry, pred) {
		return true
	}

	for propertyName, s := range s.Properties {
		if schemaHasRecurse(&s, fldPath.Child("properties").Key(propertyName), simpleLocation.Child(propertyName), nextAncestry, pred) {
			return true
		}
	}

	if s.AdditionalProperties != nil {
		if schemaHasRecurse(s.AdditionalProperties.Schema, fldPath.Child("additionalProperties", "schema"), simpleLocation.Key("*"), nextAncestry, pred) {
			return true
		}
	}

	for patternName, s := range s.PatternProperties {
		if schemaHasRecurse(&s, fldPath.Child("allOf").Key(patternName), simpleLocation, nextAncestry, pred) {
			return true
		}
	}

	if s.AdditionalItems != nil {
		if schemaHasRecurse(s.AdditionalItems.Schema, fldPath.Child("additionalItems", "schema"), simpleLocation, nextAncestry, pred) {
			return true
		}
	}

	for _, s := range s.Definitions {
		if schemaHasRecurse(&s, fldPath.Child("definitions"), simpleLocation, nextAncestry, pred) {
			return true
		}
	}

	for dependencyName, d := range s.Dependencies {
		if schemaHasRecurse(d.Schema, fldPath.Child("dependencies").Key(dependencyName).Child("schema"), simpleLocation, nextAncestry, pred) {
			return true
		}
	}

	return false
}

//nolint:gochecknoglobals
var schemaPool = sync.Pool{
	New: func() any {
		return new(apiextensionsv1.JSONSchemaProps)
	},
}

func schemaHasRecurse(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, ancestry []*apiextensionsv1.JSONSchemaProps, pred SchemaWalkerFunc) bool {
	if s == nil {
		return false
	}

	schema, ok := schemaPool.Get().(*apiextensionsv1.JSONSchemaProps)
	if !ok {
		return false
	}
	defer schemaPool.Put(schema)

	*schema = *s

	return SchemaHas(schema, fldPath, simpleLocation, ancestry, pred)
}

// ComparatorsForValidations extracts the Comparators of type T from the provided set of Validations.
func ComparatorsForValidations[T Comparable](vals ...Validation) []Comparator[T] {
	comparators := []Comparator[T]{}

	for _, val := range vals {
		comp, ok := val.(Comparator[T])
		if !ok {
			continue
		}

		comparators = append(comparators, comp)
	}

	return comparators
}

// LoadValidationsFromRegistry initializes all validations registered with the provided
// registry using an empty configuration.
// It returns the validations in a map to make it easier to update the Validations in one-shot operations.
// Any errors encountered during the initialization process are aggregated and returned as a single error.
func LoadValidationsFromRegistry(registry Registry) (map[string]Validation, error) {
	vals := map[string]Validation{}
	errs := []error{}

	for _, validation := range registry.Registered() {
		val, err := registry.Validation(validation, make(map[string]interface{}))
		if err != nil {
			errs = append(errs, fmt.Errorf("initializing validation %q: %w", validation, err))
		}

		val.SetEnforcement(config.EnforcementPolicyError)

		vals[validation] = val
	}

	return vals, errors.Join(errs...)
}

// ConfigureValidations is a utility function for configuring the provided set of validations
// using the provided registyr and configuration.
// It returns a copy of the original validations mapping with validations that had specific
// configurations replaced with a newly initialized validation.
// Any errors encountered during the initialization process are aggregated and returned as a single error.
func ConfigureValidations(validations map[string]Validation, registry Registry, cfg config.Config) (map[string]Validation, error) {
	modified := validations
	errs := []error{}

	for _, validation := range cfg.Validations {
		val, err := registry.Validation(validation.Name, validation.Configuration)
		if err != nil {
			errs = append(errs, fmt.Errorf("configuring validation %q: %w", validation.Name, err))
			continue
		}

		switch validation.Enforcement {
		case config.EnforcementPolicyError, config.EnforcementPolicyWarn, config.EnforcementPolicyNone:
			val.SetEnforcement(validation.Enforcement)
		default:
			errs = append(errs, fmt.Errorf("configuring validation %q: %w : %q", validation.Name, errUnknownEnforcementPolicy, validation.Enforcement))
		}

		modified[validation.Name] = val
	}

	return modified, errors.Join(errs...)
}

var errUnknownEnforcementPolicy = errors.New("unknown enforcement policy")

// HandleErrors is a utility function for Comparators to generate a ComparisonResult
// based on the provided Comparator name, enforcement policy, and any errors it encountered.
func HandleErrors(name string, policy config.EnforcementPolicy, errs ...error) ComparisonResult {
	result := ComparisonResult{
		Name: name,
	}

	switch policy {
	case config.EnforcementPolicyError:
		if errors.Join(errs...) != nil {
			result.Errors = slices.Translate(func(err error) string { return err.Error() }, errs...)
		}
	case config.EnforcementPolicyWarn:
		if errors.Join(errs...) != nil {
			result.Warnings = slices.Translate(func(err error) string { return err.Error() }, errs...)
		}
	case config.EnforcementPolicyNone:
		return result
	}

	return result
}
