package manifestcomparators

import (
	"fmt"
	"strings"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/validation/field"
)

type conditionsMustHaveProperSSATags struct{}

func ConditionsMustHaveProperSSATags() CRDComparator {
	return conditionsMustHaveProperSSATags{}
}

func (conditionsMustHaveProperSSATags) Name() string {
	return "ConditionsMustHaveProperSSATags"
}

func (conditionsMustHaveProperSSATags) WhyItMatters() string {
	return "Conditions should follow the standard schema included in  " +
		"https://github.com/kubernetes/apimachinery/blob/release-1.29/pkg/apis/meta/v1/types.go#L1482-L1542" +
		"and collection of conditions should be treated as a map with a key of type. " +
		"This is indicated in kubebuilder tags with '// +listType=map' and '// +listMapKey=type'."

}

func (c conditionsMustHaveProperSSATags) Validate(crd *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	errsToReport := []string{}

	for _, newVersion := range crd.Spec.Versions {
		conditionsWithoutMapListType := []string{}
		conditionsWithoutListMapKeysType := []string{}
		invalidConditionsProperties := []string{}
		SchemaHas(newVersion.Schema.OpenAPIV3Schema, field.NewPath("^"), field.NewPath("^"), nil,
			func(s *apiextensionsv1.JSONSchemaProps, fldPath, simpleLocation *field.Path, _ []*apiextensionsv1.JSONSchemaProps) bool {
				if s.Type != "array" {
					return false
				}

				if !strings.Contains(simpleLocation.String(), ".conditions") {
					return false
				}

				errs := validateConditionProperties(s.Items.Schema.Properties)
				if len(errs) != 0 {
					invalidConditionsProperties = append(invalidConditionsProperties, errs...)
				}

				if s.XListType == nil || *s.XListType != "map" {
					conditionsWithoutMapListType = append(conditionsWithoutMapListType, simpleLocation.String())
				}

				if len(s.XListMapKeys) == 0 || !listMapKeysHasSingleTypeElement(s.XListMapKeys) {
					conditionsWithoutListMapKeysType = append(conditionsWithoutListMapKeysType, simpleLocation.String())
				}

				return false
			})

		for _, invalidConditionProp := range invalidConditionsProperties {
			errStr := fmt.Sprintf("crd/%v version/%v field/^.status.condition must define valid condition properties: %s", crd.Name, newVersion.Name, invalidConditionProp)
			errsToReport = append(errsToReport, errStr)
		}

		for _, affectedField := range conditionsWithoutMapListType {
			errStr := fmt.Sprintf("crd/%v version/%v field/%v must set x-kubernetes-list-type with value \"map\"", crd.Name, newVersion.Name, affectedField)
			errsToReport = append(errsToReport, errStr)
		}
		for _, affectedField := range conditionsWithoutListMapKeysType {
			errStr := fmt.Sprintf("crd/%v version/%v field/%v must set x-kubernetes-list-map-keys with single \"type\" value", crd.Name, newVersion.Name, affectedField)
			errsToReport = append(errsToReport, errStr)
		}
	}

	return ComparisonResults{
		Name:         c.Name(),
		WhyItMatters: c.WhyItMatters(),

		Errors:   errsToReport,
		Warnings: nil,
		Infos:    nil,
	}, nil
}

func (b conditionsMustHaveProperSSATags) Compare(existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	return RatchetCompare(b, existingCRD, newCRD)
}

func listMapKeysHasSingleTypeElement(keys []string) bool {
	if len(keys) != 1 {
		return false
	}
	return keys[0] == "type"
}

func jsonSliceContainsString(a []apiextensionsv1.JSON, s string) bool {
	for _, v := range a {
		if string(v.Raw) == s {
			return true
		}
	}
	return false
}

func validateConditionProperties(properties map[string]apiextensionsv1.JSONSchemaProps) []string {
	var errString []string
	typeP, ok := properties["type"]
	if !ok {
		errString = append(errString, "type attribute is missing")
	}
	reason, ok := properties["reason"]
	if !ok {
		errString = append(errString, "reason attribute is missing")
	}
	status, ok := properties["status"]
	if !ok {
		errString = append(errString, "status attribute is missing")
	}
	lastTransitionTime, ok := properties["lastTransitionTime"]
	if !ok {
		errString = append(errString, "lastTransitionTime attribute is missing")
	}
	observedGeneration, ok := properties["observedGeneration"]
	if !ok {
		errString = append(errString, "observedGeneration attribute is missing")
	}
	message, ok := properties["message"]
	if !ok {
		errString = append(errString, "message attribute is missing")
	}

	if len(errString) > 0 {
		return errString
	}

	errString = append(errString, validateConditionType(typeP)...)
	errString = append(errString, validateConditionReason(reason)...)
	errString = append(errString, validateConditionStatus(status)...)
	errString = append(errString, validateConditionObservedGeneration(observedGeneration)...)
	errString = append(errString, validateConditionLastTransitionTime(lastTransitionTime)...)
	errString = append(errString, validateConditionMessage(message)...)
	return errString
}

