package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +openshift:compatibility-gen:level=1
// +openshift:api-approved.openshift.io=https://github.com/openshift/api/pull/xxx
// +openshift:file-pattern=cvoRunLevel=0000_50,operatorName=my-operator,operatorOrdering=01

// StableConfigType is a stable config type that may include TechPreviewNoUpgrade fields.
//
// Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer).
// +openshift:compatibility-gen:level=1
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=stableconfigtypes,scope=Cluster
// +kubebuilder:subresource:status
type StableConfigType struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ObjectMeta `json:"metadata,omitempty"`

	// spec is the specification of the desired behavior of the StableConfigType.
	Spec StableConfigTypeSpec `json:"spec,omitempty"`
	// status is the most recently observed status of the StableConfigType.
	Status StableConfigTypeStatus `json:"status,omitempty"`
}

// StableConfigTypeSpec is the desired state
// +openshift:validation:FeatureGateAwareXValidation:featureGate=Example,rule="has(oldSelf.coolNewField) ? has(self.coolNewField) : true",message="coolNewField may not be removed once set"
// +openshift:validation:FeatureGateAwareXValidation:requiredFeatureGate=Example;Example2,rule="has(oldSelf.stableField) ? has(self.stableField) : true",message="stableField may not be removed once set (this should only show up with both the Example and Example2 feature gates)"
type StableConfigTypeSpec struct {
	// coolNewField is a field that is for tech preview only.  On normal clusters this shouldn't be present
	//
	// +openshift:enable:FeatureGate=Example
	// +optional
	CoolNewField string `json:"coolNewField"`

	// stableField is a field that is present on default clusters and on tech preview clusters
	//
	// If empty, the platform will choose a good default, which may change over time without notice.
	//
	// +optional
	StableField string `json:"stableField"`

	// immutableField is a field that is immutable once the object has been created.
	// It is required at all times.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="immutableField is immutable"
	// +required
	ImmutableField string `json:"immutableField"`

	// optionalImmutableField is a field that is immutable once set.
	// It is optional but may not be changed once set.
	// +kubebuilder:validation:XValidation:rule="oldSelf == '' || self == oldSelf",message="optionalImmutableField is immutable once set"
	// +optional
	OptionalImmutableField string `json:"optionalImmutableField"`

	// evolvingUnion demonstrates how to phase in new values into discriminated union
	// +optional
	EvolvingUnion EvolvingUnion `json:"evolvingUnion"`

	// celUnion demonstrates how to validate a discrminated union using CEL
	// +optional
	CELUnion CELUnion `json:"celUnion,omitempty"`

	// nonZeroDefault is a demonstration of creating an integer field that has a non zero default.
	// It required two default tags (one for CRD generation, one for client generation) and must have `omitempty` and be optional.
	// A minimum value is added to demonstrate that a zero value would not be accepted.
	// +kubebuilder:default:=8
	// +default=8
	// +kubebuilder:validation:Minimum:=8
	// +optional
	NonZeroDefault int32 `json:"nonZeroDefault,omitempty"`

	// evolvingCollection demonstrates how to have a collection where the maximum number of items varies on cluster type.
	// For default clusters, this will be "1" but on TechPreview clusters, this value will be "3".
	// +openshift:validation:FeatureGateAwareMaxItems:featureGate="",maxItems=1
	// +openshift:validation:FeatureGateAwareMaxItems:featureGate=Example,maxItems=3
	// +optional
	// +listType=atomic
	EvolvingCollection []string `json:"evolvingCollection,omitempty"`

	// set demonstrates how to define and validate set of strings
	// +optional
	Set StringSet `json:"set,omitempty"`

	// subdomainNameField represents a kubenetes name field.
	// The intention is that it validates the name in the same way metadata.Name is validated.
	// That is, it is a DNS-1123 subdomain.
	// +kubebuilder:validation:XValidation:rule="!format.dns1123Subdomain().validate(self).hasValue()",message="a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character."
	// +kubebuilder:validation:MaxLength:=253
	// +optional
	SubdomainNameField string `json:"subdomainNameField,omitempty"`

	// subnetsWithExclusions demonstrates how to validate a list of subnets with exclusions
	// +optional
	SubnetsWithExclusions SubnetsWithExclusions `json:"subnetsWithExclusions,omitempty"`

	// formatMarkerExamples demonstrates all Kubebuilder Format markers supported as of Kubernetes 1.33.
	// This field serves as a comprehensive reference for format marker validation.
	// +optional
	FormatMarkerExamples *FormatMarkerExamples `json:"formatMarkerExamples,omitempty"`
}

