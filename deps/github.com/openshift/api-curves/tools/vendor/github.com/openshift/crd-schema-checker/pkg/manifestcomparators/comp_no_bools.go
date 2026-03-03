package manifestcomparators

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type noBools struct{}

func NoBools() CRDComparator {
	return noBools{}
}

func (noBools) Name() string {
	return "NoBools"
}

func (noBools) WhyItMatters() string {
	return "Booleans rarely stay booleans and can never develop new options.  This frequently leads to cases where there " +
		"are multiple boolean fields, with some combinations of values not being allowed.  Additionally, strings provide " +
		"expressive names and values, describing degrees or conditions of a thing.  Also, booleans cannot be defaulted, " +
		"pointers to booleans can be, but at that point you've already got a tri-state, so it's not a boolean is it..."
}

func (b noBools) Validate(crd *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	errsToReport := []string{}

	for _, newVersion := range crd.Spec.Versions {
		newBoolFields := []string{}
		SchemaHas(newVersion.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
			func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, _ []*apiextensionsv1.JSONSchemaProps) bool {
				if s.Type == "boolean" {
					newBoolFields = append(newBoolFields, simpleLocation.String())
				}
				return false
			})

		for _, newBoolField := range newBoolFields {
			errsToReport = append(errsToReport, fmt.Sprintf("crd/%v version/%v field/%v may not be a boolean", crd.Name, newVersion.Name, newBoolField))
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

func (b noBools) Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	return RatchetCompare(b, existingCRD, newCRD)
}
