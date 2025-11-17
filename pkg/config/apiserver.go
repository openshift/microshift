package config

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	configv1 "github.com/openshift/api/config/v1"
	featuresUtils "github.com/openshift/api/features"
	"github.com/openshift/library-go/pkg/crypto"
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

type FeatureGates struct {
	FeatureSet      string          `json:"featureSet"`
	CustomNoUpgrade CustomNoUpgrade `json:"customNoUpgrade"`
}

func (fg FeatureGates) ConvertToCLIFlags() ([]string, error) {
	ret := []string{}

	switch fg.FeatureSet {
	case FeatureSetCustomNoUpgrade:
		for _, feature := range fg.CustomNoUpgrade.Enabled {
			ret = append(ret, fmt.Sprintf("%s=true", feature))
		}
		for _, feature := range fg.CustomNoUpgrade.Disabled {
			ret = append(ret, fmt.Sprintf("%s=false", feature))
		}
	case FeatureSetDevPreviewNoUpgrade, FeatureSetTechPreviewNoUpgrade:
		fgEnabledDisabled, err := featuresUtils.FeatureSets(featuresUtils.SelfManaged, configv1.FeatureSet(fg.FeatureSet))
		if err != nil {
			return nil, fmt.Errorf("failed to get feature set gates: %w", err)
		}
		for _, f := range fgEnabledDisabled.Enabled {
			ret = append(ret, fmt.Sprintf("%s=true", f.FeatureGateAttributes.Name))
		}
		for _, f := range fgEnabledDisabled.Disabled {
			ret = append(ret, fmt.Sprintf("%s=false", f.FeatureGateAttributes.Name))
		}
	}
	return ret, nil
}

// Implement the GoStringer interface for better %#v printing
func (fg FeatureGates) GoString() string {
	return fmt.Sprintf("FeatureGates{FeatureSet: %q, CustomNoUpgrade: %#v}", fg.FeatureSet, fg.CustomNoUpgrade)
}

func (fg *FeatureGates) validateFeatureGates() error {
	// FG is unset
	if fg == nil || reflect.DeepEqual(*fg, FeatureGates{}) {
		return nil
	}
	// Must use a recognized feature set, or else empty
	if fg.FeatureSet != "" && fg.FeatureSet != FeatureSetCustomNoUpgrade && fg.FeatureSet != FeatureSetTechPreviewNoUpgrade && fg.FeatureSet != FeatureSetDevPreviewNoUpgrade {
		return fmt.Errorf("invalid feature set: %s", fg.FeatureSet)
	}
	// Must set FeatureSet to CustomNoUpgrade to use custom feature gates
	if fg.FeatureSet != FeatureSetCustomNoUpgrade && (len(fg.CustomNoUpgrade.Enabled) > 0 || len(fg.CustomNoUpgrade.Disabled) > 0) {
		return fmt.Errorf("CustomNoUpgrade must be empty when FeatureSet is empty")
	}
	// Must set CustomNoUpgrade enabled or disabled lists when FeatureSet is CustomNoUpgrade
	if fg.FeatureSet == FeatureSetCustomNoUpgrade && len(fg.CustomNoUpgrade.Enabled) == 0 && len(fg.CustomNoUpgrade.Disabled) == 0 {
		return fmt.Errorf("CustomNoUpgrade enabled or disabled lists must be set when FeatureSet is CustomNoUpgrade")
	}
	// Must not have any feature gates that are enabled and disabled at the same time
	var illegalFeatures []string
	for _, enabledFeature := range fg.CustomNoUpgrade.Enabled {
		if slices.Contains(fg.CustomNoUpgrade.Disabled, enabledFeature) {
			illegalFeatures = append(illegalFeatures, enabledFeature)
		}
	}
	if len(illegalFeatures) > 0 {
		return fmt.Errorf("featuregates cannot be enabled and disabled at the same time: %s", strings.Join(illegalFeatures, ", "))
	}
	return nil
}
