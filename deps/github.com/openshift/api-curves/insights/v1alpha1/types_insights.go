package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=datagathers,scope=Cluster
// +kubebuilder:subresource:status
// +openshift:api-approved.openshift.io=https://github.com/openshift/api/pull/1365
// +openshift:file-pattern=cvoRunLevel=0000_10,operatorName=insights,operatorOrdering=01
// +openshift:enable:FeatureGate=InsightsOnDemandDataGather
// +kubebuilder:printcolumn:name=State,type=string,JSONPath=.status.dataGatherState,description=DataGather job state
// +kubebuilder:printcolumn:name=StartTime,type=date,JSONPath=.status.startTime,description=DataGather start time
// +kubebuilder:printcolumn:name=FinishTime,type=date,JSONPath=.status.finishTime,description=DataGather finish time
// +openshift:capability=Insights
//
// DataGather provides data gather configuration options and status for the particular Insights data gathering.
//
// Compatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support.
// +openshift:compatibility-gen:level=4
type DataGather struct {
	metav1.TypeMeta `json:",inline"`
	// metadata is the standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// spec holds user settable values for configuration
	// +required
	Spec DataGatherSpec `json:"spec"`
	// status holds observed values from the cluster. They may not be overridden.
	// +optional
	Status DataGatherStatus `json:"status"`
}

// DataGatherSpec contains the configuration for the DataGather.
type DataGatherSpec struct {
	// dataPolicy allows user to enable additional global obfuscation of the IP addresses and base domain
	// in the Insights archive data. Valid values are "ClearText" and "ObfuscateNetworking".
	// When set to ClearText the data is not obfuscated.
	// When set to ObfuscateNetworking the IP addresses and the cluster domain name are obfuscated.
	// When omitted, this means no opinion and the platform is left to choose a reasonable default, which is subject to change over time.
	// The current default is ClearText.
	// +optional
	DataPolicy DataPolicy `json:"dataPolicy"`
	// gatherers is an optional list of gatherers configurations.
	// The list must not exceed 100 items.
	// The particular gatherers IDs can be found at https://github.com/openshift/insights-operator/blob/master/docs/gathered-data.md.
	// Run the following command to get the names of last active gatherers:
	// "oc get insightsoperators.operator.openshift.io cluster -o json | jq '.status.gatherStatus.gatherers[].name'"
	// +kubebuilder:validation:MaxItems=100
	// +optional
	Gatherers []GathererConfig `json:"gatherers,omitempty"`
	// storage is an optional field that allows user to define persistent storage for gathering jobs to store the Insights data archive.
	// If omitted, the gathering job will use ephemeral storage.
	// +optional
	Storage *Storage `json:"storage,omitempty"`
}

// storage provides persistent storage configuration options for gathering jobs.
// If the type is set to PersistentVolume, then the PersistentVolume must be defined.
// If the type is set to Ephemeral, then the PersistentVolume must not be defined.
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'PersistentVolume' ?  has(self.persistentVolume) : !has(self.persistentVolume)",message="persistentVolume is required when type is PersistentVolume, and forbidden otherwise"
type Storage struct {
	// type is a required field that specifies the type of storage that will be used to store the Insights data archive.
	// Valid values are "PersistentVolume" and "Ephemeral".
	// When set to Ephemeral, the Insights data archive is stored in the ephemeral storage of the gathering job.
	// When set to PersistentVolume, the Insights data archive is stored in the PersistentVolume that is
	// defined by the PersistentVolume field.
	// +required
	Type StorageType `json:"type"`
	// persistentVolume is an optional field that specifies the PersistentVolume that will be used to store the Insights data archive.
	// The PersistentVolume must be created in the openshift-insights namespace.
	// +optional
	PersistentVolume *PersistentVolumeConfig `json:"persistentVolume,omitempty"`
}

// storageType declares valid storage types
// +kubebuilder:validation:Enum=PersistentVolume;Ephemeral
type StorageType string

const (
	// StorageTypePersistentVolume storage type
	StorageTypePersistentVolume StorageType = "PersistentVolume"
	// StorageTypeEphemeral storage type
	StorageTypeEphemeral StorageType = "Ephemeral"
)

