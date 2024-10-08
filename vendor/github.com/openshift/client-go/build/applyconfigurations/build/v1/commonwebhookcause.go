// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

// CommonWebHookCauseApplyConfiguration represents a declarative configuration of the CommonWebHookCause type for use
// with apply.
type CommonWebHookCauseApplyConfiguration struct {
	Revision *SourceRevisionApplyConfiguration `json:"revision,omitempty"`
	Secret   *string                           `json:"secret,omitempty"`
}

// CommonWebHookCauseApplyConfiguration constructs a declarative configuration of the CommonWebHookCause type for use with
// apply.
func CommonWebHookCause() *CommonWebHookCauseApplyConfiguration {
	return &CommonWebHookCauseApplyConfiguration{}
}

// WithRevision sets the Revision field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Revision field is set to the value of the last call.
func (b *CommonWebHookCauseApplyConfiguration) WithRevision(value *SourceRevisionApplyConfiguration) *CommonWebHookCauseApplyConfiguration {
	b.Revision = value
	return b
}

// WithSecret sets the Secret field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Secret field is set to the value of the last call.
func (b *CommonWebHookCauseApplyConfiguration) WithSecret(value string) *CommonWebHookCauseApplyConfiguration {
	b.Secret = &value
	return b
}
