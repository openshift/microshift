package manifestcomparators

import (
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func RatchetCompare(validator SingleCRDValidator, existingCRD, newCRD *apiextensionsv1.CustomResourceDefinition) (ComparisonResults, error) {
	var oldResults ComparisonResults
	if existingCRD != nil {
		var err error
		oldResults, err = validator.Validate(existingCRD)
		if err != nil {
			return ComparisonResults{}, err
		}
	}

	newResults, err := validator.Validate(newCRD)
	if err != nil {
		return ComparisonResults{}, err
	}

	ret := ComparisonResults{
		Name:         newResults.Name,
		WhyItMatters: newResults.WhyItMatters,
		Errors:       stringDiff(newResults.Errors, oldResults.Errors),
		Warnings:     stringDiff(newResults.Warnings, oldResults.Warnings),
		Infos:        stringDiff(newResults.Infos, oldResults.Infos),
	}

	return ret, nil
}

func stringDiff(s1 []string, s2 []string) []string {
	ret := []string{}
	for _, curr := range s1 {
		if stringListContains(s2, curr) {
			continue
		}

		ret = append(ret, curr)
	}

	return ret
}

func stringListContains(haystack []string, needle string) bool {
	for _, straw := range haystack {
		if straw == needle {
			return true
		}
	}
	return false
}