// SetValue defines the types allowed in string set type
// +kubebuilder:validation:Enum:=Foo;Bar;Baz;Qux;Corge
type SetValue string

// StringSet defines the set of strings
// +listType=set
// +kubebuilder:validation:XValidation:rule="self.all(x,self.exists_one(y,x == y))"
// +kubebuilder:validation:MaxItems=5
type StringSet []SetValue

type EvolvingUnion struct {
	// type is the discriminator. It has different values for Default and for TechPreviewNoUpgrade
	// +required
	Type EvolvingDiscriminator `json:"type"`
}

// EvolvingDiscriminator defines the audit policy profile type.
// +openshift:validation:FeatureGateAwareEnum:featureGate="",enum="";StableValue
// +openshift:validation:FeatureGateAwareEnum:featureGate=Example,enum="";StableValue;TechPreviewOnlyValue
type EvolvingDiscriminator string

const (
	// "StableValue" is always present.
	StableValue EvolvingDiscriminator = "StableValue"

	// "TechPreviewOnlyValue" should only be allowed when TechPreviewNoUpgrade is set in the cluster
	TechPreviewOnlyValue EvolvingDiscriminator = "TechPreviewOnlyValue"
)

// CELUnion demonstrates how to use a discriminated union and how to validate it using CEL.
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'RequiredMember' ?  has(self.requiredMember) : !has(self.requiredMember)",message="requiredMember is required when type is RequiredMember, and forbidden otherwise"
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'OptionalMember' ?  true : !has(self.optionalMember)",message="optionalMember is forbidden when type is not OptionalMember"
// +union
type CELUnion struct {
	// type determines which of the union members should be populated.
	// +required
	// +unionDiscriminator
	Type CELUnionDiscriminator `json:"type"`

	// requiredMember is a union member that is required.
	// +unionMember
	RequiredMember *string `json:"requiredMember,omitempty"`

	// optionalMember is a union member that is optional.
	// +unionMember,optional
	OptionalMember *string `json:"optionalMember,omitempty"`
}

// CELUnionDiscriminator is a union discriminator for the CEL union.
// +kubebuilder:validation:Enum:="RequiredMember";"OptionalMember";"EmptyMember"
type CELUnionDiscriminator string

const (
	// RequiredMember represents a required union member.
	RequiredMember CELUnionDiscriminator = "RequiredMember"

	// OptionalMember represents an optional union member.
	OptionalMember CELUnionDiscriminator = "OptionalMember"

	// EmptyMember represents an empty union member.
	EmptyMember CELUnionDiscriminator = "EmptyMember"
)

