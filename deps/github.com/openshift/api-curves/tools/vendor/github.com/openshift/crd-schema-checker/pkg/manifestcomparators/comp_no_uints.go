package manifestcomparators

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type noUints struct{}

func NoUints() CRDComparator {
	return noUints{}
}

func (noUints) Name() string {
	return "NoUints"
}

func (noUints) WhyItMatters() string {
	return "Unsigned integers don't have consistent support across languages and libraries."
}

func (n noUints) Validate(crd *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	errsToReport := []string{}

	for _, newVersion := range crd.Spec.Versions {
		uintFields := []string{}
		SchemaHas(newVersion.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
			func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, _ []*apiextensionsv1.JSONSchemaProps) bool {
				if s.Format == "uint" {
					uintFields = append(uintFields, simpleLocation.String())
				}
				return false
			})

		for _, newUintField := range uintFields {
			errsToReport = append(errsToReport, fmt.Sprintf("crd/%v version/%v field/%v may not be a uint", crd.Name, newVersion.Name, newUintField))
		}

	}

	return ComparisonResults{
		Name:         n.Name(),
		WhyItMatters: n.WhyItMatters(),

		Errors:   errsToReport,
		Warnings: nil,
		Infos:    nil,
	}, nil
}

func (n noUints) Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	return RatchetCompare(n, existingCRD, newCRD)
}
