package manifestcomparators

import (
	"sync"

	"k8s.io/apimachinery/pkg/util/validation/field"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

// GetVersionByName can be nil if the version doesn't exist
func GetVersionByName(crd *apiextensionsv1.CustomResourceDefinition, versionName string) *apiextensionsv1.CustomResourceDefinitionVersion {
	if crd == nil {
		return nil
	}

	for i := range crd.Spec.Versions {
		if crd.Spec.Versions[i].Name == versionName {
			return &crd.Spec.Versions[i]
		}
	}

	return nil
}

// ancestry is an order list of ancestors of s, where index 0 is the root and index len-1 is the direct parent
type SchemaWalkerFunc func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, ancestry []*apiextensionsv1.JSONSchemaProps) bool

// SchemaHas recursively traverses the Schema and calls the `pred`
// predicate to see if the schema contains specific values.
//
// The predicate MUST NOT keep a copy of the json schema NOR modify the
// schema.
func SchemaHas(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, ancestry []*apiextensionsv1.JSONSchemaProps, pred SchemaWalkerFunc) bool {
	if s == nil {
		return false
	}

	if pred(s, fldPath, simpleLocation, ancestry) {
		return true
	}

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

var schemaPool = sync.Pool{
	New: func() any {
		return new(apiextensionsv1.JSONSchemaProps)
	},
}

func schemaHasRecurse(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, ancestry []*apiextensionsv1.JSONSchemaProps, pred SchemaWalkerFunc) bool {
	if s == nil {
		return false
	}
	schema := schemaPool.Get().(*apiextensionsv1.JSONSchemaProps)
	defer schemaPool.Put(schema)
	*schema = *s
	return SchemaHas(schema, fldPath, simpleLocation, ancestry, pred)
}
