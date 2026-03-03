package manifestcomparators

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type mustHaveStatus struct{}

func MustHaveStatus() CRDComparator {
	return mustHaveStatus{}
}

func (mustHaveStatus) Name() string {
	return "MustHaveStatus"
}

func (mustHaveStatus) WhyItMatters() string {
	return "When the schema has a status field, it should be controlled via a status suberesource for different permissions " +
		"to control those who can control desired state from those who can control the actual state."
}

func (b mustHaveStatus) Validate(crd *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	const statusField = "^.status"
	errsToReport := []string{}

	for _, newVersion := range crd.Spec.Versions {
		if newVersion.Subresources != nil && newVersion.Subresources.Status != nil {
			continue
		}

		hasStatus := false
		SchemaHas(newVersion.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
			func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, _ []*apiextensionsv1.JSONSchemaProps) bool {
				if simpleLocation.String() == statusField {
					hasStatus = true
					return true
				}
				return false
			})

		if hasStatus {
			errsToReport = append(errsToReport, fmt.Sprintf("crd/%v version/%v field/%v must have a status subresource in .spec.version[name=%v].subresources.status to match its schema.", crd.Name, newVersion.Name, statusField, newVersion.Name))
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

func (b mustHaveStatus) Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	return RatchetCompare(b, existingCRD, newCRD)
}
