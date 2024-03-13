package config

const (
	NamespaceOwnershipStrict  NamespaceOwnershipEnum = "Strict"
	NamespaceOwnershipAllowed NamespaceOwnershipEnum = "InterNamespaceAllowed"
	StatusManaged             IngressStatusEnum      = "Managed"
	StatusRemoved             IngressStatusEnum      = "Removed"
)

type NamespaceOwnershipEnum string
type IngressStatusEnum string

type IngressConfig struct {
	// Default router status, can be Managed or Removed.
	// +kubebuilder:default=Managed
	Status             IngressStatusEnum    `json:"status"`
	AdmissionPolicy    RouteAdmissionPolicy `json:"routeAdmissionPolicy"`
	Ports              IngressPortsConfig   `json:"ports"`
	ServingCertificate []byte               `json:"-"`
	ServingKey         []byte               `json:"-"`
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
	// Default router http port.
	// +kubebuilder:default=80
	Http uint16 `json:"http"`
	// Default router https port.
	// +kubebuilder:default=443
	Https uint16 `json:"https"`
}