// StableConfigTypeStatus defines the observed status of the StableConfigType.
type StableConfigTypeStatus struct {
	// Represents the observations of a foo's current state.
	// Known .status.conditions.type are: "Available", "Progressing", and "Degraded"
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`

	// immutableField is a field that is immutable once the object has been created.
	// It is required at all times.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="immutableField is immutable"
	// +optional
	ImmutableField string `json:"immutableField,omitempty"`
}

// SubnetsWithExclusions is used to validate a list of subnets with exclusions.
// It demonstrates how exclusions should be validated as subnetworks of the networks listed in the subnets field.
// +kubebuilder:validation:XValidation:rule="!has(self.excludeSubnets) || self.excludeSubnets.all(e, self.subnets.exists(s, cidr(s).containsCIDR(cidr(e))))",message="excludeSubnets must be subnetworks of the networks specified in the subnets field",fieldPath=".excludeSubnets"
type SubnetsWithExclusions struct {
	// subnets is a list of subnets.
	// It may contain up to 2 subnets.
	// The list may be either 1 IPv4 subnet, 1 IPv6 subnet, or 1 of each.
	// +kubebuilder:validation:XValidation:rule="size(self) != 2 || !isCIDR(self[0]) || !isCIDR(self[1]) || cidr(self[0]).ip().family() != cidr(self[1]).ip().family()",message="subnets must not contain 2 subnets of the same IP family"
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=2
	// +listType=atomic
	// +required
	Subnets []CIDR `json:"subnets"`

	// excludeSubnets is a list of CIDR exclusions.
	// The subnets in this list must be subnetworks of the subnets in the subnets list.
	// +kubebuilder:validation:MaxItems=25
	// +optional
	// +listType=atomic
	ExcludeSubnets []CIDR `json:"excludeSubnets,omitempty"`
}

// CIDR is used to validate a CIDR notation network.
// The longest CIDR notation is 43 characters.
// +kubebuilder:validation:XValidation:rule="isCIDR(self)",message="value must be a valid CIDR"
// +kubebuilder:validation:MaxLength:=43
type CIDR string

// FormatMarkerExamples demonstrates all Kubebuilder Format markers supported as of Kubernetes 1.33.
// This struct provides a comprehensive reference for format marker validation.
// Each field uses a different format marker to validate its value.
type FormatMarkerExamples struct {
	// ipv4Address must be a valid IPv4 address in dotted-quad notation.
	// Valid values range from 0.0.0.0 to 255.255.255.255 (e.g., 192.168.1.1).
	//
	// Use of Format=ipv4 is not recommended due to CVE-2021-29923 and CVE-2024-24790.
	// Instead, use the CEL expression `isIP(self) && ip(self).family() == 4` to validate IPv4 addresses.
	//
	// +kubebuilder:validation:Format=ipv4
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=15
	// +optional
	IPv4Address string `json:"ipv4Address,omitempty"`

	// ipv6Address must be a valid IPv6 address.
	// Valid examples include full form (2001:0db8:0000:0000:0000:0000:0000:0001) or compressed form (2001:db8::1 or ::1).
	//
	// Use of Format=ipv6 is not recommended due to CVE-2021-29923 and CVE-2024-24790.
	// Instead, use the CEL expression `isIP(self) && ip(self).family() == 6` to validate IPv6 addresses.
	//
	// +kubebuilder:validation:Format=ipv6
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=45
	// +optional
	IPv6Address string `json:"ipv6Address,omitempty"`

	// cidrNotation must be a valid CIDR notation IP address range.
	// Valid examples include IPv4 CIDR (10.0.0.0/8, 192.168.1.0/24) or IPv6 CIDR (fd00::/8, 2001:db8::/32).
	//
	// Use of Format=cidr is not recommended due to CVE-2021-29923 and CVE-2024-24790.
	// Instead, use the CEL expression `isCIDR(self)` to validate CIDR notation.
	// Additionally, use `isCIDR(self) && cidr(self).ip().family() == X` to validate IPvX specifically.
	//
	// +kubebuilder:validation:Format=cidr
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=49
	// +optional
	CIDRNotation string `json:"cidrNotation,omitempty"`

	// uriField must be a valid URI following RFC 3986 syntax.
	// Valid examples include https://example.com/path?query=value or /absolute-path.
	// +kubebuilder:validation:Format=uri
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	// +optional
	URIField string `json:"uriField,omitempty"`

	// emailAddress must be a valid email address.
	// Valid examples include user@example.com or firstname.lastname@company.co.uk.
	// +kubebuilder:validation:Format=email
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=254
	// +optional
	EmailAddress string `json:"emailAddress,omitempty"`

	// hostnameField must be a valid Internet hostname per RFC 1034.
	// Valid examples include example.com, api.example.com, or my-service.
	// +kubebuilder:validation:Format=hostname
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +optional
	HostnameField string `json:"hostnameField,omitempty"`

	// macAddress must be a valid MAC address.
	// Valid examples include 00:1A:2B:3C:4D:5E or 00-1A-2B-3C-4D-5E.
	// +kubebuilder:validation:Format=mac
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=17
	// +optional
	MACAddress string `json:"macAddress,omitempty"`

	// uuidField must be a valid UUID (any version) in 8-4-4-4-12 format.
	// Valid examples include 550e8400-e29b-41d4-a716-446655440000 or 123e4567-e89b-12d3-a456-426614174000.
	// +kubebuilder:validation:Format=uuid
	// +kubebuilder:validation:MinLength=36
	// +kubebuilder:validation:MaxLength=36
	// +optional
	UUIDField string `json:"uuidField,omitempty"`

	// uuid3Field must be a valid UUID version 3 (MD5 hash-based).
	// Version 3 UUIDs are generated using MD5 hashing of a namespace and name.
	// Valid example: a3bb189e-8bf9-3888-9912-ace4e6543002.
	// +kubebuilder:validation:Format=uuid3
	// +kubebuilder:validation:MinLength=36
	// +kubebuilder:validation:MaxLength=36
	// +optional
	UUID3Field string `json:"uuid3Field,omitempty"`

	// uuid4Field must be a valid UUID version 4 (random).
	// Version 4 UUIDs are randomly generated.
	// Valid example: 550e8400-e29b-41d4-a716-446655440000.
	// +kubebuilder:validation:Format=uuid4
	// +kubebuilder:validation:MinLength=36
	// +kubebuilder:validation:MaxLength=36
	// +optional
	UUID4Field string `json:"uuid4Field,omitempty"`

	// uuid5Field must be a valid UUID version 5 (SHA-1 hash-based).
	// Version 5 UUIDs are generated using SHA-1 hashing of a namespace and name.
	// Valid example: 74738ff5-5367-5958-9aee-98fffdcd1876.
	// +kubebuilder:validation:Format=uuid5
	// +kubebuilder:validation:MinLength=36
	// +kubebuilder:validation:MaxLength=36
	// +optional
	UUID5Field string `json:"uuid5Field,omitempty"`

	// dateField must be a valid date in RFC 3339 full-date format (YYYY-MM-DD).
	// Valid examples include 2024-01-15 or 2023-12-31.
	// +kubebuilder:validation:Format=date
	// +kubebuilder:validation:MinLength=10
	// +kubebuilder:validation:MaxLength=10
	// +optional
	DateField string `json:"dateField,omitempty"`

	// dateTimeField must be a valid RFC 3339 date-time.
	// Valid examples include 2024-01-15T14:30:00Z, 2024-01-15T14:30:00+00:00, or 2024-01-15T14:30:00.123Z.
	// +kubebuilder:validation:Format=date-time
	// +kubebuilder:validation:MinLength=20
	// +kubebuilder:validation:MaxLength=35
	// +optional
	DateTimeField string `json:"dateTimeField,omitempty"`

	// durationField must be a valid duration string parseable by Go's time.ParseDuration.
	// Valid time units are ns, us (or µs), ms, s, m, h.
	// Valid examples include 30s, 5m, 1h30m, 100ms, or 1h.
	// +kubebuilder:validation:Format=duration
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=255
	// +optional
	DurationField string `json:"durationField,omitempty"`

	// base64Data must be valid base64-encoded data.
	// Valid examples include aGVsbG8= (encodes "hello") or SGVsbG8gV29ybGQh (encodes "Hello World!").
	// +kubebuilder:validation:Format=byte
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	// +optional
	Base64Data string `json:"base64Data,omitempty"`

	// passwordField is a marker for sensitive data.
	// Note that the password format marker does not perform any actual validation - it accepts any string value.
	// This marker is primarily used to signal that the field contains sensitive information.
	// +kubebuilder:validation:Format=password
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=255
	// +optional
	PasswordField string `json:"passwordField,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +openshift:compatibility-gen:level=1

// StableConfigTypeList contains a list of StableConfigTypes.
//
// Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer).
// +openshift:compatibility-gen:level=1
type StableConfigTypeList struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is the standard list's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []StableConfigType `json:"items"`
}
