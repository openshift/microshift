package manifestcomparators

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type noDataTypeChange struct{}

func NoDataTypeChange() CRDComparator {
	return noDataTypeChange{}
}

func (noDataTypeChange) Name() string {
	return "NoDataTypeChange"
}

func (noDataTypeChange) WhyItMatters() string {
	return "If the data type of fields are changed, then clients that rely on those fields will not be able to read them or write them."
}

func getFieldsAndTypes(version *apiextensionsv1.CustomResourceDefinitionVersion) map[string]string {
	fieldsAndTypes := make(map[string]string)
	SchemaHas(version.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
		func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, _ []*apiextensionsv1.JSONSchemaProps) bool {
			fieldsAndTypes[simpleLocation.String()] = s.Type
			return false
		})

	return fieldsAndTypes
}

type TypeChange struct {
	ExistingType string
	NewType      string
}

func getChangedTypes(existingFieldsAndTypes map[string]string, newFieldsAndTypes map[string]string) map[string]TypeChange {
	changedTypes := make(map[string]TypeChange)

	for existingField, existingType := range existingFieldsAndTypes {
		if newType, ok := newFieldsAndTypes[existingField]; ok {
			if existingType != newType {
				changedTypes[existingField] = TypeChange{ExistingType: existingType, NewType: newType}
			}
		}
	}

	return changedTypes
}

func (b noDataTypeChange) Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
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

		existingFieldsAndTypes := getFieldsAndTypes(existingVersion)
		newFieldsAndTypes := getFieldsAndTypes(&newVersion)

		changedTypes := getChangedTypes(existingFieldsAndTypes, newFieldsAndTypes)
		for changedField, changedType := range changedTypes {
			errsToReport = append(errsToReport, fmt.Sprintf("crd/%v version/%v data type of field/%v may not be changed from %v to %v", newCRD.Name, newVersion.Name, changedField, changedType.ExistingType, changedType.NewType))
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
