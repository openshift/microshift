/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package field

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestMakeFuncs(t *testing.T) {
	testCases := []struct {
		fn       func() *Error
		expected ErrorType
	}{
		{
			func() *Error { return Invalid(NewPath("f"), "v", "d") },
			ErrorTypeInvalid,
		},
		{
			func() *Error { return NotSupported[string](NewPath("f"), "v", nil) },
			ErrorTypeNotSupported,
		},
		{
			func() *Error { return Duplicate(NewPath("f"), "v") },
			ErrorTypeDuplicate,
		},
		{
			func() *Error { return NotFound(NewPath("f"), "v") },
			ErrorTypeNotFound,
		},
		{
			func() *Error { return Required(NewPath("f"), "d") },
			ErrorTypeRequired,
		},
		{
			func() *Error { return InternalError(NewPath("f"), fmt.Errorf("e")) },
			ErrorTypeInternal,
		},
	}

	for _, testCase := range testCases {
		err := testCase.fn()
		if err.Type != testCase.expected {
			t.Errorf("expected Type %q, got %q", testCase.expected, err.Type)
		}
	}
}

func TestErrorUsefulMessage(t *testing.T) {
	{
		s := Invalid(nil, nil, "").Error()
		t.Logf("message: %v", s)
		if !strings.Contains(s, "null") {
			t.Errorf("error message did not contain 'null': %s", s)
		}
	}

	s := Invalid(NewPath("foo"), "bar", "deet").Error()
	t.Logf("message: %v", s)
	for _, part := range []string{"foo", "bar", "deet", ErrorTypeInvalid.String()} {
		if !strings.Contains(s, part) {
			t.Errorf("error message did not contain expected part '%v'", part)
		}
	}

	type complicated struct {
		Baz   int
		Qux   string
		Inner interface{}
		KV    map[string]int
	}
	s = Invalid(
		NewPath("foo"),
		&complicated{
			Baz:   1,
			Qux:   "aoeu",
			Inner: &complicated{Qux: "asdf"},
			KV:    map[string]int{"Billy": 2},
		},
		"detail",
	).Error()
	t.Logf("message: %v", s)
	for _, part := range []string{
		"foo", ErrorTypeInvalid.String(),
		"Baz", "Qux", "Inner", "KV", "detail",
		"1", "aoeu", "Billy", "2",
		// "asdf", TODO: re-enable once we have a better nested printer
	} {
		if !strings.Contains(s, part) {
			t.Errorf("error message did not contain expected part '%v'", part)
		}
	}
}

func TestToAggregate(t *testing.T) {
	testCases := struct {
		ErrList         []ErrorList
		NumExpectedErrs []int
	}{
		[]ErrorList{
			nil,
			{},
			{Invalid(NewPath("f"), "v", "d")},
			{Invalid(NewPath("f"), "v", "d"), Invalid(NewPath("f"), "v", "d")},
			{Invalid(NewPath("f"), "v", "d"), InternalError(NewPath(""), fmt.Errorf("e"))},
		},
		[]int{
			0,
			0,
			1,
			1,
			2,
		},
	}

	if len(testCases.ErrList) != len(testCases.NumExpectedErrs) {
		t.Errorf("Mismatch: length of NumExpectedErrs does not match length of ErrList")
	}
	for i, tc := range testCases.ErrList {
		agg := tc.ToAggregate()
		numErrs := 0

		if agg != nil {
			numErrs = len(agg.Errors())
		}
		if numErrs != testCases.NumExpectedErrs[i] {
			t.Errorf("[%d] Expected %d, got %d", i, testCases.NumExpectedErrs[i], numErrs)
		}

		if len(tc) == 0 {
			if agg != nil {
				t.Errorf("[%d] Expected nil, got %#v", i, agg)
			}
		} else if agg == nil {
			t.Errorf("[%d] Expected non-nil", i)
		}
	}
}

func TestErrListFilter(t *testing.T) {
	list := ErrorList{
		Invalid(NewPath("test.field"), "", ""),
		Invalid(NewPath("field.test"), "", ""),
		Duplicate(NewPath("test"), "value"),
	}
	if len(list.Filter(NewErrorTypeMatcher(ErrorTypeDuplicate))) != 2 {
		t.Errorf("should not filter")
	}
	if len(list.Filter(NewErrorTypeMatcher(ErrorTypeInvalid))) != 1 {
		t.Errorf("should filter")
	}
}

