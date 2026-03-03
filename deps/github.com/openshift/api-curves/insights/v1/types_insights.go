package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DataGather provides data gather configuration options and status for the particular Insights data gathering.
//
// +genclient
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:object:root=true
// +kubebuilder:resource:path=datagathers,scope=Cluster
// +kubebuilder:subresource:status
// +openshift:api-approved.openshift.io=https://github.com/openshift/api/pull/2448
// +openshift:file-pattern=cvoRunLevel=0000_10,operatorName=insights,operatorOrdering=01
// +openshift:enable:FeatureGate=InsightsOnDemandDataGather
// +kubebuilder:printcolumn:name=StartTime,type=date,JSONPath=.status.startTime,description=DataGather start time
// +kubebuilder:printcolumn:name=FinishTime,type=date,JSONPath=.status.finishTime,description=DataGather finish time
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp",description="DataGather age"
// +openshift:capability=Insights
//
// Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer).
// +openshift:compatibility-gen:level=1
type DataGather struct {
	metav1.TypeMeta `json:",inline"`
	// metadata is the standard object's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`
	// spec holds user settable values for configuration
	// +required
	Spec DataGatherSpec `json:"spec,omitempty,omitzero"`
	// status holds observed values from the cluster. They may not be overridden.
	// +optional
	Status DataGatherStatus `json:"status,omitempty,omitzero"`
}

// DataGatherSpec contains the configuration for the DataGather.
type DataGatherSpec struct {
	// dataPolicy is an optional list of DataPolicyOptions that allows user to enable additional obfuscation of the Insights archive data.
	// It may not exceed 2 items and must not contain duplicates.
	// Valid values are ObfuscateNetworking and WorkloadNames.
	// When set to ObfuscateNetworking the IP addresses and the cluster domain name are obfuscated.
	// When set to WorkloadNames, the gathered data about cluster resources will not contain the workload names for your deployments. Resources UIDs will be used instead.
	// When omitted no obfuscation is applied.
	// +kubebuilder:validation:MaxItems=2
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:XValidation:rule="self.all(x, self.exists_one(y, x == y))",message="dataPolicy items must be unique"
	// +listType=atomic
	// +optional
	DataPolicy []DataPolicyOption `json:"dataPolicy,omitempty"`
	// gatherers is a required field that specifies the configuration of the gatherers.
	// +required
	Gatherers Gatherers `json:"gatherers,omitempty,omitzero"`
	// storage is an optional field that allows user to define persistent storage for gathering jobs to store the Insights data archive.
	// If omitted, the gathering job will use ephemeral storage.
	// +optional
	Storage Storage `json:"storage,omitempty,omitzero"`
}

// Gatherers specifies the configuration of the gatherers
// +kubebuilder:validation:XValidation:rule="has(self.mode) && self.mode == 'Custom' ?  has(self.custom) : !has(self.custom)",message="custom is required when mode is Custom, and forbidden otherwise"
// +union
type Gatherers struct {
	// mode is a required field that specifies the mode for gatherers. Allowed values are All and Custom.
	// When set to All, all gatherers will run and gather data.
	// When set to Custom, the custom configuration from the custom field will be applied.
	// +unionDiscriminator
	// +required
	Mode GatheringMode `json:"mode,omitempty"`
	// custom provides gathering configuration.
	// It is required when mode is Custom, and forbidden otherwise.
	// Custom configuration allows user to disable only a subset of gatherers.
	// Gatherers that are not explicitly disabled in custom configuration will run.
	// +unionMember
	// +optional
	Custom Custom `json:"custom,omitempty,omitzero"`
}

// Custom provides the custom configuration of gatherers
type Custom struct {
	// configs is a required list of gatherers configurations that can be used to enable or disable specific gatherers.
	// It may not exceed 100 items and each gatherer can be present only once.
	// It is possible to disable an entire set of gatherers while allowing a specific function within that set.
	// The particular gatherers IDs can be found at https://github.com/openshift/insights-operator/blob/master/docs/gathered-data.md.
	// Run the following command to get the names of last active gatherers:
	// "oc get insightsoperators.operator.openshift.io cluster -o json | jq '.status.gatherStatus.gatherers[].name'"
	// +kubebuilder:validation:MaxItems=100
	// +kubebuilder:validation:MinItems=1
	// +listType=map
	// +listMapKey=name
	// +required
	Configs []GathererConfig `json:"configs,omitempty"`
}

// GatheringMode defines the valid gathering modes.
// +kubebuilder:validation:Enum=All;Custom
type GatheringMode string

