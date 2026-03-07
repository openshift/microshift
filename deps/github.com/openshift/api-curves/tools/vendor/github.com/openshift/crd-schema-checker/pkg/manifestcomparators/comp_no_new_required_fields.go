package manifestcomparators

import (
	"fmt"
	"regexp"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type noNewRequiredFields struct{}

func NoNewRequiredFields() CRDComparator {
	return noNewRequiredFields{}
}

func (noNewRequiredFields) Name() string {
	return "NoNewRequiredFields"
}

func (noNewRequiredFields) WhyItMatters() string {
	return "If new fields are required, then old clients will not function properly.  Even if CRD defaulting is used, " +
		"CRD defaulting requires allowing an object with an empty or missing value to then get defaulted."
}

// isFieldOptional checks if a field is optional (ie not required by its parent)
func isFieldOptional(
	s *apiextensionsv1.JSONSchemaProps,
	ancestors []*apiextensionsv1.JSONSchemaProps,
	fldPath *field.Path,
	newToFldPath map[*apiextensionsv1.JSONSchemaProps]*field.Path,
	newFldPathToRequiredFields map[string]sets.Set[string]) bool {

	// Check if field is an optional array
	if s.Type == "array" && (s.MinLength == nil || *s.MinLength == 0) {
		return true
	}

	// Check if field is not required by its parent
	if len(ancestors) > 0 {
		parentOfField := ancestors[len(ancestors)-1]
		groups := lastIndexOrKeyRegexp.FindStringSubmatch(fldPath.String())
		if len(groups) == 2 {
			lastStep := groups[1]
			parentRequiredFields := newFldPathToRequiredFields[newToFldPath[parentOfField].String()]
			return !parentRequiredFields.Has(lastStep)
		}
	}

	return false
}

func (b noNewRequiredFields) Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	if existingCRD == nil {
		return ComparisonResults{
			Name:         b.Name(),
			WhyItMatters: b.WhyItMatters(),

			Errors:   nil,
			Warnings: nil,
			Infos:    nil,
		}, nil
	}
	errsToReport := []string{}

	for _, newVersion := range newCRD.Spec.Versions {

		existingVersion := GetVersionByName(existingCRD, newVersion.Name)
		if existingVersion == nil {
			continue
		}

		existingRequiredFields := map[string]sets.String{}
		existingFldPathToJSONSchemaProps := map[string]*apiextensionsv1.JSONSchemaProps{}
		fldPathToSimpleLocation := map[string]string{}
		SchemaHas(existingVersion.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
			func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, _ []*apiextensionsv1.JSONSchemaProps) bool {
				existingRequiredFields[fldPath.String()] = sets.NewString(s.Required...)
				existingFldPathToJSONSchemaProps[fldPath.String()] = s
				fldPathToSimpleLocation[fldPath.String()] = simpleLocation.String()
				return false
			})

		// New fields can be required if they are wrapped inside new structs that are themselves optional.
		// For instance, you cannot add .spec.thingy as required, but if you add .spec.top as optional and at the same
		// time add .spec.top.thingy as required, this is allowed.
		// Similar logic exists for adding an array with minlength > 0
		newFldPathToRequiredFields := map[string]sets.Set[string]{}

		// First collect all possible paths.
		// There can be multiple invocations per fldPath (e.g. from s.OneOf and s.Properties).
		SchemaHas(newVersion.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
			func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, ancestors []*apiextensionsv1.JSONSchemaProps) bool {
				newFldPathToRequiredFields[fldPath.String()] = sets.New(s.Required...)
				fldPathToSimpleLocation[fldPath.String()] = simpleLocation.String()
				return false
			})

		newRequiredFields := sets.NewString()
		newToFldPath := map[*apiextensionsv1.JSONSchemaProps]*field.Path{}
		SchemaHas(newVersion.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
			func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, ancestors []*apiextensionsv1.JSONSchemaProps) bool {
				newToFldPath[s] = fldPath
				if s.Type == "array" {
					// if it's an array, we have a different property to check.  A new array cannot be required unless it's ancestor is new.
					if s.MinLength == nil || *s.MinLength == 0 {
						// if there is no required length, this is fine
						return false
					}
					// this means we're an array with a minLength, check to see if any parent wrapper is both new and optional.
					if isAnyAncestorNewAndNullable(ancestors, existingFldPathToJSONSchemaProps, newToFldPath, newFldPathToRequiredFields) {
						return false
					}

					// if we search all ancestors and couldn't find a new, optional element, then the current array cannot
					// have a minLength greater than zero.
					newRequiredFields.Insert(fldPathToSimpleLocation[fldPath.String()])
					return false
				}

				if len(s.Required) == 0 {
					// if nothing is required, nothing to check.
					return false
				}

				existingRequired, _ := existingRequiredFields[fldPath.String()]

				// Check if this field actually existed in the old schema
				_, fieldExistedBefore := existingFldPathToJSONSchemaProps[fldPath.String()]

				if !fieldExistedBefore && isFieldOptional(s, ancestors, fldPath, newToFldPath, newFldPathToRequiredFields) {
					// if the parent of the required field didn't exist before AND is optional,
					// then we can allow a child to be required.
					return false
				}

				if isAnyAncestorNewAndNullable(ancestors, existingFldPathToJSONSchemaProps, newToFldPath, newFldPathToRequiredFields) {
					// if any ancestor of the parent of the required field is new and nullable, then required is allowed.
					return false
				}

				// this covers newly required fields.
				newRequired := sets.NewString(s.Required...)
				if disallowedRequired := newRequired.Difference(existingRequired); len(disallowedRequired) > 0 {
					for _, curr := range disallowedRequired.List() {
						newRequiredFields.Insert(fmt.Sprintf("%s.%s", fldPathToSimpleLocation[fldPath.String()], curr))
					}
					return false
				}

				return false
			})

		for _, newRequiredField := range newRequiredFields.List() {
			errsToReport = append(errsToReport, fmt.Sprintf("crd/%v version/%v field/%v is new and may not be required", newCRD.Name, newVersion.Name, newRequiredField))
		}

	}

	return ComparisonResults{
		Name:         b.Name(),
		WhyItMatters: b.WhyItMatters(),

		Errors:   errsToReport,
		Warnings: nil,
		Infos:    nil,
	}, nil
}

