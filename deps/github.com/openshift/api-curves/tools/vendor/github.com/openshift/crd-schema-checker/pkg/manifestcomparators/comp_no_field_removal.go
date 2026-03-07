package manifestcomparators

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type noFieldRemoval struct{}

func NoFieldRemoval() CRDComparator {
	return noFieldRemoval{}
}

func (noFieldRemoval) Name() string {
	return "NoFieldRemoval"
}

func (noFieldRemoval) WhyItMatters() string {
	return "If fields are removed, then clients that rely on those fields will not be able to read them or write them."
}

func getFields(version *apiextensionsv1.CustomResourceDefinitionVersion) sets.String {
	fields := sets.NewString()
	SchemaHas(version.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
		func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, _ []*apiextensionsv1.JSONSchemaProps) bool {
			fields.Insert(simpleLocation.String())
			return false
		})

	return fields
}

func (b noFieldRemoval) Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
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

		existingFields := getFields(existingVersion)
		newFields := getFields(&newVersion)

		removedFields := existingFields.Difference(newFields)
		for _, removedField := range removedFields.List() {
			errsToReport = append(errsToReport, fmt.Sprintf("crd/%v version/%v field/%v may not be removed", newCRD.Name, newVersion.Name, removedField))
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
