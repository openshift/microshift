package config

const (
	NamespaceOwnershipStrict  = "Strict"
	NamespaceOwnershipAllowed = "InterNamespaceAllowed"
)

type NamespaceOwnershipEnum string

type IngressConfig struct {
	AdmissionPolicy    RouteAdmissionPolicy `json:"routeAdmissionPolicy"`
	Ports              IngressPortsConfig   `json:"ports"`
	ServingCertificate []byte               `json:"-"`
	ServingKey         []byte               `json:"-"`
}

type RouteAdmissionPolicy struct {
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
