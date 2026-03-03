package manifestcomparators

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type noEnumRemoval struct{}

func NoEnumRemoval() CRDComparator {
	return noEnumRemoval{}
}

func (noEnumRemoval) Name() string {
	return "NoEnumRemoval"
}

func (noEnumRemoval) WhyItMatters() string {
	return "If enums are removed, then clients that use those enum values will not be able to upgrade to the newest CRD."
}

func getEnums(version *apiextensionsv1.CustomResourceDefinitionVersion) map[string]sets.String {
	enumsMap := make(map[string]sets.String)
	SchemaHas(version.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
		func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, _ []*apiextensionsv1.JSONSchemaProps) bool {
			for _, enum := range s.Enum {
				_, exists := enumsMap[simpleLocation.String()]
				if !exists {
					enumsMap[simpleLocation.String()] = sets.NewString()
				}
				enumsMap[simpleLocation.String()].Insert(string(enum.Raw))
			}
			return false
		})

	return enumsMap
}

func (b noEnumRemoval) Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
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

		existingEnumsMap := getEnums(existingVersion)
		newEnumsMap := getEnums(&newVersion)

		for field, existingEnums := range existingEnumsMap {
			newEnums, exists := newEnumsMap[field]
			if exists {
				removedEnums := existingEnums.Difference(newEnums)
				for _, removedEnum := range removedEnums.List() {
					errsToReport = append(errsToReport, fmt.Sprintf("crd/%v version/%v enum/%v may not be removed for field/%v", newCRD.Name, newVersion.Name, removedEnum, field))
				}
			}
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
