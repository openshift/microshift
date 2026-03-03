// Copyright 2025 The Kubernetes Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// 	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package validations

import (
	"errors"
	"fmt"

	"github.com/google/go-cmp/cmp"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"sigs.k8s.io/crdify/pkg/config"
)

// CompareVersions calculates the diff in the provided old and new CustomResourceDefinitionVersions and
// compares the differing properties using the provided comparators.
// An 'unhandled' comparator will be injected to evaluate any unhandled changes by the provided comparators
// that will be enforced based on the provided unhandled enforcement policy.
// Returns a map[string][]ComparisonResult, where the map key is the flattened property path (i.e ^.spec.foo.bar).
func CompareVersions(a, b apiextensionsv1.CustomResourceDefinitionVersion, unhandledEnforcement config.EnforcementPolicy, comparators ...Comparator[apiextensionsv1.JSONSchemaProps]) map[string][]ComparisonResult {
	oldFlattened := FlattenCRDVersion(a)
	newFlattened := FlattenCRDVersion(b)

	diffs := FlattenedCRDVersionDiff(oldFlattened, newFlattened)

	result := map[string][]ComparisonResult{}

	for property, diff := range diffs {
		result[property] = CompareProperties(diff.Old, diff.New, unhandledEnforcement, comparators...)
	}

	return result
}

// CompareProperties compares the provided JSONSchemaProps using the provided comparators.
// An 'unhandled' comparator will be injected to evaluate any unhandled changes by the provided
// comparators that will be enforced based on the provided unhandled enforcement policy.
// Returns a slice containing all the comparison results.
func CompareProperties(a, b *apiextensionsv1.JSONSchemaProps, unhandledEnforcement config.EnforcementPolicy, comparators ...Comparator[apiextensionsv1.JSONSchemaProps]) []ComparisonResult {
	result := []ComparisonResult{}
	aCopy, bCopy := a.DeepCopy(), b.DeepCopy()

	for _, comparator := range comparators {
		comparisonResult := comparator.Compare(aCopy, bCopy)
		result = append(result, comparisonResult)
	}

	// checking for unhandled changes is _always_ performed last.
	result = append(result, checkUnhandled(aCopy, bCopy, unhandledEnforcement))

	return result
}

// checkUnhandled is a utility function for checking if a provided set of comparators
// handled validating all differences between the JSONSchemaProps.
// It returns a ComparisonResult so that the results are treated generically just like a standard Comparator.
func checkUnhandled(a, b *apiextensionsv1.JSONSchemaProps, enforcement config.EnforcementPolicy) ComparisonResult {
	var err error

	if !equality.Semantic.DeepEqual(a, b) {
		diff := cmp.Diff(a, b)
		err = fmt.Errorf("%w :\n%s", ErrUnhandledChangesFound, diff)
	}

	return HandleErrors("unhandled", enforcement, err)
}

// ErrUnhandledChangesFound represents an error state where changes have been found that are not
// handled by an existing validation check.
var ErrUnhandledChangesFound = errors.New("unhandled changes found")
