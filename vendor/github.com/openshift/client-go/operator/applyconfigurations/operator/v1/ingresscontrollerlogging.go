// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

// IngressControllerLoggingApplyConfiguration represents a declarative configuration of the IngressControllerLogging type for use
// with apply.
type IngressControllerLoggingApplyConfiguration struct {
	Access *AccessLoggingApplyConfiguration `json:"access,omitempty"`
}

// IngressControllerLoggingApplyConfiguration constructs a declarative configuration of the IngressControllerLogging type for use with
// apply.
func IngressControllerLogging() *IngressControllerLoggingApplyConfiguration {
	return &IngressControllerLoggingApplyConfiguration{}
}

// WithAccess sets the Access field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Access field is set to the value of the last call.
func (b *IngressControllerLoggingApplyConfiguration) WithAccess(value *AccessLoggingApplyConfiguration) *IngressControllerLoggingApplyConfiguration {
	b.Access = value
	return b
}