// captures parent from ^.properties[spec].properties[parent]
var lastIndexOrKeyRegexp = regexp.MustCompile(`.*\[([^\]]+)\]$`)

func isAnyAncestorNewAndNullable(
	ancestors []*apiextensionsv1.JSONSchemaProps,
	existingFldPathToJSONSchemaProps map[string]*apiextensionsv1.JSONSchemaProps,
	newToFldPath map[*apiextensionsv1.JSONSchemaProps]*field.Path,
	newFldPathToRequiredFields map[string]sets.Set[string]) bool {

	for i := len(ancestors) - 1; i >= 0; i-- {
		ancestor := ancestors[i]
		ancestorFldPath := newToFldPath[ancestor]

		// check if the ancestor is optional
		if !isFieldOptional(ancestor, ancestors[:i], ancestorFldPath, newToFldPath, newFldPathToRequiredFields) {
			// if this ancestor isn't optional, then it cannot allow the current element to be required
			continue
		}

		if _, existed := existingFldPathToJSONSchemaProps[ancestorFldPath.String()]; existed {
			// if this ancestor previously existed, then it cannot allow the current element to be required
			continue
		}
		if i == 0 {
			// if the current accessor is the top level and Nullable, then it isn't required
			return true
		}

		// does the current ancestor require
		parentOfAncestor := ancestors[i-1]
		groups := lastIndexOrKeyRegexp.FindStringSubmatch(ancestorFldPath.String())
		if len(groups) != 2 {
			// should not happen: not a valid last step (has no index)
			continue
		}
		lastStep := groups[1]
		prevAncestorRequiredFields := newFldPathToRequiredFields[newToFldPath[parentOfAncestor].String()]
		if !prevAncestorRequiredFields.Has(lastStep) {
			// the current ancestor is not required, then we're ok and don't need to search further
			return true
		}
	}

	return false
}