const (
	// Enabled enables all gatherers
	GatheringModeAll GatheringMode = "All"
	// Custom applies the configuration from GatheringConfig.
	GatheringModeCustom GatheringMode = "Custom"
)

// Storage provides persistent storage configuration options for gathering jobs.
// If the type is set to PersistentVolume, then the PersistentVolume must be defined.
// If the type is set to Ephemeral, then the PersistentVolume must not be defined.
// +kubebuilder:validation:XValidation:rule="has(self.type) && self.type == 'PersistentVolume' ?  has(self.persistentVolume) : !has(self.persistentVolume)",message="persistentVolume is required when type is PersistentVolume, and forbidden otherwise"
// +union
type Storage struct {
	// type is a required field that specifies the type of storage that will be used to store the Insights data archive.
	// Valid values are "PersistentVolume" and "Ephemeral".
	// When set to Ephemeral, the Insights data archive is stored in the ephemeral storage of the gathering job.
	// When set to PersistentVolume, the Insights data archive is stored in the PersistentVolume that is
	// defined by the PersistentVolume field.
	// +unionDiscriminator
	// +required
	Type StorageType `json:"type,omitempty"`
	// persistentVolume is an optional field that specifies the PersistentVolume that will be used to store the Insights data archive.
	// The PersistentVolume must be created in the openshift-insights namespace.
	// +unionMember
	// +optional
	PersistentVolume PersistentVolumeConfig `json:"persistentVolume,omitempty,omitzero"`
}

// StorageType declares valid storage types
// +kubebuilder:validation:Enum=PersistentVolume;Ephemeral
type StorageType string

const (
	// StorageTypePersistentVolume storage type
	StorageTypePersistentVolume StorageType = "PersistentVolume"
	// StorageTypeEphemeral storage type
	StorageTypeEphemeral StorageType = "Ephemeral"
)

// PersistentVolumeConfig provides configuration options for PersistentVolume storage.
type PersistentVolumeConfig struct {
	// claim is a required field that specifies the configuration of the PersistentVolumeClaim that will be used to store the Insights data archive.
	// The PersistentVolumeClaim must be created in the openshift-insights namespace.
	// +required
	Claim PersistentVolumeClaimReference `json:"claim,omitempty,omitzero"`
	// mountPath is an optional field specifying the directory where the PVC will be mounted inside the Insights data gathering Pod.
	// When omitted, this means no opinion and the platform is left to choose a reasonable default, which is subject to change over time.
	// The current default mount path is /var/lib/insights-operator
	// The path may not exceed 1024 characters and must not contain a colon.
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=1024
	// +kubebuilder:validation:XValidation:rule="!self.contains(':')",message="mountPath must not contain a colon"
	// +optional
	MountPath string `json:"mountPath,omitempty"`
}

// PersistentVolumeClaimReference is a reference to a PersistentVolumeClaim.
type PersistentVolumeClaimReference struct {
	// name is the name of the PersistentVolumeClaim that will be used to store the Insights data archive.
	// It is a string that follows the DNS1123 subdomain format.
	// It must be at most 253 characters in length, and must consist only of lower case alphanumeric characters, '-' and '.', and must start and end with an alphanumeric character.
	// +kubebuilder:validation:XValidation:rule="!format.dns1123Subdomain().validate(self).hasValue()",message="a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character."
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +required
	Name string `json:"name,omitempty"`
}

// DataPolicyOption declares valid data policy types
// +kubebuilder:validation:Enum=ObfuscateNetworking;WorkloadNames
type DataPolicyOption string

const (
	// IP addresses and cluster domain name are obfuscated
	DataPolicyOptionObfuscateNetworking DataPolicyOption = "ObfuscateNetworking"
	// Data from Deployment Validation Operator are obfuscated
	DataPolicyOptionObfuscateWorkloadNames DataPolicyOption = "WorkloadNames"
)

