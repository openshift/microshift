package utils

import (
	"fmt"
	"strings"
)

var (
	clusterProfileToShortName = map[string]string{
		"include.release.openshift.io/ibm-cloud-managed":              "Hypershift",
		"include.release.openshift.io/self-managed-high-availability": "SelfManagedHA",
	}
)

func ClusterProfileToShortName(annotation string) (string, error) {
	ret, ok := clusterProfileToShortName[annotation]
	if !ok {
		return "FAIL", fmt.Errorf("failed on %v", annotation)
	}
	return ret, nil
}

func HasClusterProfilePreference(annotations map[string]string) bool {
	for k := range annotations {
		if strings.HasPrefix(k, "include.release.openshift.io/") {
			return true
		}
	}

	return false
}
