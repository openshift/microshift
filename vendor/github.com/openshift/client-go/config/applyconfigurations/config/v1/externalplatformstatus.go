// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1

// ExternalPlatformStatusApplyConfiguration represents a declarative configuration of the ExternalPlatformStatus type for use
// with apply.
type ExternalPlatformStatusApplyConfiguration struct {
	CloudControllerManager *CloudControllerManagerStatusApplyConfiguration `json:"cloudControllerManager,omitempty"`
}

// ExternalPlatformStatusApplyConfiguration constructs a declarative configuration of the ExternalPlatformStatus type for use with
// apply.
func ExternalPlatformStatus() *ExternalPlatformStatusApplyConfiguration {
	return &ExternalPlatformStatusApplyConfiguration{}
}

// WithCloudControllerManager sets the CloudControllerManager field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the CloudControllerManager field is set to the value of the last call.
func (b *ExternalPlatformStatusApplyConfiguration) WithCloudControllerManager(value *CloudControllerManagerStatusApplyConfiguration) *ExternalPlatformStatusApplyConfiguration {
	b.CloudControllerManager = value
	return b
}
