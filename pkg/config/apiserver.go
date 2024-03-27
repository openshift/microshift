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
	// Determines if kube-apiserver controller should configure the
	// AdvertiseAddress in the loopback interface. Automatically computed.
	SkipInterface bool `json:"-"`

	AuditLog AuditLog `json:"auditLog"`

	// The URL and Port of the API server cannot be changed by the user.
	URL  string `json:"-"`
	Port int    `json:"-"`
}

type AuditLog struct {
	// maxFileAge is the maximum number of days to retain old audit log files
	MaxFileAge  int    `json:"maxFileAge"`
	// maxFiles is the maximum number of rotated audit log files to retain
	MaxFiles    int    `json:"maxFiles"`
	// maxFileSize is the maximum size in megabytes of the audit log file before it gets rotated
	MaxFileSize int    `json:"maxFileSize"`
	// profile is the OpenShift profile specifying a specific logging policy
	// +kubebuilder:example=Default
	Profile     string `json:"profile"`
}
