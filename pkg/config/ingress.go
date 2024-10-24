package config

import (
	operatorv1 "github.com/openshift/api/operator/v1"
)

const (
	NamespaceOwnershipStrict  NamespaceOwnershipEnum   = "Strict"
	NamespaceOwnershipAllowed NamespaceOwnershipEnum   = "InterNamespaceAllowed"
	StatusManaged             IngressStatusEnum        = "Managed"
	StatusRemoved             IngressStatusEnum        = "Removed"
	DefaultHttpVersionV1      DefaultHttpVersionPolicy = 1
	DefaultHttpVersionV2      DefaultHttpVersionPolicy = 2
)

type NamespaceOwnershipEnum string
type IngressStatusEnum string
type DefaultHttpVersionPolicy int32
type IngressConfig struct {
	// Default router status, can be Managed or Removed.
	// +kubebuilder:default=Managed
	Status          IngressStatusEnum                         `json:"status"`
	AdmissionPolicy RouteAdmissionPolicy                      `json:"routeAdmissionPolicy"`
	Ports           IngressPortsConfig                        `json:"ports"`
	TuningOptions   operatorv1.IngressControllerTuningOptions `json:"tuningOptions"`
	// List of IP addresses and NIC names where the router will be listening. The NIC
	// names get translated to all their configured IPs dynamically. Defaults to the
	// configured IPs in the host at MicroShift start.
	ListenAddress      []string `json:"listenAddress"`
	ServingCertificate []byte   `json:"-"`
	ServingKey         []byte   `json:"-"`
	// logEmptyRequests specifies how connections on which no request is
	// received should be logged.  Typically, these empty requests come from
	// load balancers' health probes or Web browsers' speculative
	// connections ("preconnect"), in which case logging these requests may
	// be undesirable.  However, these requests may also be caused by
	// network errors, in which case logging empty requests may be useful
	// for diagnosing the errors.  In addition, these requests may be caused
	// by port scans, in which case logging empty requests may aid in
	// detecting intrusion attempts.  Allowed values for this field are
	// "Log" and "Ignore".  The default value is "Log".
	//
	// +optional
	// +kubebuilder:default:="Log"
	LogEmptyRequests operatorv1.LoggingPolicy `json:"logEmptyRequests,omitempty"`

	// forwardedHeaderPolicy specifies when and how ingress router
	// sets the Forwarded, X-Forwarded-For, X-Forwarded-Host,
	// X-Forwarded-Port, X-Forwarded-Proto, and X-Forwarded-Proto-Version
	// HTTP headers.  The value may be one of the following:
	//
	// * "Append", which specifies that ingress router appends the
	//   headers, preserving existing headers.
	//
	// * "Replace", which specifies that ingress router sets the
	//   headers, replacing any existing Forwarded or X-Forwarded-* headers.
	//
	// * "IfNone", which specifies that ingress router sets the
	//   headers if they are not already set.
	//
	// * "Never", which specifies that ingress router never sets the
	//   headers, preserving any existing headers.
	//
	// By default, the policy is "Append".
	//
	// +optional
	ForwardedHeaderPolicy operatorv1.IngressControllerHTTPHeaderPolicy `json:"forwardedHeaderPolicy,omitempty"`

	// httpEmptyRequestsPolicy describes how HTTP connections should be
	// handled if the connection times out before a request is received.
	// Allowed values for this field are "Respond" and "Ignore".  If the
	// field is set to "Respond", the ingress controller sends an HTTP 400
	// or 408 response, logs the connection (if access logging is enabled),
	// and counts the connection in the appropriate metrics.  If the field
	// is set to "Ignore", the ingress controller closes the connection
	// without sending a response, logging the connection, or incrementing
	// metrics.  The default value is "Respond".
	//
	// Typically, these connections come from load balancers' health probes
	// or Web browsers' speculative connections ("preconnect") and can be
	// safely ignored.  However, these requests may also be caused by
	// network errors, and so setting this field to "Ignore" may impede
	// detection and diagnosis of problems.  In addition, these requests may
	// be caused by port scans, in which case logging empty requests may aid
	// in detecting intrusion attempts.
	//
	// +optional
	// +kubebuilder:default:="Respond"
	HTTPEmptyRequestsPolicy operatorv1.HTTPEmptyRequestsPolicy `json:"httpEmptyRequestsPolicy,omitempty"`

	// httpCompression defines a policy for HTTP traffic compression.
	// By default, there is no HTTP compression.
	//
	// +optional
	HTTPCompressionPolicy operatorv1.HTTPCompressionPolicy `json:"httpCompression,omitempty"`

	// Determines default http version should be used for the ingress backends
	// By default,  using version 1.
	//
	// +optional
	// +kubebuilder:default:="1"
	DefaultHttpVersionPolicy DefaultHttpVersionPolicy `json:"defaultHTTPVersion,omitempty"`
}

type RouteAdmissionPolicy struct {
	// Describes how host name claims across namespaces should be handled.
	//
	// Value must be one of:
	//
	// - Strict: Do not allow routes in different namespaces to claim the same host.
	//
	// - InterNamespaceAllowed: Allow routes to claim different paths of the same
	//   host name across namespaces.
	//
	// If empty, the default is InterNamespaceAllowed.
	// +kubebuilder:default="InterNamespaceAllowed"
	NamespaceOwnership NamespaceOwnershipEnum `json:"namespaceOwnership"`
}

type IngressPortsConfig struct {
	// Default router http port. Must be in range 1-65535.
	// +kubebuilder:default=80
	Http *int `json:"http"`
	// Default router https port. Must be in range 1-65535.
	// +kubebuilder:default=443
	Https *int `json:"https"`
}
