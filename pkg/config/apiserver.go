package config

import (
	"fmt"
	"reflect"
	"slices"

	configv1 "github.com/openshift/api/config/v1"
	"github.com/openshift/library-go/pkg/crypto"

	"k8s.io/apimachinery/pkg/util/sets"
)

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

	TLS TLSConfig `json:"tls"`

	FeatureGates FeatureGates `json:"featureGates"`

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
	// accept and serve. Defaults to cipher suites from the minVersion config
	// parameter.
	CipherSuites []string `json:"cipherSuites"`

	// MinVersion specifies which TLS version is the minimum version of TLS
	// to serve from the API server. Allowed values: VersionTLS12, VersionTLS13.
	// Defaults to VersionTLS12.
	// +kubebuilder:validation:Enum:=VersionTLS12;VersionTLS13
	// +kubebuilder:default=VersionTLS12
	MinVersion string `json:"minVersion"`
}

func (t *TLSConfig) UpdateValues() {
	if t.MinVersion == "" {
		t.MinVersion = string(configv1.VersionTLS12)
	}
	switch t.MinVersion {
	case string(configv1.VersionTLS12):
		if len(t.CipherSuites) == 0 {
			t.CipherSuites = getIANACipherSuites(configv1.TLSProfiles[configv1.TLSProfileIntermediateType].Ciphers)
		} else {
			// Either of these cipher suites are required for TLS 1.2 in Golang, include one of them if user didnt.
			if !slices.Contains(t.CipherSuites, "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256") && !slices.Contains(t.CipherSuites, "TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256") {
				t.CipherSuites = append(t.CipherSuites, "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256")
			}
		}
	case string(configv1.VersionTLS13):
		// Golang does not allow specifying cipher suites when using tls 1.3, so we
		// override whatever the user has configured to match the cipher suites that
		// will actually be used. Note that this is only for informational purposes,
		// Golang will ignore whatever is configured.
		t.CipherSuites = getIANACipherSuites(configv1.TLSProfiles[configv1.TLSProfileModernType].Ciphers)
	}
}

func (t *TLSConfig) Validate() error {
	if len(t.CipherSuites) == 0 {
		return fmt.Errorf("unsupported empty cipher suites")
	}
	var cipherSuites []string
	switch t.MinVersion {
	case string(configv1.VersionTLS12):
		cipherSuites = getIANACipherSuites(configv1.TLSProfiles[configv1.TLSProfileIntermediateType].Ciphers)
	case string(configv1.VersionTLS13):
		cipherSuites = getIANACipherSuites(configv1.TLSProfiles[configv1.TLSProfileModernType].Ciphers)
	default:
		return fmt.Errorf("unsupported value %s for tls.MinVersion", t.MinVersion)
	}
	for _, suite := range t.CipherSuites {
		if !slices.Contains(cipherSuites, suite) {
			return fmt.Errorf("unsupported cipher suite %s for TLS version %s", suite, t.MinVersion)
		}
	}
	return nil
}

func getIANACipherSuites(suites []string) []string {
	return crypto.OpenSSLToIANACipherSuites(suites)
}

const (
	FeatureSetCustomNoUpgrade      = "CustomNoUpgrade"
	FeatureSetTechPreviewNoUpgrade = "TechPreviewNoUpgrade"
	FeatureSetDevPreviewNoUpgrade  = "DevPreviewNoUpgrade"
)

type CustomNoUpgrade struct {
	Enabled  []string `json:"enabled"`
	Disabled []string `json:"disabled"`
}

// RequiredFeatureGates are the feature gates that are always enabled by MicroShift. They are defined here to enable config validation.
// They are injected into the feature-gates field later by the microshift kube-apiserver controller.
var RequiredFeatureGates = []string{"UserNamespacesSupport", "UserNamespacesPodSecurityStandards"}

type FeatureGates struct {
	FeatureSet      string          `json:"featureSet"`
	CustomNoUpgrade CustomNoUpgrade `json:"customNoUpgrade"`
}

// ToApiserverArgs converts the FeatureGates struct to a list of feature-gates arguments for the kube-apiserver.
// Validation checks should be performed before calling this function to ensure the FeatureGates struct is valid.
func (fg FeatureGates) ToApiserverArgs() ([]string, error) {
	ret := sets.NewString()
	addFeatures := func(features []string, enabled bool) {
		for _, feature := range features {
			ret.Insert(fmt.Sprintf("%s=%t", feature, enabled))
		}
	}

	addFeatures(fg.CustomNoUpgrade.Enabled, true)
	addFeatures(fg.CustomNoUpgrade.Disabled, false)
	return ret.List(), nil
}

// Implement the GoStringer interface for better %#v printing
func (fg FeatureGates) GoString() string {
	return fmt.Sprintf("FeatureGates{FeatureSet: %q, CustomNoUpgrade: %#v}", fg.FeatureSet, fg.CustomNoUpgrade)
}

// validateFeatureGates validates the FeatureGates struct according to the following rules:
// 1. FeatureGates may be unset.
// 2. FeatureSet must be empty or CustomNoUpgrade.
// 3. If FeatureSet is DevPreviewNoUpgrade or TechPreviewNoUpgrade, return an error.
// 4. If FeatureSet is CustomNoUpgrade, CustomNoUpgrade.Enabled/Disabled lists may be set but are not required.
// 5. Required feature gates cannot be disabled.
// 6. Feature gates cannot be both enabled and disabled within the same object.
func (fg *FeatureGates) validateFeatureGates() error {
	if fg == nil || reflect.DeepEqual(*fg, FeatureGates{}) {
		return nil
	}

	switch fg.FeatureSet {
	case "":
		return nil
	case FeatureSetCustomNoUpgrade:
		// Valid - continue with validation
	case FeatureSetDevPreviewNoUpgrade, FeatureSetTechPreviewNoUpgrade:
		return fmt.Errorf("FeatureSet %s is not supported. Use CustomNoUpgrade to enable/disable feature gates", fg.FeatureSet)
	default:
		return fmt.Errorf("invalid feature set: %s", fg.FeatureSet)
	}

	enabledCustom := sets.New(fg.CustomNoUpgrade.Enabled...)
	disabledCustom := sets.New(fg.CustomNoUpgrade.Disabled...)

	// checkFeatureGateConflict checks if two sets of feature gates have any intersection and returns an error if they do.
	checkFeatureGateConflict := func(a, b sets.Set[string], errorMsg string) error {
		if intersect := a.Intersection(b); intersect.Len() > 0 {
			return fmt.Errorf("%s: %s", errorMsg, intersect.UnsortedList())
		}
		return nil
	}

	conflictChecks := []struct {
		setA sets.Set[string]
		setB sets.Set[string]
		msg  string
	}{
		{disabledCustom, sets.New(RequiredFeatureGates...), "required feature gates cannot be disabled"},
		{enabledCustom, disabledCustom, "feature gates cannot be both enabled and disabled"},
	}

	for _, check := range conflictChecks {
		if err := checkFeatureGateConflict(check.setA, check.setB, check.msg); err != nil {
			return err
		}
	}

	return nil
}
