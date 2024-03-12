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
	// +kubebuilder:default="InterNamespaceAllowed"
	NamespaceOwnership NamespaceOwnershipEnum `json:"namespaceOwnership"`
}
