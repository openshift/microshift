package config

type ApiServer struct {
	// SubjectAltNames added to API server certs
	SubjectAltNames []string `json:"subjectAltNames"`
	// Kube apiserver advertise address to work around the certificates issue
	// when requiring external access using the node IP. This will turn into
	// the IP configured in the endpoint slice for kubernetes service. Must be
	// a reachable IP from pods. Defaults to service network CIDR first
	// address.
	AdvertiseAddress string `json:"advertiseAddress,omitempty"`
	// List of custom certificates used to secure requests to specific host names
	NamedCertificates []NamedCertificateEntry `json:"namedCertificates"`
	// Determines if kube-apiserver controller should configure the
	// AdvertiseAddress in the loopback interface. Automatically computed.
	SkipInterface bool `json:"-"`

	// The URL and Port of the API server cannot be changed by the user.
	URL  string `json:"-"`
	Port int    `json:"-"`
}

// NamedCertificateEntry provides certificate details
type NamedCertificateEntry struct {
	Names    []string `json:"names"`
	CertPath string   `json:"certPath"`
	KeyPath  string   `json:"keyPath"`
}
