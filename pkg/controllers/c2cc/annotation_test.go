package c2cc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewAnnotationManager_SortsAndMarshalsCIDRs(t *testing.T) {
	tests := []struct {
		name     string
		cidrs    []string
		expected string
	}{
		{
			name:     "single cidr",
			cidrs:    []string{"10.45.0.0/16"},
			expected: `["10.45.0.0/16"]`,
		},
		{
			name:     "multiple cidrs already sorted",
			cidrs:    []string{"10.45.0.0/16", "10.46.0.0/16"},
			expected: `["10.45.0.0/16","10.46.0.0/16"]`,
		},
		{
			name:     "multiple cidrs unsorted",
			cidrs:    []string{"172.31.0.0/16", "10.45.0.0/16", "10.46.0.0/16"},
			expected: `["10.45.0.0/16","10.46.0.0/16","172.31.0.0/16"]`,
		},
		{
			name:     "empty cidrs",
			cidrs:    []string{},
			expected: `[]`,
		},
		{
			name:     "nil cidrs",
			cidrs:    nil,
			expected: `[]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mgr := newAnnotationManager(nil, "test-node", tt.cidrs)
			assert.Equal(t, tt.expected, mgr.desiredAnnotation)
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