// persistentVolumeConfig provides configuration options for PersistentVolume storage.
type PersistentVolumeConfig struct {
	// claim is a required field that specifies the configuration of the PersistentVolumeClaim that will be used to store the Insights data archive.
	// The PersistentVolumeClaim must be created in the openshift-insights namespace.
	// +required
	Claim PersistentVolumeClaimReference `json:"claim"`
	// mountPath is an optional field specifying the directory where the PVC will be mounted inside the Insights data gathering Pod.
	// When omitted, this means no opinion and the platform is left to choose a reasonable default, which is subject to change over time.
	// The current default mount path is /var/lib/insights-operator
	// The path may not exceed 1024 characters and must not contain a colon.
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:XValidation:rule="!self.contains(':')",message="mountPath must not contain a colon"
	// +optional
	MountPath string `json:"mountPath,omitempty"`
}

// persistentVolumeClaimReference is a reference to a PersistentVolumeClaim.
type PersistentVolumeClaimReference struct {
	// name is a string that follows the DNS1123 subdomain format.
	// It must be at most 253 characters in length, and must consist only of lower case alphanumeric characters, '-' and '.', and must start and end with an alphanumeric character.
	// +kubebuilder:validation:XValidation:rule="!format.dns1123Subdomain().validate(self).hasValue()",message="a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character."
	// +kubebuilder:validation:MaxLength:=253
	// +required
	Name string `json:"name"`
}

const (
	// No data obfuscation
	NoPolicy DataPolicy = "ClearText"
	// IP addresses and cluster domain name are obfuscated
	ObfuscateNetworking DataPolicy = "ObfuscateNetworking"
	// Data gathering is running
	Running DataGatherState = "Running"
	// Data gathering is completed
	Completed DataGatherState = "Completed"
	// Data gathering failed
	Failed DataGatherState = "Failed"
	// Data gathering is pending
	Pending DataGatherState = "Pending"
	// Gatherer state marked as disabled, which means that the gatherer will not run.
	Disabled GathererState = "Disabled"
	// Gatherer state marked as enabled, which means that the gatherer will run.
	Enabled GathererState = "Enabled"
)

// dataPolicy declares valid data policy types
// +kubebuilder:validation:Enum="";ClearText;ObfuscateNetworking
type DataPolicy string

// state declares valid gatherer state types.
// +kubebuilder:validation:Enum="";Enabled;Disabled
type GathererState string

// gathererConfig allows to configure specific gatherers
type GathererConfig struct {
	// name is the required name of specific gatherer
	// It must be at most 256 characters in length.
	// The format for the gatherer name should be: {gatherer}/{function} where the function is optional.
	// Gatherer consists of a lowercase letters only that may include underscores (_).
	// Function consists of a lowercase letters only that may include underscores (_) and is separated from the gatherer by a forward slash (/).
	// The particular gatherers can be found at https://github.com/openshift/insights-operator/blob/master/docs/gathered-data.md.
	// +kubebuilder:validation:MaxLength=256
	// +kubebuilder:validation:XValidation:rule=`self.matches("^[a-z]+[_a-z]*[a-z]([/a-z][_a-z]*)?[a-z]$")`,message=`gatherer name must be in the format of {gatherer}/{function} where the gatherer and function are lowercase letters only that may include underscores (_) and are separated by a forward slash (/) if the function is provided`
	// +required
	Name string `json:"name"`
	// state allows you to configure specific gatherer. Valid values are "Enabled", "Disabled" and omitted.
	// When omitted, this means no opinion and the platform is left to choose a reasonable default.
	// The current default is Enabled.
	// +optional
	State GathererState `json:"state"`
}

// dataGatherState declares valid gathering state types
// +kubebuilder:validation:Enum=Running;Completed;Failed;Pending
// +kubebuilder:validation:XValidation:rule="!(oldSelf == 'Running' && self == 'Pending')", message="dataGatherState cannot transition from Running to Pending"
// +kubebuilder:validation:XValidation:rule="!(oldSelf == 'Completed' && self == 'Pending')", message="dataGatherState cannot transition from Completed to Pending"
// +kubebuilder:validation:XValidation:rule="!(oldSelf == 'Failed' && self == 'Pending')", message="dataGatherState cannot transition from Failed to Pending"
// +kubebuilder:validation:XValidation:rule="!(oldSelf == 'Completed' && self == 'Running')", message="dataGatherState cannot transition from Completed to Running"
// +kubebuilder:validation:XValidation:rule="!(oldSelf == 'Failed' && self == 'Running')", message="dataGatherState cannot transition from Failed to Running"
type DataGatherState string

