package apivalidation

import (
	"strings"

	"k8s.io/apimachinery/pkg/api/validation/path"
)

func ValidateUserName(name string, _ bool) []string {
	if reasons := path.ValidatePathSegmentName(name, false); len(reasons) != 0 {
		return reasons
	}

	if strings.Contains(name, ":") {
		return []string{`may not contain ":"`}
	}
	if name == "~" {
		return []string{`may not equal "~"`}
	}
	return nil
}

func ValidateGroupName(name string, _ bool) []string {
	if reasons := path.ValidatePathSegmentName(name, false); len(reasons) != 0 {
		return reasons
	}

	if strings.Contains(name, ":") {
		return []string{`may not contain ":"`}
	}
	if name == "~" {
		return []string{`may not equal "~"`}
	}
	return nil
}
