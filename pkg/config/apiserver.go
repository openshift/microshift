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

	AuditLog AuditLog `json:"auditLog"`

	// The URL and Port of the API server cannot be changed by the user.
	URL  string `json:"-"`
	Port int    `json:"-"`

	// In dual stack mode, ovnk requires ovn.OVNGatewayInterface to have one IP
	// per family or else it wont start. When configuring advertiseAddress,
	// whether that is manual or automatic, this IP is configured in that
	// bridge afterwards in node package. Since there is only one IP, ovnk will
	// return an error complaining about the other IP family for the secondary
	// cluster/service network gateway. This variable holds all the different
	// IP addresses that ovn.OVNGatewayInterface needs. Note that this IP is
	// not configurable by users and it will not be used for apiserver
	// advertising because of dual stack limitations there. This is only to
	// make ovnk work properly.
	AdvertiseAddresses []string `json:"-"`

	TLS TLSConfig `json:"tls"`
}

// NamedCertificateEntry provides certificate details
type NamedCertificateEntry struct {
	Names    []string `json:"names"`
	CertPath string   `json:"certPath"`
	KeyPath  string   `json:"keyPath"`
}

type AuditLog struct {
	// maxFileAge is the maximum number of days to retain old audit log files
	// +kubebuilder:default=0
	MaxFileAge int `json:"maxFileAge"`
	// maxFiles is the maximum number of rotated audit log files to retain
	// +kubebuilder:default=10
	MaxFiles int `json:"maxFiles"`
	// maxFileSize is the maximum size in megabytes of the audit log file before it gets rotated
	// +kubebuilder:default=200
	MaxFileSize int `json:"maxFileSize"`
	// profile is the OpenShift profile specifying a specific logging policy
	// +kubebuilder:default=Default
	Profile string `json:"profile"`
}

type TLSConfig struct {
	// CipherSuites lists the allowed cipher suites that the API server will
	// accept and serve.
	CipherSuites []string `json:"cipherSuites"`

	// MinVersion specifies which TLS version is the minimum version of TLS
	// to serve from the API server.
	// +kubebuilder:validation:Enum:=v1.2;v1.3
	// +kubebuilder:default=v1.2
	MinVersion string `json:"minVersion"`
}
