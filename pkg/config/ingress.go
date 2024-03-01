package config

const (
	NamespaceOwnershipStrict  = "Strict"
	NamespaceOwnershipAllowed = "InterNamespaceAllowed"
	StatusEnabled             = "Enabled"
	StatusDisabled            = "Disabled"
)

type NamespaceOwnershipEnum string
type IngressStatusEnum string

type IngressConfig struct {
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