// GathererConfig allows to configure specific gatherers
type GathererConfig struct {
	// name is the required name of a specific gatherer.
	// It may not exceed 256 characters.
	// The format for a gatherer name is: {gatherer}/{function} where the function is optional.
	// Gatherer consists of a lowercase letters only that may include underscores (_).
	// Function consists of a lowercase letters only that may include underscores (_) and is separated from the gatherer by a forward slash (/).
	// The particular gatherers can be found at https://github.com/openshift/insights-operator/blob/master/docs/gathered-data.md.
	// Run the following command to get the names of last active gatherers:
	// "oc get insightsoperators.operator.openshift.io cluster -o json | jq '.status.gatherStatus.gatherers[].name'"
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	// +kubebuilder:validation:XValidation:rule=`self.matches("^[a-z]+[_a-z]*[a-z]([/a-z][_a-z]*)?[a-z]$")`,message=`gatherer name must be in the format of {gatherer}/{function} where the gatherer and function are lowercase letters only that may include underscores (_) and are separated by a forward slash (/) if the function is provided`
	// +required
	Name string `json:"name,omitempty"`
	// state is a required field that allows you to configure specific gatherer. Valid values are "Enabled" and "Disabled".
	// When set to Enabled the gatherer will run.
	// When set to Disabled the gatherer will not run.
	// +required
	State GathererState `json:"state,omitempty"`
}

// GathererState declares valid gatherer state types.
// +kubebuilder:validation:Enum=Enabled;Disabled
type GathererState string

const (
	// GathererStateEnabled gatherer state, which means that the gatherer will run.
	GathererStateEnabled GathererState = "Enabled"
	// GathererStateDisabled gatherer state, which means that the gatherer will not run.
	GathererStateDisabled GathererState = "Disabled"
)

// DataGatherStatus contains information relating to the DataGather state.
// +kubebuilder:validation:XValidation:rule="(!has(oldSelf.insightsRequestID) || has(self.insightsRequestID))",message="cannot remove insightsRequestID attribute from status"
// +kubebuilder:validation:XValidation:rule="(!has(oldSelf.startTime) || has(self.startTime))",message="cannot remove startTime attribute from status"
// +kubebuilder:validation:XValidation:rule="(!has(oldSelf.finishTime) || has(self.finishTime))",message="cannot remove finishTime attribute from status"
// +kubebuilder:validation:MinProperties=1
type DataGatherStatus struct {
	// conditions is an optional field that provides details on the status of the gatherer job.
	// It may not exceed 100 items and must not contain duplicates.
	//
	// The current condition types are DataUploaded, DataRecorded, DataProcessed, RemoteConfigurationNotAvailable, RemoteConfigurationInvalid
	//
	// The DataUploaded condition is used to represent whether or not the archive was successfully uploaded for further processing.
	// When it has a status of True and a reason of Succeeded, the archive was successfully uploaded.
	// When it has a status of Unknown and a reason of NoUploadYet, the upload has not occurred, or there was no data to upload.
	// When it has a status of False and a reason Failed, the upload failed. The accompanying message will include the specific error encountered.
	//
	// The DataRecorded condition is used to represent whether or not the archive was successfully recorded.
	// When it has a status of True and a reason of Succeeded, the archive was recorded successfully.
	// When it has a status of Unknown and a reason of NoDataGatheringYet, the data gathering process has not started yet.
	// When it has a status of False and a reason of RecordingFailed, the recording failed and a message will include the specific error encountered.
	//
	// The DataProcessed condition is used to represent whether or not the archive was processed by the processing service.
	// When it has a status of True and a reason of Processed, the data was processed successfully.
	// When it has a status of Unknown and a reason of NothingToProcessYet, there is no data to process at the moment.
	// When it has a status of False and a reason of Failure, processing failed and a message will include the specific error encountered.
	//
	// The RemoteConfigurationAvailable condition is used to represent whether the remote configuration is available.
	// When it has a status of Unknown and a reason of Unknown or RemoteConfigNotRequestedYet, the state of the remote configuration is unknown—typically at startup.
	// When it has a status of True and a reason of Succeeded, the configuration is available.
	// When it has a status of False and a reason of NoToken, the configuration was disabled by removing the cloud.openshift.com field from the pull secret.
	// When it has a status of False and a reason of DisabledByConfiguration, the configuration was disabled in insightsdatagather.config.openshift.io.
	//
	// The RemoteConfigurationValid condition is used to represent whether the remote configuration is valid.
	// When it has a status of Unknown and a reason of Unknown or NoValidationYet, the validity of the remote configuration is unknown—typically at startup.
	// When it has a status of True and a reason of Succeeded, the configuration is valid.
	// When it has a status of False and a reason of Invalid, the configuration is invalid.
	//
	// The Progressing condition is used to represent the phase of gathering
	// When it has a status of False and the reason is DataGatherPending, the gathering has not started yet.
	// When it has a status of True and reason is Gathering, the gathering is running.
	// When it has a status of False and reason is GatheringSucceeded, the gathering successfully finished.
	// When it has a status of False and reason is GatheringFailed, the gathering failed.
	//
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=100
	// +kubebuilder:validation:MinItems=1
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// gatherers is a list of active gatherers (and their statuses) in the last gathering.
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MaxItems=100
	// +kubebuilder:validation:MinItems=1
	// +optional
	Gatherers []GathererStatus `json:"gatherers,omitempty"`
	// startTime is the time when Insights data gathering started.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="startTime is immutable once set"
	// +optional
	StartTime metav1.Time `json:"startTime,omitempty,omitzero"`
	// finishTime is the time when Insights data gathering finished.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="finishTime is immutable once set"
	// +optional
	FinishTime metav1.Time `json:"finishTime,omitempty,omitzero"`
	// relatedObjects is an optional list of resources which are useful when debugging or inspecting the data gathering Pod
	// It may not exceed 100 items and must not contain duplicates.
	// +listType=map
	// +listMapKey=name
	// +listMapKey=namespace
	// +kubebuilder:validation:MaxItems=100
	// +kubebuilder:validation:MinItems=1
	// +optional
	RelatedObjects []ObjectReference `json:"relatedObjects,omitempty"`
	// insightsRequestID is an optional Insights request ID to track the status of the Insights analysis (in console.redhat.com processing pipeline) for the corresponding Insights data archive.
	// It may not exceed 256 characters and is immutable once set.
	// +kubebuilder:validation:XValidation:rule="self == oldSelf",message="insightsRequestID is immutable once set"
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=256
	// +optional
	InsightsRequestID string `json:"insightsRequestID,omitempty"`
	// insightsReport provides general Insights analysis results.
	// When omitted, this means no data gathering has taken place yet or the
	// corresponding Insights analysis (identified by "insightsRequestID") is not available.
	// +optional
	InsightsReport InsightsReport `json:"insightsReport,omitzero"`
}

