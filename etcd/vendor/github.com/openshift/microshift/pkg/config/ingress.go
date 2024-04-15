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
	Status          IngressStatusEnum    `json:"status"`
	AdmissionPolicy RouteAdmissionPolicy `json:"routeAdmissionPolicy"`
	Ports           IngressPortsConfig   `json:"ports"`
	// List of IP addresses and NIC names where the router will be listening. The NIC
	// names get translated to all their configured IPs dynamically. Defaults to the
	// configured IPs in the host at MicroShift start.
	ListenAddress      []string `json:"listenAddress"`
	ServingCertificate []byte   `json:"-"`
	ServingKey         []byte   `json:"-"`
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