// DataGatherStatus contains information relating to the DataGather state.
// +kubebuilder:validation:XValidation:rule="(!has(oldSelf.insightsRequestID) || has(self.insightsRequestID))",message="cannot remove insightsRequestID attribute from status"
// +kubebuilder:validation:XValidation:rule="(!has(oldSelf.startTime) || has(self.startTime))",message="cannot remove startTime attribute from status"
// +kubebuilder:validation:XValidation:rule="(!has(oldSelf.finishTime) || has(self.finishTime))",message="cannot remove finishTime attribute from status"
// +kubebuilder:validation:XValidation:rule="(!has(oldSelf.dataGatherState) || has(self.dataGatherState))",message="cannot remove dataGatherState attribute from status"
type DataGatherStatus struct {
	// conditions provide details on the status of the gatherer job.
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=100
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// dataGatherState reflects the current state of the data gathering process.
	// +optional
	State DataGatherState `json:"dataGatherState,omitempty"`
	// gatherers is a list of active gatherers (and their statuses) in the last gathering.
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=100
	// +optional
	Gatherers []GathererStatus `json:"gatherers,omitempty"`
	// startTime is the time when Insights data gathering started.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="startTime is immutable once set"
	// +optional
	StartTime metav1.Time `json:"startTime,omitempty"`
	// finishTime is the time when Insights data gathering finished.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="finishTime is immutable once set"
	// +optional
	FinishTime metav1.Time `json:"finishTime,omitempty"`
	// relatedObjects is a list of resources which are useful when debugging or inspecting the data
	// gathering Pod
	// +kubebuilder:validation:MaxItems=100
	// +optional
	RelatedObjects []ObjectReference `json:"relatedObjects,omitempty"`
	// insightsRequestID is an Insights request ID to track the status of the
	// Insights analysis (in console.redhat.com processing pipeline) for the corresponding Insights data archive.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="insightsRequestID is immutable once set"
	// +kubebuilder:validation:MaxLength=256
	// +optional
	InsightsRequestID string `json:"insightsRequestID,omitempty"`
	// insightsReport provides general Insights analysis results.
	// When omitted, this means no data gathering has taken place yet or the
	// corresponding Insights analysis (identified by "insightsRequestID") is not available.
	// +optional
	InsightsReport InsightsReport `json:"insightsReport,omitempty"`
}

// gathererStatus represents information about a particular
// data gatherer.
type GathererStatus struct {
	// conditions provide details on the status of each gatherer.
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:MaxItems=100
	// +required
	Conditions []metav1.Condition `json:"conditions"`
	// name is the name of the gatherer.
	// +required
	// +kubebuilder:validation:MaxLength=256
	// +kubebuilder:validation:MinLength=5
	Name string `json:"name"`
	// lastGatherDuration represents the time spent gathering.
	// +required
	// +kubebuilder:validation:Type=string
	// +kubebuilder:validation:Pattern="^(([0-9]+(?:\\.[0-9]+)?(ns|us|µs|μs|ms|s|m|h))+)$"
	LastGatherDuration metav1.Duration `json:"lastGatherDuration"`
}

// insightsReport provides Insights health check report based on the most
// recently sent Insights data.
type InsightsReport struct {
	// downloadedAt is the time when the last Insights report was downloaded.
	// An empty value means that there has not been any Insights report downloaded yet and
	// it usually appears in disconnected clusters (or clusters when the Insights data gathering is disabled).
	// +optional
	DownloadedAt metav1.Time `json:"downloadedAt,omitempty"`
	// healthChecks provides basic information about active Insights health checks
	// in a cluster.
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=100
	// +optional
	HealthChecks []HealthCheck `json:"healthChecks,omitempty"`
	// uri is optional field that provides the URL link from which the report was downloaded.
	// The link must be a valid HTTPS URL and the maximum length is 2048 characters.
	// +kubebuilder:validation:XValidation:rule=`isURL(self) && url(self).getScheme() == "https"`,message=`URI must be a valid HTTPS URL (e.g., https://example.com)`
	// +kubebuilder:validation:MaxLength=2048
	// +optional
	URI string `json:"uri,omitempty"`
}