// GathererStatus represents information about a particular
// data gatherer.
type GathererStatus struct {
	// conditions provide details on the status of each gatherer.
	//
	// The current condition type is DataGathered
	//
	// The DataGathered condition is used to represent whether or not the data was gathered by a gatherer specified by name.
	// When it has a status of True and a reason of GatheredOK, the data has been successfully gathered as expected.
	// When it has a status of False and a reason of NoData, no data was gathered—for example, when the resource is not present in the cluster.
	// When it has a status of False and a reason of GatherError, an error occurred and no data was gathered.
	// When it has a status of False and a reason of GatherPanic, a panic occurred during gathering and no data was collected.
	// When it has a status of False and a reason of GatherWithErrorReason, data was partially gathered or gathered with an error message.
	//
	// +listType=map
	// +listMapKey=type
	// +kubebuilder:validation:MaxItems=100
	// +kubebuilder:validation:MinItems=1
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
	// name is the required name of the gatherer.
	// It must contain at least 5 characters and may not exceed 256 characters.
	// +kubebuilder:validation:MaxLength=256
	// +kubebuilder:validation:MinLength=5
	// +required
	Name string `json:"name,omitempty"`
	// lastGatherSeconds is required field that represents the time spent gathering in seconds
	// +kubebuilder:validation:Minimum=0
	// +required
	LastGatherSeconds *int32 `json:"lastGatherSeconds,omitempty"`
}

// InsightsReport provides Insights health check report based on the most
// recently sent Insights data.
type InsightsReport struct {
	// downloadedTime is a required field that specifies when the Insights report was last downloaded.
	// +required
	DownloadedTime metav1.Time `json:"downloadedTime,omitempty"`
	// healthChecks is an optional field that provides basic information about active Insights
	// recommendations, which serve as proactive notifications for potential issues in the cluster.
	// When omitted, it means that there are no active recommendations in the cluster.
	// +listType=map
	// +listMapKey=advisorURI
	// +listMapKey=totalRisk
	// +listMapKey=description
	// +kubebuilder:validation:MaxItems=100
	// +kubebuilder:validation:MinItems=1
	// +optional
	HealthChecks []HealthCheck `json:"healthChecks,omitempty"`
	// uri is a required field that provides the URL link from which the report was downloaded.
	// The link must be a valid HTTPS URL and the maximum length is 2048 characters.
	// +kubebuilder:validation:XValidation:rule=`isURL(self) && url(self).getScheme() == "https"`,message=`URI must be a valid HTTPS URL (e.g., https://example.com)`
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	// +required
	URI string `json:"uri,omitempty,omitzero"`
}

