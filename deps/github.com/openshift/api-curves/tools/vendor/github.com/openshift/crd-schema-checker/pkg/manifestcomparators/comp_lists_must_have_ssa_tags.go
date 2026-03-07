package manifestcomparators

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type listsMustHaveSSATags struct{}

func ListsMustHaveSSATags() CRDComparator {
	return listsMustHaveSSATags{}
}

func (listsMustHaveSSATags) Name() string {
	return "ListsMustHaveSSATags"
}

func (listsMustHaveSSATags) WhyItMatters() string {
	return "Lists require x-kubernetes-list-type tags in order to properly merge different requests from different field managers.  " +
		"Valid value are 'atomic', 'set', and 'map' and are indicated in kubebuilder tags with '// +listType=<val>' and " +
		"'// +listMapKey=<val>'."
}

func (b listsMustHaveSSATags) Validate(crd *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	errsToReport := []string{}

	for _, newVersion := range crd.Spec.Versions {
		fieldsWithoutListType := []string{}
		SchemaHas(newVersion.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
			func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, _ []*apiextensionsv1.JSONSchemaProps) bool {
				if s.Type != "array" {
					return false
				}
				if s.XListType == nil || len(*s.XListType) == 0 {
					fieldsWithoutListType = append(fieldsWithoutListType, simpleLocation.String())
				}
				return false
			})

		for _, newMapField := range fieldsWithoutListType {
			errsToReport = append(errsToReport, fmt.Sprintf("crd/%v version/%v field/%v must set x-kubernetes-list-type", crd.Name, newVersion.Name, newMapField))
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

func (b listsMustHaveSSATags) Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	return RatchetCompare(b, existingCRD, newCRD)
}
