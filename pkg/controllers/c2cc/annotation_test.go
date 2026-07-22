package c2cc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAnnotationManager_SortsDesiredCIDRs(t *testing.T) {
	tests := []struct {
		name     string
		cidrs    []string
		expected []string
	}{
		{
			name:     "single cidr",
			cidrs:    []string{"10.45.0.0/16"},
			expected: []string{"10.45.0.0/16"},
		},
		{
			name:     "multiple cidrs already sorted",
			cidrs:    []string{"10.45.0.0/16", "10.46.0.0/16"},
			expected: []string{"10.45.0.0/16", "10.46.0.0/16"},
		},
		{
			name:     "multiple cidrs unsorted",
			cidrs:    []string{"172.31.0.0/16", "10.45.0.0/16", "10.46.0.0/16"},
			expected: []string{"10.45.0.0/16", "10.46.0.0/16", "172.31.0.0/16"},
		},
		{
			name:     "empty cidrs",
			cidrs:    []string{},
			expected: []string{},
		},
		{
			name:     "nil cidrs",
			cidrs:    nil,
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := newAnnotationManager(nil, "test-node", tt.cidrs)
			assert.Equal(t, tt.expected, mgr.desiredCIDRs)
			assert.Equal(t, "test-node", mgr.nodeName)
		})
	}
}

func TestNewAnnotationManager_DoesNotMutateInput(t *testing.T) {
	input := []string{"172.31.0.0/16", "10.45.0.0/16"}
	original := make([]string, len(input))
	copy(original, input)

	newAnnotationManager(nil, "test-node", input)

	assert.Equal(t, original, input, "input slice should not be modified")
}

func TestParseCIDRAnnotation(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected []string
	}{
		{name: "empty string", value: "", expected: nil},
		{name: "valid json", value: `["10.0.0.0/16","172.16.0.0/12"]`, expected: []string{"10.0.0.0/16", "172.16.0.0/12"}},
		{name: "invalid json", value: "not-json", expected: nil},
		{name: "empty array", value: "[]", expected: []string{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseCIDRAnnotation(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCidrSetContainsAll(t *testing.T) {
	tests := []struct {
		name     string
		superset []string
		subset   []string
		expected bool
	}{
		{name: "empty subset", superset: []string{"10.0.0.0/16"}, subset: nil, expected: true},
		{name: "exact match", superset: []string{"10.0.0.0/16"}, subset: []string{"10.0.0.0/16"}, expected: true},
		{name: "superset contains all", superset: []string{"10.0.0.0/16", "172.16.0.0/12"}, subset: []string{"10.0.0.0/16"}, expected: true},
		{name: "missing element", superset: []string{"10.0.0.0/16"}, subset: []string{"172.16.0.0/12"}, expected: false},
		{name: "empty superset nonempty subset", superset: nil, subset: []string{"10.0.0.0/16"}, expected: false},
		{name: "both empty", superset: nil, subset: nil, expected: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, cidrSetContainsAll(tt.superset, tt.subset))
		})
	}
}
