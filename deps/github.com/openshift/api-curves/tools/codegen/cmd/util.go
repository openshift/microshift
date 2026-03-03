package main

import (
	"fmt"

	"github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/sets"
)

var _ pflag.Value = &enum{}

// newEnum create a new enum flag value.
func newEnum(value *string, allowedValues ...string) *enum {
	return &enum{
		value:         value,
		allowedValues: sets.NewString(allowedValues...),
	}
}

// enum implements the pflag.Value interface.

type enum struct {
	value         *string
	allowedValues sets.String
}

// String returns the string representation of the enum value.
func (e *enum) String() string {
	if e.value == nil {
		return ""
	}
	return *e.value
}

// Type returns the type of the enum value.
func (e *enum) Type() string {
	return "enum"
}

// Set sets the enum value.
// It returns an error if the value is not allowed.
func (e *enum) Set(value string) error {
	if !e.allowedValues.Has(value) {
		return fmt.Errorf("invalid value %q, allowed values are %v", value, e.allowedValues.List())
	}

	*e.value = value
	return nil
}
