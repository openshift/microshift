package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

const (
	APIVersion                   = "microshift.openshift.io/v1alpha1"
	CertificateStatusListKind    = "CertificateStatusList"
	CertificateRenewalResultKind = "CertificateRenewalResult"
	ErrorKind                    = "Error"
)

// CertificateRole identifies how a managed certificate is used.
// +kubebuilder:validation:Enum=ca;serving;client;peer
type CertificateRole string

const (
	CertificateRoleCA      CertificateRole = "ca"
	CertificateRoleServing CertificateRole = "serving"
	CertificateRoleClient  CertificateRole = "client"
	CertificateRolePeer    CertificateRole = "peer"
)

// RotationPolicy identifies the thresholds used to calculate a certificate zone.
// +kubebuilder:validation:Enum=standard;extended
type RotationPolicy string

const (
	RotationPolicyStandard RotationPolicy = "standard"
	RotationPolicyExtended RotationPolicy = "extended"
)

// CertificateZone identifies the current renewal urgency of a certificate.
// +kubebuilder:validation:Enum=green;yellow;red
type CertificateZone string

const (
	CertificateZoneGreen  CertificateZone = "green"
	CertificateZoneYellow CertificateZone = "yellow"
	CertificateZoneRed    CertificateZone = "red"
)

// CertificateStatusList reports the state of all managed certificates.
type CertificateStatusList struct {

	// +kubebuilder:validation:Enum="microshift.openshift.io/v1alpha1"
	APIVersion string `json:"apiVersion"`

	// +kubebuilder:validation:Enum=CertificateStatusList
	Kind string `json:"kind"`

	// GeneratedAt is the time used to calculate zones and remaining validity.
	GeneratedAt metav1.Time             `json:"generatedAt"`
	Config      CertificateStatusConfig `json:"config"`

	// Items are sorted by service and then certificate name.
	Items []CertificateStatusItem `json:"items"`

	// Warnings is an empty array when no warnings apply.
	Warnings []string `json:"warnings"`
}

// CertificateStatusConfig reports the effective certificate policy.
type CertificateStatusConfig struct {
	ForceRestartOnRedZone bool `json:"forceRestartOnRedZone"`

	// ServingValidity is a positive Go duration string.
	// +kubebuilder:validation:MinLength=1
	ServingValidity string `json:"servingValidity"`

	// CAValidity is a positive Go duration string.
	// +kubebuilder:validation:MinLength=1
	CAValidity string `json:"caValidity"`
}

// CertificateStatusItem reports one managed certificate without private material.
type CertificateStatusItem struct {

	// +kubebuilder:validation:MinLength=1
	Service string `json:"service"`

	// +kubebuilder:validation:MinLength=1
	Name           string          `json:"name"`
	Role           CertificateRole `json:"role"`
	RotationPolicy RotationPolicy  `json:"rotationPolicy"`
	Zone           CertificateZone `json:"zone"`
	NotBefore      metav1.Time     `json:"notBefore"`
	NotAfter       metav1.Time     `json:"notAfter"`
	// RemainingSeconds is negative when the certificate has expired.
	RemainingSeconds int64 `json:"remainingSeconds"`
}

// RenewalMode identifies the complete certificate category selected for renewal.
// +kubebuilder:validation:Enum=serving;ca
type RenewalMode string

const (
	RenewalModeServing RenewalMode = "serving"
	RenewalModeCA      RenewalMode = "ca"
)

// RenewalStatus distinguishes a validated dry-run from a committed renewal.
// +kubebuilder:validation:Enum=validated;completed
type RenewalStatus string

const (
	RenewalStatusValidated RenewalStatus = "validated"
	RenewalStatusCompleted RenewalStatus = "completed"
)

