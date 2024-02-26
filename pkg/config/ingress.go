package config

const (
	NamespaceOwnershipStrict  = "Strict"
	NamespaceOwnershipAllowed = "InterNamespaceAllowed"
)

type NamespaceOwnershipEnum string

type IngressConfig struct {
	AdmissionPolicy    RouteAdmissionPolicy `json:"routeAdmissionPolicy"`
	ServingCertificate []byte               `json:"-"`
	ServingKey         []byte               `json:"-"`
}

type RouteAdmissionPolicy struct {
	NamespaceOwnership NamespaceOwnershipEnum `json:"namespaceOwnership"`
}
