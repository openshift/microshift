package manifestcomparators

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"

	"github.com/openshift/crd-schema-checker/pkg/resourceread"
	"gopkg.in/yaml.v2"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

func AllTestsInDir(directory string) ([]ComparatorTest, error) {
	ret := []ComparatorTest{}
	err := filepath.WalkDir(directory, func(path string, info os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}

		if containsDirectory, err := containsDir(path); err != nil {
			return err
		} else if containsDirectory {
			return nil
		}

		// so now we have only leave nodes
		relativePath, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}

		currTest, err := TestInDir(relativePath, path)
		if err != nil {
			return err
		}
		ret = append(ret, currTest)

		return nil
	})
	if err != nil {
		return nil, err
	}

	return ret, nil
}

func AllTestsInDirForComparator(comparator CRDComparator, directory string) ([]*simpleComparatorTest, error) {
	registry := NewRegistry()
	registry.AddComparator(comparator)
	return AllTestsInDirForRegistry(registry, directory)
}

func AllTestsInDirForComparators(comparators []CRDComparator, directory string) ([]*simpleComparatorTest, error) {
	registry := NewRegistry()
	for _, c := range comparators {
		registry.AddComparator(c)
	}
	return AllTestsInDirForRegistry(registry, directory)
}

func RunAllTestsInDirForComparators(t *testing.T, comparators []CRDComparator, directory string) {
	tests, err := AllTestsInDirForComparators(comparators, directory)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		t.Run(test.ComparatorTest.Name, test.Test)
	}
}

func RunAllTestsInDirForComparator(t *testing.T, comparator CRDComparator, directory string) {
	tests, err := AllTestsInDirForComparator(comparator, directory)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		t.Run(test.ComparatorTest.Name, test.Test)
	}
}

func RunAllTestsInDirForRegistry(t *testing.T, registry CRDComparatorRegistry, directory string) {
	tests, err := AllTestsInDirForRegistry(registry, directory)
	if err != nil {
		t.Fatal(err)
	}

	for _, test := range tests {
		t.Run(test.ComparatorTest.Name, test.Test)
	}
}

func AllTestsInDirForRegistry(registry CRDComparatorRegistry, directory string) ([]*simpleComparatorTest, error) {
	tests, err := AllTestsInDir(directory)
	if err != nil {
		return nil, err
	}
	ret := []*simpleComparatorTest{}

	for i := range tests {
		ret = append(ret, &simpleComparatorTest{
			ComparatorTest: tests[i],
			registry:       registry,
		})
	}

	return ret, nil
}

func TestInDir(testName, directory string) (ComparatorTest, error) {
	ret := ComparatorTest{
		Name: testName,
	}

	optionalExistingCRDFile := filepath.Join(directory, "existing.yaml")
	existingBytes, err := os.ReadFile(optionalExistingCRDFile)
	if err != nil && !os.IsNotExist(err) {
		return ComparatorTest{}, err
	}
	if len(existingBytes) > 0 {
		crd, err := resourceread.ReadCustomResourceDefinitionV1(existingBytes)
		if err != nil {
			return ComparatorTest{}, err
		}
		ret.ExistingCRD = crd
	}

	requiredNewCRDFile := filepath.Join(directory, "new.yaml")
	newBytes, err := os.ReadFile(requiredNewCRDFile)
	if err != nil {
		return ComparatorTest{}, err
	}
	newCRD, err := resourceread.ReadCustomResourceDefinitionV1(newBytes)
	if err != nil {
		return ComparatorTest{}, err
	}
	ret.NewCRD = newCRD

	optionalExpectedFile := filepath.Join(directory, "expected.yaml")
	expectedBytes, err := os.ReadFile(optionalExpectedFile)
	if err != nil && !os.IsNotExist(err) {
		return ComparatorTest{}, err
	}
	if len(expectedBytes) > 0 {
		expected := &ComparisonResultsList{}
		if err := yaml.Unmarshal(expectedBytes, expected); err != nil {
			return ComparatorTest{}, err
		}
		ret.ExpectedResults = expected.Items
	}

	optionalExpectedErrorsFile := filepath.Join(directory, "errors.txt")
	expectedErrorsBytes, err := os.ReadFile(optionalExpectedErrorsFile)
	if err != nil && !os.IsNotExist(err) {
		return ComparatorTest{}, err
	}
	if len(expectedErrorsBytes) > 0 {
		expectedErrors := []string{}
		scanner := bufio.NewScanner(bytes.NewBuffer(expectedErrorsBytes))
		for scanner.Scan() {
			expectedErrors = append(expectedErrors, scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			return ComparatorTest{}, err
		}
		ret.ExpectedErrors = expectedErrors
	}

	return ret, nil
}

func containsDir(path string) (bool, error) {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if entry.IsDir() {
			return true, nil
		}
	}
	return false, nil
}

type ComparisonResultsList struct {
	Items []ComparisonResults `yaml:"items"`
}

// ComparatorTest represents the directory style test we have.
type ComparatorTest struct {
	Name        string
	ExistingCRD *apiextensionsv1.CustomResourceDefinition
	NewCRD      *apiextensionsv1.CustomResourceDefinition

	ExpectedResults []ComparisonResults
	ExpectedErrors  []string
}

type simpleComparatorTest struct {
	ComparatorTest ComparatorTest
	registry       CRDComparatorRegistry
}

