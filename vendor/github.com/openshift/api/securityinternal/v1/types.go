package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RangeAllocation is used so we can easily expose a RangeAllocation typed for security group
// This is an internal API, not intended for external consumption.
//
// Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer).
// +openshift:compatibility-gen:level=1
type RangeAllocation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// range is a string representing a unique label for a range of uids, "1000000000-2000000000/10000".
	Range string `json:"range"`

	// data is a byte array representing the serialized state of a range allocation.  It is a bitmap
	// with each bit set to one to represent a range is taken.
	Data []byte `json:"data"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RangeAllocationList is a list of RangeAllocations objects
// This is an internal API, not intended for external consumption.
//
// Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer).
// +openshift:compatibility-gen:level=1
type RangeAllocationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	// List of RangeAllocations.
	Items []RangeAllocation `json:"items"`
}
