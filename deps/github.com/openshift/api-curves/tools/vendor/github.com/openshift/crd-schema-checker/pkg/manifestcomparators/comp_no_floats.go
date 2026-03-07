package manifestcomparators

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type noFloats struct{}

func NoFloats() CRDComparator {
	return noFloats{}
}

func (noFloats) Name() string {
	return "NoFloats"
}

func (noFloats) WhyItMatters() string {
	return "Floating-point values cannot be reliably round-tripped (encoded and re-decoded) without changing, " +
		"and have varying precision and representations across languages and architectures."
}

func (b noFloats) Validate(crd *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	errsToReport := []string{}

	for _, newVersion := range crd.Spec.Versions {
		newFloatFields := []string{}
		SchemaHas(newVersion.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
			func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, _ []*apiextensionsv1.JSONSchemaProps) bool {
				if s.Type == "number" {
					newFloatFields = append(newFloatFields, simpleLocation.String())
				}
				return false
			})

		for _, newFloatField := range newFloatFields {
			errsToReport = append(errsToReport, fmt.Sprintf("crd/%v version/%v field/%v may not be a float", crd.Name, newVersion.Name, newFloatField))
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

func (b noFloats) Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	return RatchetCompare(b, existingCRD, newCRD)
}