// healthCheck represents an Insights health check attributes.
type HealthCheck struct {
	// description provides basic description of the healtcheck.
	// +required
	// +kubebuilder:validation:MaxLength=2048
	// +kubebuilder:validation:MinLength=10
	Description string `json:"description"`
	// totalRisk of the healthcheck. Indicator of the total risk posed
	// by the detected issue; combination of impact and likelihood. The values can be from 1 to 4,
	// and the higher the number, the more important the issue.
	// +required
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:validation:Maximum=4
	TotalRisk int32 `json:"totalRisk"`
	// advisorURI is required field that provides the URL link to the Insights Advisor.
	// The link must be a valid HTTPS URL and the maximum length is 2048 characters.
	// +kubebuilder:validation:XValidation:rule=`isURL(self) && url(self).getScheme() == "https"`,message=`advisorURI must be a valid HTTPS URL (e.g., https://example.com)`
	// +kubebuilder:validation:MaxLength=2048
	// +required
	AdvisorURI string `json:"advisorURI"`
	// state determines what the current state of the health check is.
	// Health check is enabled by default and can be disabled
	// by the user in the Insights advisor user interface.
	// +required
	State HealthCheckState `json:"state"`
}

// healthCheckState provides information about the status of the
// health check (for example, the health check may be marked as disabled by the user).
// +kubebuilder:validation:Enum:=Enabled;Disabled
type HealthCheckState string

const (
	// enabled marks the health check as enabled
	HealthCheckEnabled HealthCheckState = "Enabled"
	// disabled marks the health check as disabled
	HealthCheckDisabled HealthCheckState = "Disabled"
)

// ObjectReference contains enough information to let you inspect or modify the referred object.
type ObjectReference struct {
	// group is the API Group of the Resource.
	// Enter empty string for the core group.
	// This value is empty or should follow the DNS1123 subdomain format and it must be at most 253 characters in length.
	// Example: "", "apps", "build.openshift.io", etc.
	// +kubebuilder:validation:XValidation:rule="self.size() == 0 || !format.dns1123Subdomain().validate(self).hasValue()",message="a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character."
	// +kubebuilder:validation:MaxLength:=253
	// +required
	Group string `json:"group"`
	// resource is required field of the type that is being referenced.
	// It is normally the plural form of the resource kind in lowercase.
	// This value should consist of only lowercase alphanumeric characters and hyphens.
	// Example: "deployments", "deploymentconfigs", "pods", etc.
	// +kubebuilder:validation:XValidation:rule=`self.matches("^[a-z0-9]([-a-z0-9]*[a-z0-9])?$")`,message=`resource must consist of only lowercase alphanumeric characters and hyphens`
	// +kubebuilder:validation:MaxLength=512
	// +required
	Resource string `json:"resource"`
	// name of the referent that follows the DNS1123 subdomain format.
	// It must be at most 256 characters in length.
	// +kubebuilder:validation:XValidation:rule="!format.dns1123Subdomain().validate(self).hasValue()",message="a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character."
	// +kubebuilder:validation:MaxLength=256
	// +required
	Name string `json:"name"`
	// namespace of the referent that follows the DNS1123 subdomain format.
	// It must be at most 253 characters in length.
	// +kubebuilder:validation:XValidation:rule="!format.dns1123Subdomain().validate(self).hasValue()",message="a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character."
	// +kubebuilder:validation:MaxLength=253
	// +optional
	Namespace string `json:"namespace,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataGatherList is a collection of items
//
// Compatibility level 4: No compatibility is provided, the API can change at any point for any reason. These capabilities should not be used by applications needing long term support.
// +openshift:compatibility-gen:level=4
type DataGatherList struct {
	metav1.TypeMeta `json:",inline"`
	// metadata is the standard list's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	// items contains a list of DataGather resources.
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=100
	// +optional
	Items []DataGather `json:"items,omitempty"`
}