// CertificateRenewalResult describes a successful dry-run or completed renewal.
// Failed operations return an Error document instead.
// +kubebuilder:validation:XValidation:rule="self.status == 'validated' ? self.dryRun : !self.dryRun",message="validated results require dryRun=true; completed results require dryRun=false"
// +kubebuilder:validation:XValidation:rule="self.items.all(item, item.changed == (self.status == 'completed'))",message="changed must be false for every validated item and true for every completed item"
type CertificateRenewalResult struct {

	// +kubebuilder:validation:Enum="microshift.openshift.io/v1alpha1"
	APIVersion string `json:"apiVersion"`

	// +kubebuilder:validation:Enum=CertificateRenewalResult
	Kind        string        `json:"kind"`
	GeneratedAt metav1.Time   `json:"generatedAt"`
	Mode        RenewalMode   `json:"mode"`
	Status      RenewalStatus `json:"status"`
	DryRun      bool          `json:"dryRun"`

	// +kubebuilder:validation:MinItems=1
	Items  []CertificateRenewalItem `json:"items"`
	Impact CertificateRenewalImpact `json:"impact"`

	// Warnings is an empty array when no warnings apply.
	Warnings []string `json:"warnings"`
}

// CertificateRenewalItem describes the before and after expiry of one certificate.
type CertificateRenewalItem struct {

	// +kubebuilder:validation:MinLength=1
	Service string `json:"service"`

	// +kubebuilder:validation:MinLength=1
	Name string          `json:"name"`
	Role CertificateRole `json:"role"`

	// ParentCA is null for a root CA.
	// +nullable
	ParentCA        *string     `json:"parentCA"`
	CurrentNotAfter metav1.Time `json:"currentNotAfter"`

	// NewNotAfter is proposed for validated results and applied for completed results.
	NewNotAfter metav1.Time `json:"newNotAfter"`
	Changed     bool        `json:"changed"`
}

// CertificateRenewalImpact describes the actions required after renewal.
type CertificateRenewalImpact struct {
	ServiceRestartRequired           bool `json:"serviceRestartRequired"`
	KubeconfigRedistributionRequired bool `json:"kubeconfigRedistributionRequired"`
	ApplicationReloadMayBeRequired   bool `json:"applicationReloadMayBeRequired"`
}

// ErrorCode identifies the failure class without requiring callers to parse messages.
// +kubebuilder:validation:Enum=InvalidArguments;InvalidConfiguration;InsufficientPrivileges;MicroShiftRunning;CertificateInventoryFailed;RenewalFailed;RecoveryFailed;InternalError
type ErrorCode string

const (
	ErrorCodeInvalidArguments           ErrorCode = "InvalidArguments"
	ErrorCodeInvalidConfiguration       ErrorCode = "InvalidConfiguration"
	ErrorCodeInsufficientPrivileges     ErrorCode = "InsufficientPrivileges"
	ErrorCodeMicroShiftRunning          ErrorCode = "MicroShiftRunning"
	ErrorCodeCertificateInventoryFailed ErrorCode = "CertificateInventoryFailed"
	ErrorCodeRenewalFailed              ErrorCode = "RenewalFailed"
	ErrorCodeRecoveryFailed             ErrorCode = "RecoveryFailed"
	ErrorCodeInternalError              ErrorCode = "InternalError"
)

// Error is the versioned certificate-command failure document.
type Error struct {

	// +kubebuilder:validation:Enum="microshift.openshift.io/v1alpha1"
	APIVersion string `json:"apiVersion"`

	// +kubebuilder:validation:Enum=Error
	Kind        string      `json:"kind"`
	GeneratedAt metav1.Time `json:"generatedAt"`
	Code        ErrorCode   `json:"code"`

	// +kubebuilder:validation:MinLength=1
	Message string `json:"message"`

	// Details holds code-specific context, or null when none is available.
	// +nullable
	// +kubebuilder:validation:Schemaless
	// +kubebuilder:validation:Type=object
	// +kubebuilder:pruning:PreserveUnknownFields
	Details map[string]interface{} `json:"details"`
}