func validateConditionReason(reasonProp apiextensionsv1.JSONSchemaProps) []string {
	expectedPattern := "^[A-Za-z]([A-Za-z0-9_,:]*[A-Za-z0-9_])?$"
	expectedMaxLength := 1024
	var errs []string
	if reasonProp.Pattern != expectedPattern {
		errs = append(errs, fmt.Sprintf("reason attribute does not set correct pattern %s", expectedPattern))
	}

	if *reasonProp.MaxLength != int64(expectedMaxLength) {
		errs = append(errs, fmt.Sprintf("reason attribute does not set maxLength to %d", expectedMaxLength))
	}

	return errs
}

func validateConditionType(typeProp apiextensionsv1.JSONSchemaProps) []string {
	expectedPattern := "^([a-z0-9]([-a-z0-9]*[a-z0-9])?(\\.[a-z0-9]([-a-z0-9]*[a-z0-9])?)*/)?(([A-Za-z0-9][-A-Za-z0-9_.]*)?[A-Za-z0-9])$"
	expectedMaxLength := 316
	var errs []string
	if typeProp.Pattern != expectedPattern {
		errs = append(errs, fmt.Sprintf("type attribute must set correct pattern %s", expectedPattern))
	}

	if *typeProp.MaxLength != int64(expectedMaxLength) {
		errs = append(errs, fmt.Sprintf("type attribute must set maxLength to %d", expectedMaxLength))
	}

	return errs
}

func validateConditionStatus(statusProp apiextensionsv1.JSONSchemaProps) []string {
	var errs []string
	expectedEnumLength := 3
	expectedEnumTrue := "\"True\""
	expectedEnumFalse := "\"False\""
	expectedEnumUnknown := "\"Unknown\""
	if len(statusProp.Enum) != expectedEnumLength {
		errs = append(errs, fmt.Sprintf("status attribute is expected as Enum type with %d values", expectedEnumLength))
	}

	if !jsonSliceContainsString(statusProp.Enum, expectedEnumTrue) {
		errs = append(errs, fmt.Sprintf("status attribute is missing expected %s Enum value", expectedEnumTrue))
	}

	if !jsonSliceContainsString(statusProp.Enum, expectedEnumFalse) {
		errs = append(errs, fmt.Sprintf("status attribute is missing expected %s Enum value", expectedEnumFalse))
	}

	if !jsonSliceContainsString(statusProp.Enum, expectedEnumUnknown) {
		errs = append(errs, fmt.Sprintf("status attribute is missing expected %s Enum value", expectedEnumUnknown))
	}

	return errs
}

func validateConditionLastTransitionTime(lastTransitionTimeProp apiextensionsv1.JSONSchemaProps) []string {
	var errs []string
	expectedType := "string"
	expectedFormat := "date-time"
	if lastTransitionTimeProp.Type != expectedType {
		errs = append(errs, fmt.Sprintf("lastTransitionTime attribute must be of Type %s", expectedType))
	}
	if lastTransitionTimeProp.Format != expectedFormat {
		errs = append(errs, fmt.Sprintf("lastTransitionTime attribute must set %s Format", expectedFormat))
	}
	return errs
}

func validateConditionObservedGeneration(observedGenerationProp apiextensionsv1.JSONSchemaProps) []string {
	expectedMinimum := 0
	var errs []string
	if *observedGenerationProp.Minimum != float64(expectedMinimum) {
		errs = append(errs, fmt.Sprintf("observedGeneration attribute must set Minimum value to %d", expectedMinimum))
	}

	return errs
}

func validateConditionMessage(messageProp apiextensionsv1.JSONSchemaProps) []string {
	var errs []string
	expectedMaxLength := 32768
	if *messageProp.MaxLength != int64(expectedMaxLength) {
		errs = append(errs, fmt.Sprintf("message attribute must set maxLength to %d", expectedMaxLength))
	}
	return errs
}