// HealthCheck represents an Insights health check attributes.
type HealthCheck struct {
	// description is required field that provides basic description of the healthcheck.
	// It must contain at least 10 characters and may not exceed 2048 characters.
	// +kubebuilder:validation:MinLength=10
	// +kubebuilder:validation:MaxLength=2048
	// +required
	Description string `json:"description,omitempty"`
	// totalRisk is the required field of the healthcheck.
	// It is indicator of the total risk posed by the detected issue; combination of impact and likelihood.
	// Allowed values are Low, Moderate, Important and Critical.
	// The value represents the severity of the issue.
	// +required
	TotalRisk TotalRisk `json:"totalRisk,omitempty"`
	// advisorURI is required field that provides the URL link to the Insights Advisor.
	// The link must be a valid HTTPS URL and the maximum length is 2048 characters.
	// +kubebuilder:validation:XValidation:rule=`isURL(self) && url(self).getScheme() == "https"`,message=`advisorURI must be a valid HTTPS URL (e.g., https://example.com)`
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=2048
	// +required
	AdvisorURI string `json:"advisorURI,omitempty"`
}

// TotalRisk defines the valid totalRisk values.
// +kubebuilder:validation:Enum=Low;Moderate;Important;Critical
type TotalRisk string

const (
	TotalRiskLow       TotalRisk = "Low"
	TotalRiskModerate  TotalRisk = "Moderate"
	TotalRiskImportant TotalRisk = "Important"
	TotalRiskCritical  TotalRisk = "Critical"
)

// ObjectReference contains enough information to let you inspect or modify the referred object.
type ObjectReference struct {
	// group is required field that specifies the API Group of the Resource.
	// Enter empty string for the core group.
	// This value is empty or it should follow the DNS1123 subdomain format.
	// It must be at most 253 characters in length, and must consist only of lower case alphanumeric characters, '-' and '.', and must start with an alphabetic character and end with an alphanumeric character.
	// Example: "", "apps", "build.openshift.io", etc.
	// +kubebuilder:validation:XValidation:rule="self.size() == 0 || !format.dns1123Subdomain().validate(self).hasValue()",message="a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start with an alphabetic character and end with an alphanumeric character."
	// +kubebuilder:validation:MinLength=0
	// +kubebuilder:validation:MaxLength=253
	// +required
	Group *string `json:"group,omitempty"`
	// resource is required field of the type that is being referenced and follows the DNS1035 format.
	// It is normally the plural form of the resource kind in lowercase.
	// It must be at most 63 characters in length, and must must consist of only lowercase alphanumeric characters and hyphens, and must start with an alphabetic character and end with an alphanumeric character.
	// Example: "deployments", "deploymentconfigs", "pods", etc.
	// +kubebuilder:validation:XValidation:rule=`!format.dns1035Label().validate(self).hasValue()`,message="the value must consist of only lowercase alphanumeric characters and hyphens, and must start with an alphabetic character and end with an alphanumeric character."
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +required
	Resource string `json:"resource,omitempty"`
	// name is required field that specifies the referent that follows the DNS1123 subdomain format.
	// It must be at most 253 characters in length, and must consist only of lower case alphanumeric characters, '-' and '.', and must start with an alphabetic character and end with an alphanumeric character..
	// +kubebuilder:validation:XValidation:rule="!format.dns1123Subdomain().validate(self).hasValue()",message="a lowercase RFC 1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start with an alphabetic character and end with an alphanumeric character."
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=253
	// +required
	Name string `json:"name,omitempty"`
	// namespace if required field of the referent that follows the DNS1123 labels format.
	// It must be at most 63 characters in length, and must must consist of only lowercase alphanumeric characters and hyphens, and must start with an alphabetic character and end with an alphanumeric character.
	// +kubebuilder:validation:XValidation:rule=`!format.dns1123Label().validate(self).hasValue()`,message="the value must consist of only lowercase alphanumeric characters and hyphens, and must start with an alphabetic character and end with an alphanumeric character."
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:MaxLength=63
	// +required
	Namespace string `json:"namespace,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DataGatherList is a collection of items
//
// Compatibility level 1: Stable within a major release for a minimum of 12 months or 3 minor releases (whichever is longer).
// +openshift:compatibility-gen:level=1
type DataGatherList struct {
	metav1.TypeMeta `json:",inline"`
	// metadata is the standard list's metadata.
	// More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#metadata
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`
	// items contains a list of DataGather resources.
	// +listType=atomic
	// +kubebuilder:validation:MaxItems=100
	// +kubebuilder:validation:MinItems=1
	// +optional
	Items []DataGather `json:"items,omitempty"`
}