func TestNotSupported(t *testing.T) {
	notSupported := NotSupported(NewPath("f"), "v", []string{"a", "b", "c"})
	expected := `Unsupported value: "v": supported values: "a", "b", "c"`
	if notSupported.ErrorBody() != expected {
		t.Errorf("Expected: %s\n, but got: %s\n", expected, notSupported.ErrorBody())
	}
}

func TestErrorOrigin(t *testing.T) {
	err := Invalid(NewPath("field"), "value", "detail")

	// Test WithOrigin
	newErr := err.WithOrigin("origin1")
	if newErr.Origin != "origin1" {
		t.Errorf("Expected Origin to be 'origin1', got '%s'", newErr.Origin)
	}
	if err.Origin != "origin1" {
		t.Errorf("Expected Origin to be 'origin1', got '%s'", err.Origin)
	}
}

func TestErrorListOrigin(t *testing.T) {
	// Create an ErrorList with multiple errors
	list := ErrorList{
		Invalid(NewPath("field1"), "value1", "detail1"),
		Invalid(NewPath("field2"), "value2", "detail2"),
		Required(NewPath("field3"), "detail3"),
	}

	// Test WithOrigin
	newList := list.WithOrigin("origin1")
	// Check that WithOrigin returns the modified list
	for i, err := range newList {
		if err.Origin != "origin1" {
			t.Errorf("Error %d: Expected Origin to be 'origin2', got '%s'", i, err.Origin)
		}
	}

	// Check that the original list was also modified (WithOrigin modifies and returns the same list)
	for i, err := range list {
		if err.Origin != "origin1" {
			t.Errorf("Error %d: Expected original list Origin to be 'origin2', got '%s'", i, err.Origin)
		}
	}
}

func TestErrorMarkDeclarative(t *testing.T) {
	// Test for single Error
	err := Invalid(NewPath("field"), "value", "detail")
	if err.CoveredByDeclarative {
		t.Errorf("New error should not be declarative by default")
	}

	// Mark as declarative
	err.MarkCoveredByDeclarative() //nolint:errcheck // The "error" here is not an unexpected error from the function.
	if !err.CoveredByDeclarative {
		t.Errorf("Error should be declarative after marking")
	}
}

func TestErrorListMarkDeclarative(t *testing.T) {
	// Test for ErrorList
	list := ErrorList{
		Invalid(NewPath("field1"), "value1", "detail1"),
		Invalid(NewPath("field2"), "value2", "detail2"),
	}

	// Verify none are declarative by default
	for i, err := range list {
		if err.CoveredByDeclarative {
			t.Errorf("Error %d should not be declarative by default", i)
		}
	}

	// Mark list as declarative
	list.MarkCoveredByDeclarative()

	// Verify all errors in the list are now declarative
	for i, err := range list {
		if !err.CoveredByDeclarative {
			t.Errorf("Error %d should be declarative after marking the list", i)
		}
	}
}

func TestErrorListExtractCoveredByDeclarative(t *testing.T) {
	testCases := []struct {
		list         ErrorList
		expectedList ErrorList
	}{
		{
			ErrorList{},
			ErrorList{},
		},
		{
			ErrorList{Invalid(NewPath("field1"), nil, "")},
			ErrorList{},
		},
		{
			ErrorList{Invalid(NewPath("field1"), nil, "").MarkCoveredByDeclarative(), Required(NewPath("field2"), "detail2")},
			ErrorList{Invalid(NewPath("field1"), nil, "").MarkCoveredByDeclarative()},
		},
	}

	for _, tc := range testCases {
		got := tc.list.ExtractCoveredByDeclarative()
		if !reflect.DeepEqual(got, tc.expectedList) {
			t.Errorf("For list %v, expected %v, got %v", tc.list, tc.expectedList, got)
		}
	}
}

func TestErrorListRemoveCoveredByDeclarative(t *testing.T) {
	testCases := []struct {
		list         ErrorList
		expectedList ErrorList
	}{
		{
			ErrorList{},
			ErrorList{},
		},
		{
			ErrorList{Invalid(NewPath("field1"), nil, "").MarkCoveredByDeclarative(), Required(NewPath("field2"), "detail2")},
			ErrorList{Required(NewPath("field2"), "detail2")},
		},
	}

	for _, tc := range testCases {
		got := tc.list.RemoveCoveredByDeclarative()
		if !reflect.DeepEqual(got, tc.expectedList) {
			t.Errorf("For list %v, expected %v, got %v", tc.list, tc.expectedList, got)
		}
	}
}
