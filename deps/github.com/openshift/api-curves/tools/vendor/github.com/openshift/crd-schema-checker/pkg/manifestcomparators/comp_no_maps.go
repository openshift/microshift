package manifestcomparators

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type noMaps struct{}

func NoMaps() CRDComparator {
	return noMaps{}
}

func (noMaps) Name() string {
	return "NoMaps"
}

func (noMaps) WhyItMatters() string {
	return "When serialized into yaml or json, maps don't have \"names\" associated with their key.  This makes " +
		"it less obvious what the key of map means or what is for.  Additionally, maps are not guaranteed stable " +
		"for serialization, but lists are always ordered.  Instead of maps, use lists with a field that functions as " +
		"a key and use a listMapKey marker for server-side-apply."
}

func (b noMaps) Validate(crd *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	errsToReport := []string{}

	for _, newVersion := range crd.Spec.Versions {
		newMapFields := []string{}
		SchemaHas(newVersion.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
			func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, _ []*apiextensionsv1.JSONSchemaProps) bool {
				if s.Type == "object" {
					// I think this is how openapi v3 marks maps: https://swagger.io/docs/specification/data-models/dictionaries/
					// "normal" objects appear to use properties, not additionalProperties.
					if s.AdditionalProperties != nil {
						newMapFields = append(newMapFields, simpleLocation.String())
					}
				}
				return false
			})

		for _, newMapField := range newMapFields {
			errsToReport = append(errsToReport, fmt.Sprintf("crd/%v version/%v field/%v may not be a map", crd.Name, newVersion.Name, newMapField))
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

func (b noMaps) Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	return RatchetCompare(b, existingCRD, newCRD)
}
