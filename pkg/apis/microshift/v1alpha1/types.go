package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RemoteCluster represents a remote cluster's healthcheck probe target.
// Created by the C2CC controller, read and updated by the probe pod.
type RemoteCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RemoteClusterSpec   `json:"spec"`
	Status RemoteClusterStatus `json:"status,omitempty"`
}

type RemoteClusterSpec struct {
	// IP:port of the remote cluster's probe service (11th IP in remote service CIDR, port 8080).
	// +kubebuilder:validation:Required
	ProbeTarget string `json:"probeTarget"`

	// Interval between probe attempts (e.g. "10s", "1m").
	// +kubebuilder:default="10s"
	ProbeInterval metav1.Duration `json:"probeInterval"`
}

// RemoteClusterStatus is populated by the probe pod with health probe results.
type RemoteClusterStatus struct {
	// +kubebuilder:validation:Enum=NeverProbed;Healthy;Unhealthy
	// +kubebuilder:default="NeverProbed"
	State string `json:"state"`
	// +optional
	LastSuccessfulProbe *metav1.Time `json:"lastSuccessfulProbe,omitempty"`
	// +optional
	LastProbeTime *metav1.Time `json:"lastProbeTime,omitempty"`
	// +optional
	Errors []string `json:"errors,omitempty"`
	// Latency statistics from recent probes (rolling window of 20 samples).
	// +optional
	Latency *LatencyStats `json:"latency,omitempty"`
}

// LatencyStats contains latency statistics computed from a rolling window of probe samples.
// All duration values are serialized as Go duration strings (e.g. "1.234ms").
type LatencyStats struct {
	Avg    metav1.Duration `json:"avg"`
	Min    metav1.Duration `json:"min"`
	Max    metav1.Duration `json:"max"`
	Last   metav1.Duration `json:"last"`
	Stddev metav1.Duration `json:"stddev"`
}

// +kubebuilder:object:root=true
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RemoteClusterList contains a list of RemoteCluster resources.
type RemoteClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []RemoteCluster `json:"items"`
}