func (tc *simpleComparatorTest) Test(t *testing.T) {
	actualResults, actualErrors := tc.registry.Compare(tc.ComparatorTest.ExistingCRD, tc.ComparatorTest.NewCRD)
	tc.ComparatorTest.Test(t, actualResults, actualErrors)
}

func (tc *ComparatorTest) Test(t *testing.T, actualResults []ComparisonResults, actualErrors []error) {
	switch {
	case len(tc.ExpectedErrors) == 0 && len(actualErrors) == 0:
	case len(tc.ExpectedErrors) == 0 && len(actualErrors) != 0:
		t.Fatalf("0 errors expected, got %v", actualErrors)
	case len(tc.ExpectedErrors) != 0 && len(actualErrors) == 0:
		t.Fatalf("expected some errors: %v, got none", tc.ExpectedErrors)
	case len(tc.ExpectedErrors) != 0 && len(actualErrors) != 0:
		if !reflect.DeepEqual(tc.ExpectedErrors, actualErrors) {
			t.Fatalf("expected some errors: %v, got different errors: %v", tc.ExpectedErrors, actualErrors)
		}
	}

	// check to be sure that every expected message appeared
	for _, expected := range tc.ExpectedResults {
		expectedBytes, err := yaml.Marshal(expected)
		if err != nil {
			t.Error(err)
		}

		actualPtr := findResultsForComparator(expected.Name, actualResults)
		if actualPtr == nil {
			// this is only an error when we expect a message
			if len(expected.Errors) == 0 && len(expected.Warnings) == 0 && len(expected.Infos) == 0 {
				continue
			}
			t.Errorf("missing expectedResults[%v]: expected\n%v\n", expected.Name, string(expectedBytes))
			continue
		}

		actual := *actualPtr
		actualBytes, err := yaml.Marshal(actual)
		if err != nil {
			t.Error(err)
		}

		sort.Strings(expected.Errors)
		sort.Strings(actual.Errors)
		sort.Strings(expected.Warnings)
		sort.Strings(actual.Warnings)
		sort.Strings(expected.Infos)
		sort.Strings(actual.Infos)

		noErrorsAsExpected := len(expected.Errors) == 0 && len(actual.Errors) == 0
		if !noErrorsAsExpected && !reflect.DeepEqual(expected.Errors, actual.Errors) {
			t.Errorf("mismatched errors for expectedResults[%v]: expected\n%v\n, got\n%v\n", expected.Name, string(expectedBytes), string(actualBytes))
		}
		noWarningsAsExpected := len(expected.Warnings) == 0 && len(actual.Warnings) == 0
		if !noWarningsAsExpected && !reflect.DeepEqual(expected.Warnings, actual.Warnings) {
			t.Errorf("mismatched warnings for expectedResults[%v]: expected\n%v\n, got\n%v\n", expected.Name, string(expectedBytes), string(actualBytes))
		}
		noInfosAsExpected := len(expected.Infos) == 0 && len(actual.Infos) == 0
		if !noInfosAsExpected && !reflect.DeepEqual(expected.Infos, actual.Infos) {
			t.Errorf("mismatched infos for expectedResults[%v]: expected\n%v\n, got\n%v\n", expected.Name, string(expectedBytes), string(actualBytes))
		}
	}

	// check to be sure that we didn't get an extra message
	for _, actual := range actualResults {
		actualBytes, err := yaml.Marshal(actual)
		if err != nil {
			t.Error(err)
		}

		expectedPtr := findResultsForComparator(actual.Name, tc.ExpectedResults)
		if expectedPtr == nil {
			// this is only an error when we expect a message
			if len(actual.Errors) == 0 && len(actual.Warnings) == 0 && len(actual.Infos) == 0 {
				continue
			}
			t.Errorf("missing expectedResults for actual[%v]: got\n%v\n", actual.Name, string(actualBytes))
			continue
		}

		expected := *expectedPtr
		expectedBytes, err := yaml.Marshal(expected)
		if err != nil {
			t.Error(err)
		}
		noErrorsAsExpected := len(expected.Errors) == 0 && len(actual.Errors) == 0
		if !noErrorsAsExpected && !reflect.DeepEqual(expected.Errors, actual.Errors) {
			t.Errorf("mismatched errors for expectedResults[%v]: expected\n%v\n, got\n%v\n", expected.Name, string(expectedBytes), string(actualBytes))
		}
		noWarningsAsExpected := len(expected.Warnings) == 0 && len(actual.Warnings) == 0
		if !noWarningsAsExpected && !reflect.DeepEqual(expected.Warnings, actual.Warnings) {
			t.Errorf("mismatched warnings for expectedResults[%v]: expected\n%v\n, got\n%v\n", expected.Name, string(expectedBytes), string(actualBytes))
		}
		noInfosAsExpected := len(expected.Infos) == 0 && len(actual.Infos) == 0
		if !noInfosAsExpected && !reflect.DeepEqual(expected.Infos, actual.Infos) {
			t.Errorf("mismatched infos for expectedResults[%v]: expected\n%v\n, got\n%v\n", expected.Name, string(expectedBytes), string(actualBytes))
		}
	}

}

func findResultsForComparator(name string, results []ComparisonResults) *ComparisonResults {
	for i := range results {
		if results[i].Name == name {
			return &results[i]
		}
	}

	return nil
}
