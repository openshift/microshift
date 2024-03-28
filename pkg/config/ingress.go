package config

const (
	NamespaceOwnershipStrict  NamespaceOwnershipEnum = "Strict"
	NamespaceOwnershipAllowed NamespaceOwnershipEnum = "InterNamespaceAllowed"
	StatusEnabled             IngressStatusEnum      = "Enabled"
	StatusDisabled            IngressStatusEnum      = "Disabled"
)

type NamespaceOwnershipEnum string
type IngressStatusEnum string

type IngressConfig struct {
	// Default router status, can be Enabled or Disabled.
	// +kubebuilder:default=Enabled
	Status             IngressStatusEnum    `json:"status"`
	AdmissionPolicy    RouteAdmissionPolicy `json:"routeAdmissionPolicy"`
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
