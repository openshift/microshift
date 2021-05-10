// Code generated for package assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// assets/crd/0000_00_cluster-version-operator_01_clusteroperator.crd.yaml
// assets/crd/0000_00_cluster-version-operator_01_clusterversion.crd.yaml
// assets/crd/0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml
// assets/crd/0000_03_config-operator_01_operatorhub.crd.yaml
// assets/crd/0000_03_config-operator_01_proxy.crd.yaml
// assets/crd/0000_03_quota-openshift_01_clusterresourcequota.crd.yaml
// assets/crd/0000_03_security-openshift_01_scc.crd.yaml
// assets/crd/0000_10_config-operator_01_apiserver.crd.yaml
// assets/crd/0000_10_config-operator_01_authentication.crd.yaml
// assets/crd/0000_10_config-operator_01_build.crd.yaml
// assets/crd/0000_10_config-operator_01_console.crd.yaml
// assets/crd/0000_10_config-operator_01_dns.crd.yaml
// assets/crd/0000_10_config-operator_01_featuregate.crd.yaml
// assets/crd/0000_10_config-operator_01_image.crd.yaml
// assets/crd/0000_10_config-operator_01_imagecontentsourcepolicy.crd.yaml
// assets/crd/0000_10_config-operator_01_infrastructure.crd.yaml
// assets/crd/0000_10_config-operator_01_ingress.crd.yaml
// assets/crd/0000_10_config-operator_01_network.crd.yaml
// assets/crd/0000_10_config-operator_01_oauth.crd.yaml
// assets/crd/0000_10_config-operator_01_project.crd.yaml
// assets/crd/0000_11_imageregistry-configs.crd.yaml
// assets/crd/0000_50_service-ca-operator_02_crd.yaml
// assets/crd/0000_70_dns-operator_00-custom-resource-definition.yaml
// assets/crd/cluster-ingress-00-custom-resource-definition.yaml
// assets/crd/cluster-network-01-crd.yml
package assets

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type asset struct {
	bytes []byte
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
}

// Name return file name
func (fi bindataFileInfo) Name() string {
	return fi.name
}

// Size return file size
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}

// Mode return file mode
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}

// Mode return file modify time
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}

// IsDir return file whether a directory
func (fi bindataFileInfo) IsDir() bool {
	return fi.mode&os.ModeDir != 0
}

// Sys return file is sys mode
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _assetsCrd0000_00_clusterVersionOperator_01_clusteroperatorCrdYaml = []byte(`kind: CustomResourceDefinition
apiVersion: apiextensions.k8s.io/v1beta1
metadata:
  name: clusteroperators.config.openshift.io
spec:
  additionalPrinterColumns:
  - JSONPath: .status.versions[?(@.name=="operator")].version
    description: The version the operator is at.
    name: Version
    type: string
  - JSONPath: .status.conditions[?(@.type=="Available")].status
    description: Whether the operator is running and stable.
    name: Available
    type: string
  - JSONPath: .status.conditions[?(@.type=="Progressing")].status
    description: Whether the operator is processing changes.
    name: Progressing
    type: string
  - JSONPath: .status.conditions[?(@.type=="Degraded")].status
    description: Whether the operator is degraded.
    name: Degraded
    type: string
  - JSONPath: .status.conditions[?(@.type=="Available")].lastTransitionTime
    description: The time the operator's Available status last changed.
    name: Since
    type: date
  group: config.openshift.io
  names:
    kind: ClusterOperator
    listKind: ClusterOperatorList
    plural: clusteroperators
    singular: clusteroperator
    shortNames:
    - co
  preserveUnknownFields: false
  scope: Cluster
  subresources:
    status: {}
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
  validation:
    openAPIV3Schema:
      description: ClusterOperator is the Custom Resource object which holds the current
        state of an operator. This object is used by operators to convey their state
        to the rest of the cluster.
      type: object
      required:
      - spec
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: spec holds configuration that could apply to any operator.
          type: object
        status:
          description: status holds the information about the state of an operator.  It
            is consistent with status information across the Kubernetes ecosystem.
          type: object
          properties:
            conditions:
              description: conditions describes the state of the operator's managed
                and monitored components.
              type: array
              items:
                description: ClusterOperatorStatusCondition represents the state of
                  the operator's managed and monitored components.
                type: object
                required:
                - lastTransitionTime
                - status
                - type
                properties:
                  lastTransitionTime:
                    description: lastTransitionTime is the time of the last update
                      to the current status property.
                    type: string
                    format: date-time
                  message:
                    description: message provides additional information about the
                      current condition. This is only to be consumed by humans.
                    type: string
                  reason:
                    description: reason is the CamelCase reason for the condition's
                      current status.
                    type: string
                  status:
                    description: status of the condition, one of True, False, Unknown.
                    type: string
                  type:
                    description: type specifies the aspect reported by this condition.
                    type: string
            extension:
              description: extension contains any additional status information specific
                to the operator which owns this status object.
              type: object
              nullable: true
              x-kubernetes-preserve-unknown-fields: true
            relatedObjects:
              description: 'relatedObjects is a list of objects that are "interesting"
                or related to this operator.  Common uses are: 1. the detailed resource
                driving the operator 2. operator namespaces 3. operand namespaces'
              type: array
              items:
                description: ObjectReference contains enough information to let you
                  inspect or modify the referred object.
                type: object
                required:
                - group
                - name
                - resource
                properties:
                  group:
                    description: group of the referent.
                    type: string
                  name:
                    description: name of the referent.
                    type: string
                  namespace:
                    description: namespace of the referent.
                    type: string
                  resource:
                    description: resource of the referent.
                    type: string
            versions:
              description: versions is a slice of operator and operand version tuples.  Operators
                which manage multiple operands will have multiple operand entries
                in the array.  Available operators must report the version of the
                operator itself with the name "operator". An operator reports a new
                "operator" version when it has rolled out the new version to all of
                its operands.
              type: array
              items:
                type: object
                required:
                - name
                - version
                properties:
                  name:
                    description: name is the name of the particular operand this version
                      is for.  It usually matches container images, not operators.
                    type: string
                  version:
                    description: version indicates which version of a particular operand
                      is currently being managed.  It must always match the Available
                      operand.  If 1.0.0 is Available, then this must indicate 1.0.0
                      even if the operator is trying to rollout 1.1.0
                    type: string
  versions:
  - name: v1
    served: true
    storage: true
`)

func assetsCrd0000_00_clusterVersionOperator_01_clusteroperatorCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_00_clusterVersionOperator_01_clusteroperatorCrdYaml, nil
}

func assetsCrd0000_00_clusterVersionOperator_01_clusteroperatorCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_00_clusterVersionOperator_01_clusteroperatorCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_00_cluster-version-operator_01_clusteroperator.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_00_clusterVersionOperator_01_clusterversionCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: clusterversions.config.openshift.io
spec:
  group: config.openshift.io
  versions:
  - name: v1
    served: true
    storage: true
  scope: Cluster
  subresources:
    status: {}
  names:
    plural: clusterversions
    singular: clusterversion
    kind: ClusterVersion
  preserveUnknownFields: false
  additionalPrinterColumns:
  - name: Version
    type: string
    JSONPath: .status.history[?(@.state=="Completed")].version
  - name: Available
    type: string
    JSONPath: .status.conditions[?(@.type=="Available")].status
  - name: Progressing
    type: string
    JSONPath: .status.conditions[?(@.type=="Progressing")].status
  - name: Since
    type: date
    JSONPath: .status.conditions[?(@.type=="Progressing")].lastTransitionTime
  - name: Status
    type: string
    JSONPath: .status.conditions[?(@.type=="Progressing")].message
  validation:
    openAPIV3Schema:
      description: ClusterVersion is the configuration for the ClusterVersionOperator.
        This is where parameters related to automatic updates can be set.
      type: object
      required:
      - spec
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: spec is the desired state of the cluster version - the operator
            will work to ensure that the desired version is applied to the cluster.
          type: object
          required:
          - clusterID
          properties:
            channel:
              description: channel is an identifier for explicitly requesting that
                a non-default set of updates be applied to this cluster. The default
                channel will be contain stable updates that are appropriate for production
                clusters.
              type: string
            clusterID:
              description: clusterID uniquely identifies this cluster. This is expected
                to be an RFC4122 UUID value (xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
                in hexadecimal values). This is a required field.
              type: string
            desiredUpdate:
              description: "desiredUpdate is an optional field that indicates the
                desired value of the cluster version. Setting this value will trigger
                an upgrade (if the current version does not match the desired version).
                The set of recommended update values is listed as part of available
                updates in status, and setting values outside that range may cause
                the upgrade to fail. You may specify the version field without setting
                image if an update exists with that version in the availableUpdates
                or history. \n If an upgrade fails the operator will halt and report
                status about the failing component. Setting the desired update value
                back to the previous version will cause a rollback to be attempted.
                Not all rollbacks will succeed."
              type: object
              properties:
                force:
                  description: "force allows an administrator to update to an image
                    that has failed verification, does not appear in the availableUpdates
                    list, or otherwise would be blocked by normal protections on update.
                    This option should only be used when the authenticity of the provided
                    image has been verified out of band because the provided image
                    will run with full administrative access to the cluster. Do not
                    use this flag with images that comes from unknown or potentially
                    malicious sources. \n This flag does not override other forms
                    of consistency checking that are required before a new update
                    is deployed."
                  type: boolean
                image:
                  description: image is a container image location that contains the
                    update. When this field is part of spec, image is optional if
                    version is specified and the availableUpdates field contains a
                    matching version.
                  type: string
                version:
                  description: version is a semantic versioning identifying the update
                    version. When this field is part of spec, version is optional
                    if image is specified.
                  type: string
            overrides:
              description: overrides is list of overides for components that are managed
                by cluster version operator. Marking a component unmanaged will prevent
                the operator from creating or updating the object.
              type: array
              items:
                description: ComponentOverride allows overriding cluster version operator's
                  behavior for a component.
                type: object
                required:
                - group
                - kind
                - name
                - namespace
                - unmanaged
                properties:
                  group:
                    description: group identifies the API group that the kind is in.
                    type: string
                  kind:
                    description: kind indentifies which object to override.
                    type: string
                  name:
                    description: name is the component's name.
                    type: string
                  namespace:
                    description: namespace is the component's namespace. If the resource
                      is cluster scoped, the namespace should be empty.
                    type: string
                  unmanaged:
                    description: 'unmanaged controls if cluster version operator should
                      stop managing the resources in this cluster. Default: false'
                    type: boolean
            upstream:
              description: upstream may be used to specify the preferred update server.
                By default it will use the appropriate update server for the cluster
                and region.
              type: string
        status:
          description: status contains information about the available updates and
            any in-progress updates.
          type: object
          required:
          - availableUpdates
          - desired
          - observedGeneration
          - versionHash
          properties:
            availableUpdates:
              description: availableUpdates contains the list of updates that are
                appropriate for this cluster. This list may be empty if no updates
                are recommended, if the update service is unavailable, or if an invalid
                channel has been specified.
              type: array
              items:
                description: Update represents a release of the ClusterVersionOperator,
                  referenced by the Image member.
                type: object
                properties:
                  force:
                    description: "force allows an administrator to update to an image
                      that has failed verification, does not appear in the availableUpdates
                      list, or otherwise would be blocked by normal protections on
                      update. This option should only be used when the authenticity
                      of the provided image has been verified out of band because
                      the provided image will run with full administrative access
                      to the cluster. Do not use this flag with images that comes
                      from unknown or potentially malicious sources. \n This flag
                      does not override other forms of consistency checking that are
                      required before a new update is deployed."
                    type: boolean
                  image:
                    description: image is a container image location that contains
                      the update. When this field is part of spec, image is optional
                      if version is specified and the availableUpdates field contains
                      a matching version.
                    type: string
                  version:
                    description: version is a semantic versioning identifying the
                      update version. When this field is part of spec, version is
                      optional if image is specified.
                    type: string
              nullable: true
            conditions:
              description: conditions provides information about the cluster version.
                The condition "Available" is set to true if the desiredUpdate has
                been reached. The condition "Progressing" is set to true if an update
                is being applied. The condition "Degraded" is set to true if an update
                is currently blocked by a temporary or permanent error. Conditions
                are only valid for the current desiredUpdate when metadata.generation
                is equal to status.generation.
              type: array
              items:
                description: ClusterOperatorStatusCondition represents the state of
                  the operator's managed and monitored components.
                type: object
                required:
                - lastTransitionTime
                - status
                - type
                properties:
                  lastTransitionTime:
                    description: lastTransitionTime is the time of the last update
                      to the current status property.
                    type: string
                    format: date-time
                  message:
                    description: message provides additional information about the
                      current condition. This is only to be consumed by humans.
                    type: string
                  reason:
                    description: reason is the CamelCase reason for the condition's
                      current status.
                    type: string
                  status:
                    description: status of the condition, one of True, False, Unknown.
                    type: string
                  type:
                    description: type specifies the aspect reported by this condition.
                    type: string
            desired:
              description: desired is the version that the cluster is reconciling
                towards. If the cluster is not yet fully initialized desired will
                be set with the information available, which may be an image or a
                tag.
              type: object
              properties:
                force:
                  description: "force allows an administrator to update to an image
                    that has failed verification, does not appear in the availableUpdates
                    list, or otherwise would be blocked by normal protections on update.
                    This option should only be used when the authenticity of the provided
                    image has been verified out of band because the provided image
                    will run with full administrative access to the cluster. Do not
                    use this flag with images that comes from unknown or potentially
                    malicious sources. \n This flag does not override other forms
                    of consistency checking that are required before a new update
                    is deployed."
                  type: boolean
                image:
                  description: image is a container image location that contains the
                    update. When this field is part of spec, image is optional if
                    version is specified and the availableUpdates field contains a
                    matching version.
                  type: string
                version:
                  description: version is a semantic versioning identifying the update
                    version. When this field is part of spec, version is optional
                    if image is specified.
                  type: string
            history:
              description: history contains a list of the most recent versions applied
                to the cluster. This value may be empty during cluster startup, and
                then will be updated when a new update is being applied. The newest
                update is first in the list and it is ordered by recency. Updates
                in the history have state Completed if the rollout completed - if
                an update was failing or halfway applied the state will be Partial.
                Only a limited amount of update history is preserved.
              type: array
              items:
                description: UpdateHistory is a single attempted update to the cluster.
                type: object
                required:
                - completionTime
                - image
                - startedTime
                - state
                - verified
                properties:
                  completionTime:
                    description: completionTime, if set, is when the update was fully
                      applied. The update that is currently being applied will have
                      a null completion time. Completion time will always be set for
                      entries that are not the current update (usually to the started
                      time of the next update).
                    type: string
                    format: date-time
                    nullable: true
                  image:
                    description: image is a container image location that contains
                      the update. This value is always populated.
                    type: string
                  startedTime:
                    description: startedTime is the time at which the update was started.
                    type: string
                    format: date-time
                  state:
                    description: state reflects whether the update was fully applied.
                      The Partial state indicates the update is not fully applied,
                      while the Completed state indicates the update was successfully
                      rolled out at least once (all parts of the update successfully
                      applied).
                    type: string
                  verified:
                    description: verified indicates whether the provided update was
                      properly verified before it was installed. If this is false
                      the cluster may not be trusted.
                    type: boolean
                  version:
                    description: version is a semantic versioning identifying the
                      update version. If the requested image does not define a version,
                      or if a failure occurs retrieving the image, this value may
                      be empty.
                    type: string
            observedGeneration:
              description: observedGeneration reports which version of the spec is
                being synced. If this value is not equal to metadata.generation, then
                the desired and conditions fields may represent a previous version.
              type: integer
              format: int64
            versionHash:
              description: versionHash is a fingerprint of the content that the cluster
                will be updated with. It is used by the operator to avoid unnecessary
                work and is for internal use only.
              type: string
  versions:
  - name: v1
    served: true
    storage: true
`)

func assetsCrd0000_00_clusterVersionOperator_01_clusterversionCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_00_clusterVersionOperator_01_clusterversionCrdYaml, nil
}

func assetsCrd0000_00_clusterVersionOperator_01_clusterversionCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_00_clusterVersionOperator_01_clusterversionCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_00_cluster-version-operator_01_clusterversion.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: rolebindingrestrictions.authorization.openshift.io
spec:
  group: authorization.openshift.io
  names:
    kind: RoleBindingRestriction
    listKind: RoleBindingRestrictionList
    plural: rolebindingrestrictions
    singular: rolebindingrestriction
  scope: Namespaced
  versions:
  - name: v1
    served: true
    storage: true
  validation:
    openAPIV3Schema:
      description: RoleBindingRestriction is an object that can be matched against
        a subject (user, group, or service account) to determine whether rolebindings
        on that subject are allowed in the namespace to which the RoleBindingRestriction
        belongs.  If any one of those RoleBindingRestriction objects matches a subject,
        rolebindings on that subject in the namespace are allowed.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object's metadata.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: Spec defines the matcher.
          properties:
            grouprestriction:
              description: GroupRestriction matches against group subjects.
              nullable: true
              properties:
                groups:
                  description: Groups is a list of groups used to match against an
                    individual user's groups. If the user is a member of one of the
                    whitelisted groups, the user is allowed to be bound to a role.
                  items:
                    type: string
                  nullable: true
                  type: array
                labels:
                  description: Selectors specifies a list of label selectors over
                    group labels.
                  items:
                    description: A label selector is a label query over a set of resources.
                      The result of matchLabels and matchExpressions are ANDed. An
                      empty label selector matches all objects. A null label selector
                      matches no objects.
                    properties:
                      matchExpressions:
                        description: matchExpressions is a list of label selector
                          requirements. The requirements are ANDed.
                        items:
                          description: A label selector requirement is a selector
                            that contains values, a key, and an operator that relates
                            the key and values.
                          properties:
                            key:
                              description: key is the label key that the selector
                                applies to.
                              type: string
                            operator:
                              description: operator represents a key's relationship
                                to a set of values. Valid operators are In, NotIn,
                                Exists and DoesNotExist.
                              type: string
                            values:
                              description: values is an array of string values. If
                                the operator is In or NotIn, the values array must
                                be non-empty. If the operator is Exists or DoesNotExist,
                                the values array must be empty. This array is replaced
                                during a strategic merge patch.
                              items:
                                type: string
                              type: array
                          required:
                          - key
                          - operator
                          type: object
                        type: array
                      matchLabels:
                        additionalProperties:
                          type: string
                        description: matchLabels is a map of {key,value} pairs. A
                          single {key,value} in the matchLabels map is equivalent
                          to an element of matchExpressions, whose key field is "key",
                          the operator is "In", and the values array contains only
                          "value". The requirements are ANDed.
                        type: object
                    type: object
                  nullable: true
                  type: array
              type: object
            serviceaccountrestriction:
              description: ServiceAccountRestriction matches against service-account
                subjects.
              nullable: true
              properties:
                namespaces:
                  description: Namespaces specifies a list of literal namespace names.
                  items:
                    type: string
                  type: array
                serviceaccounts:
                  description: ServiceAccounts specifies a list of literal service-account
                    names.
                  items:
                    description: ServiceAccountReference specifies a service account
                      and namespace by their names.
                    properties:
                      name:
                        description: Name is the name of the service account.
                        type: string
                      namespace:
                        description: Namespace is the namespace of the service account.  Service
                          accounts from inside the whitelisted namespaces are allowed
                          to be bound to roles.  If Namespace is empty, then the namespace
                          of the RoleBindingRestriction in which the ServiceAccountReference
                          is embedded is used.
                        type: string
                    type: object
                  type: array
              type: object
            userrestriction:
              description: UserRestriction matches against user subjects.
              nullable: true
              properties:
                groups:
                  description: Groups specifies a list of literal group names.
                  items:
                    type: string
                  nullable: true
                  type: array
                labels:
                  description: Selectors specifies a list of label selectors over
                    user labels.
                  items:
                    description: A label selector is a label query over a set of resources.
                      The result of matchLabels and matchExpressions are ANDed. An
                      empty label selector matches all objects. A null label selector
                      matches no objects.
                    properties:
                      matchExpressions:
                        description: matchExpressions is a list of label selector
                          requirements. The requirements are ANDed.
                        items:
                          description: A label selector requirement is a selector
                            that contains values, a key, and an operator that relates
                            the key and values.
                          properties:
                            key:
                              description: key is the label key that the selector
                                applies to.
                              type: string
                            operator:
                              description: operator represents a key's relationship
                                to a set of values. Valid operators are In, NotIn,
                                Exists and DoesNotExist.
                              type: string
                            values:
                              description: values is an array of string values. If
                                the operator is In or NotIn, the values array must
                                be non-empty. If the operator is Exists or DoesNotExist,
                                the values array must be empty. This array is replaced
                                during a strategic merge patch.
                              items:
                                type: string
                              type: array
                          required:
                          - key
                          - operator
                          type: object
                        type: array
                      matchLabels:
                        additionalProperties:
                          type: string
                        description: matchLabels is a map of {key,value} pairs. A
                          single {key,value} in the matchLabels map is equivalent
                          to an element of matchExpressions, whose key field is "key",
                          the operator is "In", and the values array contains only
                          "value". The requirements are ANDed.
                        type: object
                    type: object
                  nullable: true
                  type: array
                users:
                  description: Users specifies a list of literal user names.
                  items:
                    type: string
                  type: array
              type: object
          type: object
      type: object
`)

func assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml, nil
}

func assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_03_configOperator_01_operatorhubCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: operatorhubs.config.openshift.io
spec:
  group: config.openshift.io
  names:
    kind: OperatorHub
    listKind: OperatorHubList
    plural: operatorhubs
    singular: operatorhub
  scope: Cluster
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: OperatorHub is the Schema for the operatorhubs API. It can be used
        to change the state of the default hub sources for OperatorHub on the cluster
        from enabled to disabled and vice versa.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: ObjectMeta is metadata that all persisted resources must have,
            which includes all objects users must create.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: OperatorHubSpec defines the desired state of OperatorHub
          properties:
            disableAllDefaultSources:
              description: disableAllDefaultSources allows you to disable all the
                default hub sources. If this is true, a specific entry in sources
                can be used to enable a default source. If this is false, a specific
                entry in sources can be used to disable or enable a default source.
              type: boolean
            sources:
              description: sources is the list of default hub sources and their configuration.
                If the list is empty, it implies that the default hub sources are
                enabled on the cluster unless disableAllDefaultSources is true. If
                disableAllDefaultSources is true and sources is not empty, the configuration
                present in sources will take precedence. The list of default hub sources
                and their current state will always be reflected in the status block.
              items:
                description: HubSource is used to specify the hub source and its configuration
                properties:
                  disabled:
                    description: disabled is used to disable a default hub source
                      on cluster
                    type: boolean
                  name:
                    description: name is the name of one of the default hub sources
                    maxLength: 253
                    minLength: 1
                    type: string
                type: object
              type: array
          type: object
        status:
          description: OperatorHubStatus defines the observed state of OperatorHub.
            The current state of the default hub sources will always be reflected
            here.
          properties:
            sources:
              description: sources encapsulates the result of applying the configuration
                for each hub source
              items:
                description: HubSourceStatus is used to reflect the current state
                  of applying the configuration to a default source
                properties:
                  disabled:
                    description: disabled is used to disable a default hub source
                      on cluster
                    type: boolean
                  message:
                    description: message provides more information regarding failures
                    type: string
                  name:
                    description: name is the name of one of the default hub sources
                    maxLength: 253
                    minLength: 1
                    type: string
                  status:
                    description: status indicates success or failure in applying the
                      configuration
                    type: string
                type: object
              type: array
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
`)

func assetsCrd0000_03_configOperator_01_operatorhubCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_03_configOperator_01_operatorhubCrdYaml, nil
}

func assetsCrd0000_03_configOperator_01_operatorhubCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_03_configOperator_01_operatorhubCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_03_config-operator_01_operatorhub.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_03_configOperator_01_proxyCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: proxies.config.openshift.io
spec:
  group: config.openshift.io
  scope: Cluster
  versions:
  - name: v1
    served: true
    storage: true
  names:
    kind: Proxy
    listKind: ProxyList
    plural: proxies
    singular: proxy
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: Proxy holds cluster-wide information on how to configure default
        proxies for the cluster. The canonical name is ` + "`" + `cluster` + "`" + `
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: ObjectMeta is metadata that all persisted resources must have,
            which includes all objects users must create.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: Spec holds user-settable values for the proxy configuration
          properties:
            httpProxy:
              description: httpProxy is the URL of the proxy for HTTP requests.  Empty
                means unset and will not result in an env var.
              type: string
            httpsProxy:
              description: httpsProxy is the URL of the proxy for HTTPS requests.  Empty
                means unset and will not result in an env var.
              type: string
            noProxy:
              description: noProxy is a comma-separated list of hostnames and/or CIDRs
                for which the proxy should not be used. Empty means unset and will
                not result in an env var.
              type: string
            readinessEndpoints:
              description: readinessEndpoints is a list of endpoints used to verify
                readiness of the proxy.
              items:
                type: string
              type: array
            trustedCA:
              description: "trustedCA is a reference to a ConfigMap containing a CA
                certificate bundle used for client egress HTTPS connections. The certificate
                bundle must be from the CA that signed the proxy's certificate and
                be signed for everything. The trustedCA field should only be consumed
                by a proxy validator. The validator is responsible for reading the
                certificate bundle from required key \"ca-bundle.crt\" and copying
                it to a ConfigMap named \"trusted-ca-bundle\" in the \"openshift-config-managed\"
                namespace. The namespace for the ConfigMap referenced by trustedCA
                is \"openshift-config\". Here is an example ConfigMap (in yaml): \n
                apiVersion: v1 kind: ConfigMap metadata:  name: user-ca-bundle  namespace:
                openshift-config  data:    ca-bundle.crt: |      -----BEGIN CERTIFICATE-----
                \     Custom CA certificate bundle.      -----END CERTIFICATE-----"
              properties:
                name:
                  description: name is the metadata.name of the referenced config
                    map
                  type: string
              required:
              - name
              type: object
          type: object
        status:
          description: status holds observed values from the cluster. They may not
            be overridden.
          properties:
            httpProxy:
              description: httpProxy is the URL of the proxy for HTTP requests.
              type: string
            httpsProxy:
              description: httpsProxy is the URL of the proxy for HTTPS requests.
              type: string
            noProxy:
              description: noProxy is a comma-separated list of hostnames and/or CIDRs
                for which the proxy should not be used.
              type: string
          type: object
      required:
      - spec
      type: object
`)

func assetsCrd0000_03_configOperator_01_proxyCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_03_configOperator_01_proxyCrdYaml, nil
}

func assetsCrd0000_03_configOperator_01_proxyCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_03_configOperator_01_proxyCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_03_config-operator_01_proxy.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: clusterresourcequotas.quota.openshift.io
spec:
  group: quota.openshift.io
  names:
    kind: ClusterResourceQuota
    listKind: ClusterResourceQuotaList
    plural: clusterresourcequotas
    singular: clusterresourcequota
  scope: Cluster
  subresources:
    status: {}
  versions:
  - name: v1
    served: true
    storage: true
  validation:
    openAPIV3Schema:
      description: ClusterResourceQuota mirrors ResourceQuota at a cluster scope.  This
        object is easily convertible to synthetic ResourceQuota object to allow quota
        evaluation re-use.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object's metadata.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: Spec defines the desired quota
          properties:
            quota:
              description: Quota defines the desired quota
              properties:
                hard:
                  additionalProperties:
                    type: string
                  description: 'hard is the set of desired hard limits for each named
                    resource. More info: https://kubernetes.io/docs/concepts/policy/resource-quotas/'
                  type: object
                scopeSelector:
                  description: scopeSelector is also a collection of filters like
                    scopes that must match each object tracked by a quota but expressed
                    using ScopeSelectorOperator in combination with possible values.
                    For a resource to match, both scopes AND scopeSelector (if specified
                    in spec), must be matched.
                  properties:
                    matchExpressions:
                      description: A list of scope selector requirements by scope
                        of the resources.
                      items:
                        description: A scoped-resource selector requirement is a selector
                          that contains values, a scope name, and an operator that
                          relates the scope name and values.
                        properties:
                          operator:
                            description: Represents a scope's relationship to a set
                              of values. Valid operators are In, NotIn, Exists, DoesNotExist.
                            type: string
                          scopeName:
                            description: The name of the scope that the selector applies
                              to.
                            type: string
                          values:
                            description: An array of string values. If the operator
                              is In or NotIn, the values array must be non-empty.
                              If the operator is Exists or DoesNotExist, the values
                              array must be empty. This array is replaced during a
                              strategic merge patch.
                            items:
                              type: string
                            type: array
                        required:
                        - operator
                        - scopeName
                        type: object
                      type: array
                  type: object
                scopes:
                  description: A collection of filters that must match each object
                    tracked by a quota. If not specified, the quota matches all objects.
                  items:
                    description: A ResourceQuotaScope defines a filter that must match
                      each object tracked by a quota
                    type: string
                  type: array
              type: object
            selector:
              description: Selector is the selector used to match projects. It should
                only select active projects on the scale of dozens (though it can
                select many more less active projects).  These projects will contend
                on object creation through this resource.
              properties:
                annotations:
                  additionalProperties:
                    type: string
                  description: AnnotationSelector is used to select projects by annotation.
                  nullable: true
                  type: object
                labels:
                  description: LabelSelector is used to select projects by label.
                  nullable: true
                  properties:
                    matchExpressions:
                      description: matchExpressions is a list of label selector requirements.
                        The requirements are ANDed.
                      items:
                        description: A label selector requirement is a selector that
                          contains values, a key, and an operator that relates the
                          key and values.
                        properties:
                          key:
                            description: key is the label key that the selector applies
                              to.
                            type: string
                          operator:
                            description: operator represents a key's relationship
                              to a set of values. Valid operators are In, NotIn, Exists
                              and DoesNotExist.
                            type: string
                          values:
                            description: values is an array of string values. If the
                              operator is In or NotIn, the values array must be non-empty.
                              If the operator is Exists or DoesNotExist, the values
                              array must be empty. This array is replaced during a
                              strategic merge patch.
                            items:
                              type: string
                            type: array
                        required:
                        - key
                        - operator
                        type: object
                      type: array
                    matchLabels:
                      additionalProperties:
                        type: string
                      description: matchLabels is a map of {key,value} pairs. A single
                        {key,value} in the matchLabels map is equivalent to an element
                        of matchExpressions, whose key field is "key", the operator
                        is "In", and the values array contains only "value". The requirements
                        are ANDed.
                      type: object
                  type: object
              type: object
          required:
          - quota
          - selector
          type: object
        status:
          description: Status defines the actual enforced quota and its current usage
          properties:
            namespaces:
              description: Namespaces slices the usage by project.  This division
                allows for quick resolution of deletion reconciliation inside of a
                single project without requiring a recalculation across all projects.  This
                can be used to pull the deltas for a given project.
              items:
                description: ResourceQuotaStatusByNamespace gives status for a particular
                  project
                properties:
                  namespace:
                    description: Namespace the project this status applies to
                    type: string
                  status:
                    description: Status indicates how many resources have been consumed
                      by this project
                    properties:
                      hard:
                        additionalProperties:
                          type: string
                        description: 'Hard is the set of enforced hard limits for
                          each named resource. More info: https://kubernetes.io/docs/concepts/policy/resource-quotas/'
                        type: object
                      used:
                        additionalProperties:
                          type: string
                        description: Used is the current observed total usage of the
                          resource in the namespace.
                        type: object
                    type: object
                required:
                - namespace
                - status
                type: object
              nullable: true
              type: array
            total:
              description: Total defines the actual enforced quota and its current
                usage across all projects
              properties:
                hard:
                  additionalProperties:
                    type: string
                  description: 'Hard is the set of enforced hard limits for each named
                    resource. More info: https://kubernetes.io/docs/concepts/policy/resource-quotas/'
                  type: object
                used:
                  additionalProperties:
                    type: string
                  description: Used is the current observed total usage of the resource
                    in the namespace.
                  type: object
              type: object
          required:
          - total
          type: object
      required:
      - metadata
      - spec
      type: object
`)

func assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYaml, nil
}

func assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_03_quota-openshift_01_clusterresourcequota.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_03_securityOpenshift_01_sccCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: securitycontextconstraints.security.openshift.io
spec:
  group: security.openshift.io
  names:
    kind: SecurityContextConstraints
    listKind: SecurityContextConstraintsList
    plural: securitycontextconstraints
    singular: securitycontextconstraints
  scope: Cluster
  versions:
  - name: v1
    served: true
    storage: true
  validation:
    openAPIV3Schema:
      description: SecurityContextConstraints governs the ability to make requests
        that affect the SecurityContext that will be applied to a container. For historical
        reasons SCC was exposed under the core Kubernetes API group. That exposure
        is deprecated and will be removed in a future release - users should instead
        use the security.openshift.io group to manage SecurityContextConstraints.
      properties:
        allowHostDirVolumePlugin:
          description: AllowHostDirVolumePlugin determines if the policy allow containers
            to use the HostDir volume plugin
          type: boolean
        allowHostIPC:
          description: AllowHostIPC determines if the policy allows host ipc in the
            containers.
          type: boolean
        allowHostNetwork:
          description: AllowHostNetwork determines if the policy allows the use of
            HostNetwork in the pod spec.
          type: boolean
        allowHostPID:
          description: AllowHostPID determines if the policy allows host pid in the
            containers.
          type: boolean
        allowHostPorts:
          description: AllowHostPorts determines if the policy allows host ports in
            the containers.
          type: boolean
        allowPrivilegeEscalation:
          description: AllowPrivilegeEscalation determines if a pod can request to
            allow privilege escalation. If unspecified, defaults to true.
          nullable: true
          type: boolean
        allowPrivilegedContainer:
          description: AllowPrivilegedContainer determines if a container can request
            to be run as privileged.
          type: boolean
        allowedCapabilities:
          description: AllowedCapabilities is a list of capabilities that can be requested
            to add to the container. Capabilities in this field maybe added at the
            pod author's discretion. You must not list a capability in both AllowedCapabilities
            and RequiredDropCapabilities. To allow all capabilities you may use '*'.
          items:
            description: Capability represent POSIX capabilities type
            type: string
          nullable: true
          type: array
        allowedFlexVolumes:
          description: AllowedFlexVolumes is a whitelist of allowed Flexvolumes.  Empty
            or nil indicates that all Flexvolumes may be used.  This parameter is
            effective only when the usage of the Flexvolumes is allowed in the "Volumes"
            field.
          items:
            description: AllowedFlexVolume represents a single Flexvolume that is
              allowed to be used.
            properties:
              driver:
                description: Driver is the name of the Flexvolume driver.
                type: string
            required:
            - driver
            type: object
          nullable: true
          type: array
        allowedUnsafeSysctls:
          description: "AllowedUnsafeSysctls is a list of explicitly allowed unsafe
            sysctls, defaults to none. Each entry is either a plain sysctl name or
            ends in \"*\" in which case it is considered as a prefix of allowed sysctls.
            Single * means all unsafe sysctls are allowed. Kubelet has to whitelist
            all allowed unsafe sysctls explicitly to avoid rejection. \n Examples:
            e.g. \"foo/*\" allows \"foo/bar\", \"foo/baz\", etc. e.g. \"foo.*\" allows
            \"foo.bar\", \"foo.baz\", etc."
          items:
            type: string
          nullable: true
          type: array
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        defaultAddCapabilities:
          description: DefaultAddCapabilities is the default set of capabilities that
            will be added to the container unless the pod spec specifically drops
            the capability.  You may not list a capabiility in both DefaultAddCapabilities
            and RequiredDropCapabilities.
          items:
            description: Capability represent POSIX capabilities type
            type: string
          nullable: true
          type: array
        defaultAllowPrivilegeEscalation:
          description: DefaultAllowPrivilegeEscalation controls the default setting
            for whether a process can gain more privileges than its parent process.
          nullable: true
          type: boolean
        forbiddenSysctls:
          description: "ForbiddenSysctls is a list of explicitly forbidden sysctls,
            defaults to none. Each entry is either a plain sysctl name or ends in
            \"*\" in which case it is considered as a prefix of forbidden sysctls.
            Single * means all sysctls are forbidden. \n Examples: e.g. \"foo/*\"
            forbids \"foo/bar\", \"foo/baz\", etc. e.g. \"foo.*\" forbids \"foo.bar\",
            \"foo.baz\", etc."
          items:
            type: string
          nullable: true
          type: array
        fsGroup:
          description: FSGroup is the strategy that will dictate what fs group is
            used by the SecurityContext.
          nullable: true
          properties:
            ranges:
              description: Ranges are the allowed ranges of fs groups.  If you would
                like to force a single fs group then supply a single range with the
                same start and end.
              items:
                description: 'IDRange provides a min/max of an allowed range of IDs.
                  TODO: this could be reused for UIDs.'
                properties:
                  max:
                    description: Max is the end of the range, inclusive.
                    format: int64
                    type: integer
                  min:
                    description: Min is the start of the range, inclusive.
                    format: int64
                    type: integer
                type: object
              type: array
            type:
              description: Type is the strategy that will dictate what FSGroup is
                used in the SecurityContext.
              type: string
          type: object
        groups:
          description: The groups that have permission to use this security context
            constraints
          items:
            type: string
          nullable: true
          type: array
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: 'Standard object''s metadata. More info: http://releases.k8s.io/HEAD/docs/devel/api-conventions.md#metadata'
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        priority:
          description: Priority influences the sort order of SCCs when evaluating
            which SCCs to try first for a given pod request based on access in the
            Users and Groups fields.  The higher the int, the higher priority. An
            unset value is considered a 0 priority. If scores for multiple SCCs are
            equal they will be sorted from most restrictive to least restrictive.
            If both priorities and restrictions are equal the SCCs will be sorted
            by name.
          format: int32
          nullable: true
          type: integer
        readOnlyRootFilesystem:
          description: ReadOnlyRootFilesystem when set to true will force containers
            to run with a read only root file system.  If the container specifically
            requests to run with a non-read only root file system the SCC should deny
            the pod. If set to false the container may run with a read only root file
            system if it wishes but it will not be forced to.
          type: boolean
        requiredDropCapabilities:
          description: RequiredDropCapabilities are the capabilities that will be
            dropped from the container.  These are required to be dropped and cannot
            be added.
          items:
            description: Capability represent POSIX capabilities type
            type: string
          nullable: true
          type: array
        runAsUser:
          description: RunAsUser is the strategy that will dictate what RunAsUser
            is used in the SecurityContext.
          nullable: true
          properties:
            type:
              description: Type is the strategy that will dictate what RunAsUser is
                used in the SecurityContext.
              type: string
            uid:
              description: UID is the user id that containers must run as.  Required
                for the MustRunAs strategy if not using namespace/service account
                allocated uids.
              format: int64
              type: integer
            uidRangeMax:
              description: UIDRangeMax defines the max value for a strategy that allocates
                by range.
              format: int64
              type: integer
            uidRangeMin:
              description: UIDRangeMin defines the min value for a strategy that allocates
                by range.
              format: int64
              type: integer
          type: object
        seLinuxContext:
          description: SELinuxContext is the strategy that will dictate what labels
            will be set in the SecurityContext.
          nullable: true
          properties:
            seLinuxOptions:
              description: seLinuxOptions required to run as; required for MustRunAs
              properties:
                level:
                  description: Level is SELinux level label that applies to the container.
                  type: string
                role:
                  description: Role is a SELinux role label that applies to the container.
                  type: string
                type:
                  description: Type is a SELinux type label that applies to the container.
                  type: string
                user:
                  description: User is a SELinux user label that applies to the container.
                  type: string
              type: object
            type:
              description: Type is the strategy that will dictate what SELinux context
                is used in the SecurityContext.
              type: string
          type: object
        seccompProfiles:
          description: "SeccompProfiles lists the allowed profiles that may be set
            for the pod or container's seccomp annotations.  An unset (nil) or empty
            value means that no profiles may be specifid by the pod or container.\tThe
            wildcard '*' may be used to allow all profiles.  When used to generate
            a value for a pod the first non-wildcard profile will be used as the default."
          items:
            type: string
          nullable: true
          type: array
        supplementalGroups:
          description: SupplementalGroups is the strategy that will dictate what supplemental
            groups are used by the SecurityContext.
          nullable: true
          properties:
            ranges:
              description: Ranges are the allowed ranges of supplemental groups.  If
                you would like to force a single supplemental group then supply a
                single range with the same start and end.
              items:
                description: 'IDRange provides a min/max of an allowed range of IDs.
                  TODO: this could be reused for UIDs.'
                properties:
                  max:
                    description: Max is the end of the range, inclusive.
                    format: int64
                    type: integer
                  min:
                    description: Min is the start of the range, inclusive.
                    format: int64
                    type: integer
                type: object
              type: array
            type:
              description: Type is the strategy that will dictate what supplemental
                groups is used in the SecurityContext.
              type: string
          type: object
        users:
          description: The users who have permissions to use this security context
            constraints
          items:
            type: string
          nullable: true
          type: array
        volumes:
          description: Volumes is a white list of allowed volume plugins.  FSType
            corresponds directly with the field names of a VolumeSource (azureFile,
            configMap, emptyDir).  To allow all volumes you may use "*". To allow
            no volumes, set to ["none"].
          items:
            description: FS Type gives strong typing to different file systems that
              are used by volumes.
            type: string
          nullable: true
          type: array
      required:
      - allowHostDirVolumePlugin
      - allowHostIPC
      - allowHostNetwork
      - allowHostPID
      - allowHostPorts
      - allowPrivilegedContainer
      - allowedCapabilities
      - defaultAddCapabilities
      - priority
      - readOnlyRootFilesystem
      - requiredDropCapabilities
      - volumes
      type: object
`)

func assetsCrd0000_03_securityOpenshift_01_sccCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_03_securityOpenshift_01_sccCrdYaml, nil
}

func assetsCrd0000_03_securityOpenshift_01_sccCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_03_securityOpenshift_01_sccCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_03_security-openshift_01_scc.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_apiserverCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: apiservers.config.openshift.io
spec:
  group: config.openshift.io
  scope: Cluster
  names:
    kind: APIServer
    singular: apiserver
    plural: apiservers
    listKind: APIServerList
  versions:
  - name: v1
    served: true
    storage: true
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: APIServer holds configuration (like serving certificates, client
        CA and CORS domains) shared by all API servers in the system, among them especially
        kube-apiserver and openshift-apiserver. The canonical name of an instance
        is 'cluster'.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: ObjectMeta is metadata that all persisted resources must have,
            which includes all objects users must create.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          properties:
            additionalCORSAllowedOrigins:
              description: additionalCORSAllowedOrigins lists additional, user-defined
                regular expressions describing hosts for which the API server allows
                access using the CORS headers. This may be needed to access the API
                and the integrated OAuth server from JavaScript applications. The
                values are regular expressions that correspond to the Golang regular
                expression language.
              items:
                type: string
              type: array
            clientCA:
              description: 'clientCA references a ConfigMap containing a certificate
                bundle for the signers that will be recognized for incoming client
                certificates in addition to the operator managed signers. If this
                is empty, then only operator managed signers are valid. You usually
                only have to set this if you have your own PKI you wish to honor client
                certificates from. The ConfigMap must exist in the openshift-config
                namespace and contain the following required fields: - ConfigMap.Data["ca-bundle.crt"]
                - CA bundle.'
              properties:
                name:
                  description: name is the metadata.name of the referenced config
                    map
                  type: string
              required:
              - name
              type: object
            servingCerts:
              description: servingCert is the TLS cert info for serving secure traffic.
                If not specified, operator managed certificates will be used for serving
                secure traffic.
              properties:
                namedCertificates:
                  description: namedCertificates references secrets containing the
                    TLS cert info for serving secure traffic to specific hostnames.
                    If no named certificates are provided, or no named certificates
                    match the server name as understood by a client, the defaultServingCertificate
                    will be used.
                  items:
                    description: APIServerNamedServingCert maps a server DNS name,
                      as understood by a client, to a certificate.
                    properties:
                      names:
                        description: names is a optional list of explicit DNS names
                          (leading wildcards allowed) that should use this certificate
                          to serve secure traffic. If no names are provided, the implicit
                          names will be extracted from the certificates. Exact names
                          trump over wildcard names. Explicit names defined here trump
                          over extracted implicit names.
                        items:
                          type: string
                        type: array
                      servingCertificate:
                        description: 'servingCertificate references a kubernetes.io/tls
                          type secret containing the TLS cert info for serving secure
                          traffic. The secret must exist in the openshift-config namespace
                          and contain the following required fields: - Secret.Data["tls.key"]
                          - TLS private key. - Secret.Data["tls.crt"] - TLS certificate.'
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              secret
                            type: string
                        required:
                        - name
                        type: object
                    type: object
                  type: array
              type: object
          type: object
        status:
          type: object
      required:
      - spec
      type: object
`)

func assetsCrd0000_10_configOperator_01_apiserverCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_apiserverCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_apiserverCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_apiserverCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_apiserver.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_authenticationCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: authentications.config.openshift.io
spec:
  group: config.openshift.io
  names:
    kind: Authentication
    listKind: AuthenticationList
    plural: authentications
    singular: authentication
  scope: Cluster
  subresources:
    status: {}
  versions:
  - name: v1
    served: true
    storage: true
  validation:
    openAPIV3Schema:
      description: Authentication specifies cluster-wide settings for authentication
        (like OAuth and webhook token authenticators). The canonical name of an instance
        is ` + "`" + `cluster` + "`" + `.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object's metadata.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: spec holds user settable values for configuration
          properties:
            oauthMetadata:
              description: 'oauthMetadata contains the discovery endpoint data for
                OAuth 2.0 Authorization Server Metadata for an external OAuth server.
                This discovery document can be viewed from its served location: oc
                get --raw ''/.well-known/oauth-authorization-server'' For further
                details, see the IETF Draft: https://tools.ietf.org/html/draft-ietf-oauth-discovery-04#section-2
                If oauthMetadata.name is non-empty, this value has precedence over
                any metadata reference stored in status. The key "oauthMetadata" is
                used to locate the data. If specified and the config map or expected
                key is not found, no metadata is served. If the specified metadata
                is not valid, no metadata is served. The namespace for this config
                map is openshift-config.'
              properties:
                name:
                  description: name is the metadata.name of the referenced config
                    map
                  type: string
              required:
              - name
              type: object
            type:
              description: type identifies the cluster managed, user facing authentication
                mode in use. Specifically, it manages the component that responds
                to login attempts. The default is IntegratedOAuth.
              type: string
            webhookTokenAuthenticators:
              description: webhookTokenAuthenticators configures remote token reviewers.
                These remote authentication webhooks can be used to verify bearer
                tokens via the tokenreviews.authentication.k8s.io REST API.  This
                is required to honor bearer tokens that are provisioned by an external
                authentication service. The namespace for these secrets is openshift-config.
              items:
                description: webhookTokenAuthenticator holds the necessary configuration
                  options for a remote token authenticator
                properties:
                  kubeConfig:
                    description: 'kubeConfig contains kube config file data which
                      describes how to access the remote webhook service. For further
                      details, see: https://kubernetes.io/docs/reference/access-authn-authz/authentication/#webhook-token-authentication
                      The key "kubeConfig" is used to locate the data. If the secret
                      or expected key is not found, the webhook is not honored. If
                      the specified kube config data is not valid, the webhook is
                      not honored. The namespace for this secret is determined by
                      the point of use.'
                    properties:
                      name:
                        description: name is the metadata.name of the referenced secret
                        type: string
                    required:
                    - name
                    type: object
                type: object
              type: array
          type: object
        status:
          description: status holds observed values from the cluster. They may not
            be overridden.
          properties:
            integratedOAuthMetadata:
              description: 'integratedOAuthMetadata contains the discovery endpoint
                data for OAuth 2.0 Authorization Server Metadata for the in-cluster
                integrated OAuth server. This discovery document can be viewed from
                its served location: oc get --raw ''/.well-known/oauth-authorization-server''
                For further details, see the IETF Draft: https://tools.ietf.org/html/draft-ietf-oauth-discovery-04#section-2
                This contains the observed value based on cluster state. An explicitly
                set value in spec.oauthMetadata has precedence over this field. This
                field has no meaning if authentication spec.type is not set to IntegratedOAuth.
                The key "oauthMetadata" is used to locate the data. If the config
                map or expected key is not found, no metadata is served. If the specified
                metadata is not valid, no metadata is served. The namespace for this
                config map is openshift-config-managed.'
              properties:
                name:
                  description: name is the metadata.name of the referenced config
                    map
                  type: string
              required:
              - name
              type: object
          type: object
      required:
      - spec
      type: object
`)

func assetsCrd0000_10_configOperator_01_authenticationCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_authenticationCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_authenticationCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_authenticationCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_authentication.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_buildCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: builds.config.openshift.io
spec:
  group: config.openshift.io
  scope: Cluster
  names:
    kind: Build
    singular: build
    plural: builds
    listKind: BuildList
  versions:
  - name: v1
    served: true
    storage: true
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: "Build configures the behavior of OpenShift builds for the entire
        cluster. This includes default settings that can be overridden in BuildConfig
        objects, and overrides which are applied to all builds. \n The canonical name
        is \"cluster\""
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: ObjectMeta is metadata that all persisted resources must have,
            which includes all objects users must create.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: Spec holds user-settable values for the build controller configuration
          properties:
            additionalTrustedCA:
              description: "AdditionalTrustedCA is a reference to a ConfigMap containing
                additional CAs that should be trusted for image pushes and pulls during
                builds. The namespace for this config map is openshift-config. \n
                DEPRECATED: Additional CAs for image pull and push should be set on
                image.config.openshift.io/cluster instead."
              properties:
                name:
                  description: name is the metadata.name of the referenced config
                    map
                  type: string
              required:
              - name
              type: object
            buildDefaults:
              description: BuildDefaults controls the default information for Builds
              properties:
                defaultProxy:
                  description: "DefaultProxy contains the default proxy settings for
                    all build operations, including image pull/push and source download.
                    \n Values can be overrode by setting the ` + "`" + `HTTP_PROXY` + "`" + `, ` + "`" + `HTTPS_PROXY` + "`" + `,
                    and ` + "`" + `NO_PROXY` + "`" + ` environment variables in the build config's strategy."
                  properties:
                    httpProxy:
                      description: httpProxy is the URL of the proxy for HTTP requests.  Empty
                        means unset and will not result in an env var.
                      type: string
                    httpsProxy:
                      description: httpsProxy is the URL of the proxy for HTTPS requests.  Empty
                        means unset and will not result in an env var.
                      type: string
                    noProxy:
                      description: noProxy is a comma-separated list of hostnames
                        and/or CIDRs for which the proxy should not be used. Empty
                        means unset and will not result in an env var.
                      type: string
                    readinessEndpoints:
                      description: readinessEndpoints is a list of endpoints used
                        to verify readiness of the proxy.
                      items:
                        type: string
                      type: array
                    trustedCA:
                      description: "trustedCA is a reference to a ConfigMap containing
                        a CA certificate bundle used for client egress HTTPS connections.
                        The certificate bundle must be from the CA that signed the
                        proxy's certificate and be signed for everything. The trustedCA
                        field should only be consumed by a proxy validator. The validator
                        is responsible for reading the certificate bundle from required
                        key \"ca-bundle.crt\" and copying it to a ConfigMap named
                        \"trusted-ca-bundle\" in the \"openshift-config-managed\"
                        namespace. The namespace for the ConfigMap referenced by trustedCA
                        is \"openshift-config\". Here is an example ConfigMap (in
                        yaml): \n apiVersion: v1 kind: ConfigMap metadata:  name:
                        user-ca-bundle  namespace: openshift-config  data:    ca-bundle.crt:
                        |      -----BEGIN CERTIFICATE-----      Custom CA certificate
                        bundle.      -----END CERTIFICATE-----"
                      properties:
                        name:
                          description: name is the metadata.name of the referenced
                            config map
                          type: string
                      required:
                      - name
                      type: object
                  type: object
                env:
                  description: Env is a set of default environment variables that
                    will be applied to the build if the specified variables do not
                    exist on the build
                  items:
                    description: EnvVar represents an environment variable present
                      in a Container.
                    properties:
                      name:
                        description: Name of the environment variable. Must be a C_IDENTIFIER.
                        type: string
                      value:
                        description: 'Variable references $(VAR_NAME) are expanded
                          using the previous defined environment variables in the
                          container and any service environment variables. If a variable
                          cannot be resolved, the reference in the input string will
                          be unchanged. The $(VAR_NAME) syntax can be escaped with
                          a double $$, ie: $$(VAR_NAME). Escaped references will never
                          be expanded, regardless of whether the variable exists or
                          not. Defaults to "".'
                        type: string
                      valueFrom:
                        description: Source for the environment variable's value.
                          Cannot be used if value is not empty.
                        properties:
                          configMapKeyRef:
                            description: Selects a key of a ConfigMap.
                            properties:
                              key:
                                description: The key to select.
                                type: string
                              name:
                                description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                  TODO: Add other useful fields. apiVersion, kind,
                                  uid?'
                                type: string
                              optional:
                                description: Specify whether the ConfigMap or it's
                                  key must be defined
                                type: boolean
                            required:
                            - key
                            type: object
                          fieldRef:
                            description: 'Selects a field of the pod: supports metadata.name,
                              metadata.namespace, metadata.labels, metadata.annotations,
                              spec.nodeName, spec.serviceAccountName, status.hostIP,
                              status.podIP.'
                            properties:
                              apiVersion:
                                description: Version of the schema the FieldPath is
                                  written in terms of, defaults to "v1".
                                type: string
                              fieldPath:
                                description: Path of the field to select in the specified
                                  API version.
                                type: string
                            required:
                            - fieldPath
                            type: object
                          resourceFieldRef:
                            description: 'Selects a resource of the container: only
                              resources limits and requests (limits.cpu, limits.memory,
                              limits.ephemeral-storage, requests.cpu, requests.memory
                              and requests.ephemeral-storage) are currently supported.'
                            properties:
                              containerName:
                                description: 'Container name: required for volumes,
                                  optional for env vars'
                                type: string
                              divisor:
                                description: Specifies the output format of the exposed
                                  resources, defaults to "1"
                                type: string
                              resource:
                                description: 'Required: resource to select'
                                type: string
                            required:
                            - resource
                            type: object
                          secretKeyRef:
                            description: Selects a key of a secret in the pod's namespace
                            properties:
                              key:
                                description: The key of the secret to select from.  Must
                                  be a valid secret key.
                                type: string
                              name:
                                description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                  TODO: Add other useful fields. apiVersion, kind,
                                  uid?'
                                type: string
                              optional:
                                description: Specify whether the Secret or it's key
                                  must be defined
                                type: boolean
                            required:
                            - key
                            type: object
                        type: object
                    required:
                    - name
                    type: object
                  type: array
                gitProxy:
                  description: "GitProxy contains the proxy settings for git operations
                    only. If set, this will override any Proxy settings for all git
                    commands, such as git clone. \n Values that are not set here will
                    be inherited from DefaultProxy."
                  properties:
                    httpProxy:
                      description: httpProxy is the URL of the proxy for HTTP requests.  Empty
                        means unset and will not result in an env var.
                      type: string
                    httpsProxy:
                      description: httpsProxy is the URL of the proxy for HTTPS requests.  Empty
                        means unset and will not result in an env var.
                      type: string
                    noProxy:
                      description: noProxy is a comma-separated list of hostnames
                        and/or CIDRs for which the proxy should not be used. Empty
                        means unset and will not result in an env var.
                      type: string
                    readinessEndpoints:
                      description: readinessEndpoints is a list of endpoints used
                        to verify readiness of the proxy.
                      items:
                        type: string
                      type: array
                    trustedCA:
                      description: "trustedCA is a reference to a ConfigMap containing
                        a CA certificate bundle used for client egress HTTPS connections.
                        The certificate bundle must be from the CA that signed the
                        proxy's certificate and be signed for everything. The trustedCA
                        field should only be consumed by a proxy validator. The validator
                        is responsible for reading the certificate bundle from required
                        key \"ca-bundle.crt\" and copying it to a ConfigMap named
                        \"trusted-ca-bundle\" in the \"openshift-config-managed\"
                        namespace. The namespace for the ConfigMap referenced by trustedCA
                        is \"openshift-config\". Here is an example ConfigMap (in
                        yaml): \n apiVersion: v1 kind: ConfigMap metadata:  name:
                        user-ca-bundle  namespace: openshift-config  data:    ca-bundle.crt:
                        |      -----BEGIN CERTIFICATE-----      Custom CA certificate
                        bundle.      -----END CERTIFICATE-----"
                      properties:
                        name:
                          description: name is the metadata.name of the referenced
                            config map
                          type: string
                      required:
                      - name
                      type: object
                  type: object
                imageLabels:
                  description: ImageLabels is a list of docker labels that are applied
                    to the resulting image. User can override a default label by providing
                    a label with the same name in their Build/BuildConfig.
                  items:
                    properties:
                      name:
                        description: Name defines the name of the label. It must have
                          non-zero length.
                        type: string
                      value:
                        description: Value defines the literal value of the label.
                        type: string
                    type: object
                  type: array
                resources:
                  description: Resources defines resource requirements to execute
                    the build.
                  properties:
                    limits:
                      additionalProperties:
                        type: string
                      description: 'Limits describes the maximum amount of compute
                        resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                      type: object
                    requests:
                      additionalProperties:
                        type: string
                      description: 'Requests describes the minimum amount of compute
                        resources required. If Requests is omitted for a container,
                        it defaults to Limits if that is explicitly specified, otherwise
                        to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                      type: object
                  type: object
              type: object
            buildOverrides:
              description: BuildOverrides controls override settings for builds
              properties:
                imageLabels:
                  description: ImageLabels is a list of docker labels that are applied
                    to the resulting image. If user provided a label in their Build/BuildConfig
                    with the same name as one in this list, the user's label will
                    be overwritten.
                  items:
                    properties:
                      name:
                        description: Name defines the name of the label. It must have
                          non-zero length.
                        type: string
                      value:
                        description: Value defines the literal value of the label.
                        type: string
                    type: object
                  type: array
                nodeSelector:
                  additionalProperties:
                    type: string
                  description: NodeSelector is a selector which must be true for the
                    build pod to fit on a node
                  type: object
                tolerations:
                  description: Tolerations is a list of Tolerations that will override
                    any existing tolerations set on a build pod.
                  items:
                    description: The pod this Toleration is attached to tolerates
                      any taint that matches the triple <key,value,effect> using the
                      matching operator <operator>.
                    properties:
                      effect:
                        description: Effect indicates the taint effect to match. Empty
                          means match all taint effects. When specified, allowed values
                          are NoSchedule, PreferNoSchedule and NoExecute.
                        type: string
                      key:
                        description: Key is the taint key that the toleration applies
                          to. Empty means match all taint keys. If the key is empty,
                          operator must be Exists; this combination means to match
                          all values and all keys.
                        type: string
                      operator:
                        description: Operator represents a key's relationship to the
                          value. Valid operators are Exists and Equal. Defaults to
                          Equal. Exists is equivalent to wildcard for value, so that
                          a pod can tolerate all taints of a particular category.
                        type: string
                      tolerationSeconds:
                        description: TolerationSeconds represents the period of time
                          the toleration (which must be of effect NoExecute, otherwise
                          this field is ignored) tolerates the taint. By default,
                          it is not set, which means tolerate the taint forever (do
                          not evict). Zero and negative values will be treated as
                          0 (evict immediately) by the system.
                        format: int64
                        type: integer
                      value:
                        description: Value is the taint value the toleration matches
                          to. If the operator is Exists, the value should be empty,
                          otherwise just a regular string.
                        type: string
                    type: object
                  type: array
              type: object
          type: object
      required:
      - spec
      type: object
`)

func assetsCrd0000_10_configOperator_01_buildCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_buildCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_buildCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_buildCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_build.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_consoleCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: consoles.config.openshift.io
spec:
  scope: Cluster
  group: config.openshift.io
  names:
    kind: Console
    listKind: ConsoleList
    plural: consoles
    singular: console
  subresources:
    status: {}
  versions:
  - name: v1
    served: true
    storage: true
  validation:
    openAPIV3Schema:
      description: Console holds cluster-wide configuration for the web console, including
        the logout URL, and reports the public URL of the console. The canonical name
        is ` + "`" + `cluster` + "`" + `.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object's metadata.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: spec holds user settable values for configuration
          properties:
            authentication:
              description: ConsoleAuthentication defines a list of optional configuration
                for console authentication.
              properties:
                logoutRedirect:
                  description: 'An optional, absolute URL to redirect web browsers
                    to after logging out of the console. If not specified, it will
                    redirect to the default login page. This is required when using
                    an identity provider that supports single sign-on (SSO) such as:
                    - OpenID (Keycloak, Azure) - RequestHeader (GSSAPI, SSPI, SAML)
                    - OAuth (GitHub, GitLab, Google) Logging out of the console will
                    destroy the user''s token. The logoutRedirect provides the user
                    the option to perform single logout (SLO) through the identity
                    provider to destroy their single sign-on session.'
                  pattern: ^$|^((https):\/\/?)[^\s()<>]+(?:\([\w\d]+\)|([^[:punct:]\s]|\/?))$
                  type: string
              type: object
          type: object
        status:
          description: status holds observed values from the cluster. They may not
            be overridden.
          properties:
            consoleURL:
              description: The URL for the console. This will be derived from the
                host for the route that is created for the console.
              type: string
          type: object
      required:
      - spec
      type: object
`)

func assetsCrd0000_10_configOperator_01_consoleCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_consoleCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_consoleCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_consoleCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_console.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_dnsCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: dnses.config.openshift.io
spec:
  group: config.openshift.io
  names:
    kind: DNS
    listKind: DNSList
    plural: dnses
    singular: dns
  scope: Cluster
  versions:
  - name: v1
    served: true
    storage: true
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: DNS holds cluster-wide information about DNS. The canonical name
        is ` + "`" + `cluster` + "`" + `
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object's metadata.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: spec holds user settable values for configuration
          properties:
            baseDomain:
              description: "baseDomain is the base domain of the cluster. All managed
                DNS records will be sub-domains of this base. \n For example, given
                the base domain ` + "`" + `openshift.example.com` + "`" + `, an API server DNS record
                may be created for ` + "`" + `cluster-api.openshift.example.com` + "`" + `. \n Once set,
                this field cannot be changed."
              type: string
            privateZone:
              description: "privateZone is the location where all the DNS records
                that are only available internally to the cluster exist. \n If this
                field is nil, no private records should be created. \n Once set, this
                field cannot be changed."
              properties:
                id:
                  description: "id is the identifier that can be used to find the
                    DNS hosted zone. \n on AWS zone can be fetched using ` + "`" + `ID` + "`" + ` as id
                    in [1] on Azure zone can be fetched using ` + "`" + `ID` + "`" + ` as a pre-determined
                    name in [2], on GCP zone can be fetched using ` + "`" + `ID` + "`" + ` as a pre-determined
                    name in [3]. \n [1]: https://docs.aws.amazon.com/cli/latest/reference/route53/get-hosted-zone.html#options
                    [2]: https://docs.microsoft.com/en-us/cli/azure/network/dns/zone?view=azure-cli-latest#az-network-dns-zone-show
                    [3]: https://cloud.google.com/dns/docs/reference/v1/managedZones/get"
                  type: string
                tags:
                  additionalProperties:
                    type: string
                  description: "tags can be used to query the DNS hosted zone. \n
                    on AWS, resourcegroupstaggingapi [1] can be used to fetch a zone
                    using ` + "`" + `Tags` + "`" + ` as tag-filters, \n [1]: https://docs.aws.amazon.com/cli/latest/reference/resourcegroupstaggingapi/get-resources.html#options"
                  type: object
              type: object
            publicZone:
              description: "publicZone is the location where all the DNS records that
                are publicly accessible to the internet exist. \n If this field is
                nil, no public records should be created. \n Once set, this field
                cannot be changed."
              properties:
                id:
                  description: "id is the identifier that can be used to find the
                    DNS hosted zone. \n on AWS zone can be fetched using ` + "`" + `ID` + "`" + ` as id
                    in [1] on Azure zone can be fetched using ` + "`" + `ID` + "`" + ` as a pre-determined
                    name in [2], on GCP zone can be fetched using ` + "`" + `ID` + "`" + ` as a pre-determined
                    name in [3]. \n [1]: https://docs.aws.amazon.com/cli/latest/reference/route53/get-hosted-zone.html#options
                    [2]: https://docs.microsoft.com/en-us/cli/azure/network/dns/zone?view=azure-cli-latest#az-network-dns-zone-show
                    [3]: https://cloud.google.com/dns/docs/reference/v1/managedZones/get"
                  type: string
                tags:
                  additionalProperties:
                    type: string
                  description: "tags can be used to query the DNS hosted zone. \n
                    on AWS, resourcegroupstaggingapi [1] can be used to fetch a zone
                    using ` + "`" + `Tags` + "`" + ` as tag-filters, \n [1]: https://docs.aws.amazon.com/cli/latest/reference/resourcegroupstaggingapi/get-resources.html#options"
                  type: object
              type: object
          type: object
        status:
          description: status holds observed values from the cluster. They may not
            be overridden.
          type: object
      required:
      - spec
      type: object
`)

func assetsCrd0000_10_configOperator_01_dnsCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_dnsCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_dnsCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_dnsCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_dns.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_featuregateCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: featuregates.config.openshift.io
spec:
  group: config.openshift.io
  version: v1
  scope: Cluster
  names:
    kind: FeatureGate
    singular: featuregate
    plural: featuregates
    listKind: FeatureGateList
  versions:
  - name: v1
    served: true
    storage: true
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: Feature holds cluster-wide information about feature gates.  The
        canonical name is ` + "`" + `cluster` + "`" + `
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object's metadata.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: spec holds user settable values for configuration
          properties:
            customNoUpgrade:
              description: customNoUpgrade allows the enabling or disabling of any
                feature. Turning this feature set on IS NOT SUPPORTED, CANNOT BE UNDONE,
                and PREVENTS UPGRADES. Because of its nature, this setting cannot
                be validated.  If you have any typos or accidentally apply invalid
                combinations your cluster may fail in an unrecoverable way.  featureSet
                must equal "CustomNoUpgrade" must be set to use this field.
              nullable: true
              properties:
                disabled:
                  description: disabled is a list of all feature gates that you want
                    to force off
                  items:
                    type: string
                  type: array
                enabled:
                  description: enabled is a list of all feature gates that you want
                    to force on
                  items:
                    type: string
                  type: array
              type: object
            featureSet:
              description: featureSet changes the list of features in the cluster.  The
                default is empty.  Be very careful adjusting this setting. Turning
                on or off features may cause irreversible changes in your cluster
                which cannot be undone.
              type: string
          type: object
        status:
          description: status holds observed values from the cluster. They may not
            be overridden.
          type: object
      required:
      - spec
      type: object
`)

func assetsCrd0000_10_configOperator_01_featuregateCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_featuregateCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_featuregateCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_featuregateCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_featuregate.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_imageCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: images.config.openshift.io
spec:
  group: config.openshift.io
  scope: Cluster
  names:
    kind: Image
    singular: image
    plural: images
    listKind: ImageList
  versions:
  - name: v1
    served: true
    storage: true
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: Image governs policies related to imagestream imports and runtime
        configuration for external registries. It allows cluster admins to configure
        which registries OpenShift is allowed to import images from, extra CA trust
        bundles for external registries, and policies to blacklist/whitelist registry
        hostnames. When exposing OpenShift's image registry to the public, this also
        lets cluster admins specify the external hostname.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object's metadata.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: spec holds user settable values for configuration
          properties:
            additionalTrustedCA:
              description: additionalTrustedCA is a reference to a ConfigMap containing
                additional CAs that should be trusted during imagestream import, pod
                image pull, build image pull, and imageregistry pullthrough. The namespace
                for this config map is openshift-config.
              properties:
                name:
                  description: name is the metadata.name of the referenced config
                    map
                  type: string
              required:
              - name
              type: object
            allowedRegistriesForImport:
              description: allowedRegistriesForImport limits the container image registries
                that normal users may import images from. Set this list to the registries
                that you trust to contain valid Docker images and that you want applications
                to be able to import from. Users with permission to create Images
                or ImageStreamMappings via the API are not affected by this policy
                - typically only administrators or system integrations will have those
                permissions.
              items:
                description: RegistryLocation contains a location of the registry
                  specified by the registry domain name. The domain name might include
                  wildcards, like '*' or '??'.
                properties:
                  domainName:
                    description: domainName specifies a domain name for the registry
                      In case the registry use non-standard (80 or 443) port, the
                      port should be included in the domain name as well.
                    type: string
                  insecure:
                    description: insecure indicates whether the registry is secure
                      (https) or insecure (http) By default (if not specified) the
                      registry is assumed as secure.
                    type: boolean
                type: object
              type: array
            externalRegistryHostnames:
              description: externalRegistryHostnames provides the hostnames for the
                default external image registry. The external hostname should be set
                only when the image registry is exposed externally. The first value
                is used in 'publicDockerImageRepository' field in ImageStreams. The
                value must be in "hostname[:port]" format.
              items:
                type: string
              type: array
            registrySources:
              description: registrySources contains configuration that determines
                how the container runtime should treat individual registries when
                accessing images for builds+pods. (e.g. whether or not to allow insecure
                access).  It does not contain configuration for the internal cluster
                registry.
              properties:
                allowedRegistries:
                  description: "allowedRegistries are whitelisted for image pull/push.
                    All other registries are blocked. \n Only one of BlockedRegistries
                    or AllowedRegistries may be set."
                  items:
                    type: string
                  type: array
                blockedRegistries:
                  description: "blockedRegistries are blacklisted from image pull/push.
                    All other registries are allowed. \n Only one of BlockedRegistries
                    or AllowedRegistries may be set."
                  items:
                    type: string
                  type: array
                insecureRegistries:
                  description: insecureRegistries are registries which do not have
                    a valid TLS certificates or only support HTTP connections.
                  items:
                    type: string
                  type: array
              type: object
          type: object
        status:
          description: status holds observed values from the cluster. They may not
            be overridden.
          properties:
            externalRegistryHostnames:
              description: externalRegistryHostnames provides the hostnames for the
                default external image registry. The external hostname should be set
                only when the image registry is exposed externally. The first value
                is used in 'publicDockerImageRepository' field in ImageStreams. The
                value must be in "hostname[:port]" format.
              items:
                type: string
              type: array
            internalRegistryHostname:
              description: internalRegistryHostname sets the hostname for the default
                internal image registry. The value must be in "hostname[:port]" format.
                This value is set by the image registry operator which controls the
                internal registry hostname. For backward compatibility, users can
                still use OPENSHIFT_DEFAULT_REGISTRY environment variable but this
                setting overrides the environment variable.
              type: string
          type: object
      required:
      - spec
      type: object
`)

func assetsCrd0000_10_configOperator_01_imageCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_imageCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_imageCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_imageCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_image.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: imagecontentsourcepolicies.operator.openshift.io
spec:
  group: operator.openshift.io
  scope: Cluster
  names:
    kind: ImageContentSourcePolicy
    singular: imagecontentsourcepolicy
    plural: imagecontentsourcepolicies
    listKind: ImageContentSourcePolicyList
  versions:
  - name: v1alpha1
    served: true
    storage: true
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: ImageContentSourcePolicy holds cluster-wide information about how
        to handle registry mirror rules. When multiple policies are defined, the outcome
        of the behavior is defined on each field.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object's metadata.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: spec holds user settable values for configuration
          properties:
            repositoryDigestMirrors:
              description: "repositoryDigestMirrors allows images referenced by image
                digests in pods to be pulled from alternative mirrored repository
                locations. The image pull specification provided to the pod will be
                compared to the source locations described in RepositoryDigestMirrors
                and the image may be pulled down from any of the mirrors in the list
                instead of the specified repository allowing administrators to choose
                a potentially faster mirror. Only image pull specifications that have
                an image disgest will have this behavior applied to them - tags will
                continue to be pulled from the specified repository in the pull spec.
                \n Each “source” repository is treated independently; configurations
                for different “source” repositories don’t interact. \n When multiple
                policies are defined for the same “source” repository, the sets of
                defined mirrors will be merged together, preserving the relative order
                of the mirrors, if possible. For example, if policy A has mirrors
                ` + "`" + `a, b, c` + "`" + ` and policy B has mirrors ` + "`" + `c, d, e` + "`" + `, the mirrors will be
                used in the order ` + "`" + `a, b, c, d, e` + "`" + `.  If the orders of mirror entries
                conflict (e.g. ` + "`" + `a, b` + "`" + ` vs. ` + "`" + `b, a` + "`" + `) the configuration is not rejected
                but the resulting order is unspecified."
              items:
                description: 'RepositoryDigestMirrors holds cluster-wide information
                  about how to handle mirros in the registries config. Note: the mirrors
                  only work when pulling the images that are referenced by their digests.'
                properties:
                  mirrors:
                    description: mirrors is one or more repositories that may also
                      contain the same images. The order of mirrors in this list is
                      treated as the user's desired priority, while source is by default
                      considered lower priority than all mirrors. Other cluster configuration,
                      including (but not limited to) other repositoryDigestMirrors
                      objects, may impact the exact order mirrors are contacted in,
                      or some mirrors may be contacted in parallel, so this should
                      be considered a preference rather than a guarantee of ordering.
                    items:
                      type: string
                    type: array
                  source:
                    description: source is the repository that users refer to, e.g.
                      in image pull specifications.
                    type: string
                required:
                - source
                type: object
              type: array
          type: object
      required:
      - spec
      type: object
`)

func assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_imagecontentsourcepolicy.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_infrastructureCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: infrastructures.config.openshift.io
spec:
  group: config.openshift.io
  names:
    kind: Infrastructure
    listKind: InfrastructureList
    plural: infrastructures
    singular: infrastructure
  scope: Cluster
  versions:
  - name: v1
    served: true
    storage: true
  validation:
    openAPIV3Schema:
      description: Infrastructure holds cluster-wide information about Infrastructure.  The
        canonical name is ` + "`" + `cluster` + "`" + `
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object's metadata.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: spec holds user settable values for configuration
          properties:
            cloudConfig:
              description: cloudConfig is a reference to a ConfigMap containing the
                cloud provider configuration file. This configuration file is used
                to configure the Kubernetes cloud provider integration when using
                the built-in cloud provider integration or the external cloud controller
                manager. The namespace for this config map is openshift-config.
              properties:
                key:
                  description: Key allows pointing to a specific key/value inside
                    of the configmap.  This is useful for logical file references.
                  type: string
                name:
                  type: string
              type: object
          type: object
        status:
          description: status holds observed values from the cluster. They may not
            be overridden.
          properties:
            apiServerInternalURI:
              description: apiServerInternalURL is a valid URI with scheme(http/https),
                address and port.  apiServerInternalURL can be used by components
                like kubelets, to contact the Kubernetes API server using the infrastructure
                provider rather than Kubernetes networking.
              type: string
            apiServerURL:
              description: apiServerURL is a valid URI with scheme(http/https), address
                and port.  apiServerURL can be used by components like the web console
                to tell users where to find the Kubernetes API.
              type: string
            etcdDiscoveryDomain:
              description: 'etcdDiscoveryDomain is the domain used to fetch the SRV
                records for discovering etcd servers and clients. For more info: https://github.com/etcd-io/etcd/blob/329be66e8b3f9e2e6af83c123ff89297e49ebd15/Documentation/op-guide/clustering.md#dns-discovery'
              type: string
            infrastructureName:
              description: infrastructureName uniquely identifies a cluster with a
                human friendly name. Once set it should not be changed. Must be of
                max length 27 and must have only alphanumeric or hyphen characters.
              type: string
            platform:
              description: "platform is the underlying infrastructure provider for
                the cluster. \n Deprecated: Use platformStatus.type instead."
              type: string
            platformStatus:
              description: platformStatus holds status information specific to the
                underlying infrastructure provider.
              properties:
                aws:
                  description: AWS contains settings specific to the Amazon Web Services
                    infrastructure provider.
                  properties:
                    region:
                      description: region holds the default AWS region for new AWS
                        resources created by the cluster.
                      type: string
                  type: object
                azure:
                  description: Azure contains settings specific to the Azure infrastructure
                    provider.
                  properties:
                    resourceGroupName:
                      description: resourceGroupName is the Resource Group for new
                        Azure resources created for the cluster.
                      type: string
                  type: object
                baremetal:
                  description: BareMetal contains settings specific to the BareMetal
                    platform.
                  properties:
                    apiServerInternalIP:
                      description: apiServerInternalIP is an IP address to contact
                        the Kubernetes API server that can be used by components inside
                        the cluster, like kubelets using the infrastructure rather
                        than Kubernetes networking. It is the IP that the Infrastructure.status.apiServerInternalURI
                        points to. It is the IP for a self-hosted load balancer in
                        front of the API servers.
                      type: string
                    ingressIP:
                      description: ingressIP is an external IP which routes to the
                        default ingress controller. The IP is a suitable target of
                        a wildcard DNS record used to resolve default route host names.
                      type: string
                    nodeDNSIP:
                      description: nodeDNSIP is the IP address for the internal DNS
                        used by the nodes. Unlike the one managed by the DNS operator,
                        ` + "`" + `NodeDNSIP` + "`" + ` provides name resolution for the nodes themselves.
                        There is no DNS-as-a-service for BareMetal deployments. In
                        order to minimize necessary changes to the datacenter DNS,
                        a DNS service is hosted as a static pod to serve those hostnames
                        to the nodes in the cluster.
                      type: string
                  type: object
                gcp:
                  description: GCP contains settings specific to the Google Cloud
                    Platform infrastructure provider.
                  properties:
                    projectID:
                      description: resourceGroupName is the Project ID for new GCP
                        resources created for the cluster.
                      type: string
                    region:
                      description: region holds the region for new GCP resources created
                        for the cluster.
                      type: string
                  type: object
                openstack:
                  description: OpenStack contains settings specific to the OpenStack
                    infrastructure provider.
                  properties:
                    apiServerInternalIP:
                      description: apiServerInternalIP is an IP address to contact
                        the Kubernetes API server that can be used by components inside
                        the cluster, like kubelets using the infrastructure rather
                        than Kubernetes networking. It is the IP that the Infrastructure.status.apiServerInternalURI
                        points to. It is the IP for a self-hosted load balancer in
                        front of the API servers.
                      type: string
                    cloudName:
                      description: cloudName is the name of the desired OpenStack
                        cloud in the client configuration file (` + "`" + `clouds.yaml` + "`" + `).
                      type: string
                    ingressIP:
                      description: ingressIP is an external IP which routes to the
                        default ingress controller. The IP is a suitable target of
                        a wildcard DNS record used to resolve default route host names.
                      type: string
                    nodeDNSIP:
                      description: nodeDNSIP is the IP address for the internal DNS
                        used by the nodes. Unlike the one managed by the DNS operator,
                        ` + "`" + `NodeDNSIP` + "`" + ` provides name resolution for the nodes themselves.
                        There is no DNS-as-a-service for OpenStack deployments. In
                        order to minimize necessary changes to the datacenter DNS,
                        a DNS service is hosted as a static pod to serve those hostnames
                        to the nodes in the cluster.
                      type: string
                  type: object
                type:
                  description: type is the underlying infrastructure provider for
                    the cluster. This value controls whether infrastructure automation
                    such as service load balancers, dynamic volume provisioning, machine
                    creation and deletion, and other integrations are enabled. If
                    None, no infrastructure automation is enabled. Allowed values
                    are "AWS", "Azure", "BareMetal", "GCP", "Libvirt", "OpenStack",
                    "VSphere", "oVirt", and "None". Individual components may not
                    support all platforms, and must handle unrecognized platforms
                    as None if they do not support that platform.
                  type: string
              type: object
          type: object
      required:
      - spec
      type: object
`)

func assetsCrd0000_10_configOperator_01_infrastructureCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_infrastructureCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_infrastructureCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_infrastructureCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_infrastructure.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_ingressCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: ingresses.config.openshift.io
spec:
  group: config.openshift.io
  names:
    kind: Ingress
    listKind: IngressList
    plural: ingresses
    singular: ingress
  scope: Cluster
  versions:
  - name: v1
    served: true
    storage: true
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: Ingress holds cluster-wide information about ingress, including
        the default ingress domain used for routes. The canonical name is ` + "`" + `cluster` + "`" + `.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object's metadata.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: spec holds user settable values for configuration
          properties:
            domain:
              description: "domain is used to generate a default host name for a route
                when the route's host name is empty. The generated host name will
                follow this pattern: \"<route-name>.<route-namespace>.<domain>\".
                \n It is also used as the default wildcard domain suffix for ingress.
                The default ingresscontroller domain will follow this pattern: \"*.<domain>\".
                \n Once set, changing domain is not currently supported."
              type: string
          type: object
        status:
          description: status holds observed values from the cluster. They may not
            be overridden.
          type: object
      required:
      - spec
      type: object
`)

func assetsCrd0000_10_configOperator_01_ingressCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_ingressCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_ingressCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_ingressCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_ingress.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_networkCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: networks.config.openshift.io
spec:
  group: config.openshift.io
  names:
    kind: Network
    listKind: NetworkList
    plural: networks
    singular: network
  scope: Cluster
  versions:
  - name: v1
    served: true
    storage: true
  validation:
    openAPIV3Schema:
      description: 'Network holds cluster-wide information about Network. The canonical
        name is ` + "`" + `cluster` + "`" + `. It is used to configure the desired network configuration,
        such as: IP address pools for services/pod IPs, network plugin, etc. Please
        view network.spec for an explanation on what applies when configuring this
        resource.'
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object's metadata.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: spec holds user settable values for configuration. As a general
            rule, this SHOULD NOT be read directly. Instead, you should consume the
            NetworkStatus, as it indicates the currently deployed configuration. Currently,
            most spec fields are immutable after installation. Please view the individual
            ones for further details on each.
          properties:
            clusterNetwork:
              description: IP address pool to use for pod IPs. This field is immutable
                after installation.
              items:
                description: ClusterNetworkEntry is a contiguous block of IP addresses
                  from which pod IPs are allocated.
                properties:
                  cidr:
                    description: The complete block for pod IPs.
                    type: string
                  hostPrefix:
                    description: The size (prefix) of block to allocate to each node.
                    format: int32
                    type: integer
                type: object
              type: array
            externalIP:
              description: externalIP defines configuration for controllers that affect
                Service.ExternalIP
              properties:
                autoAssignCIDRs:
                  description: autoAssignCIDRs is a list of CIDRs from which to automatically
                    assign Service.ExternalIP. These are assigned when the service
                    is of type LoadBalancer. In general, this is only useful for bare-metal
                    clusters. In Openshift 3.x, this was misleadingly called "IngressIPs".
                    Automatically assigned External IPs are not affected by any ExternalIPPolicy
                    rules. Currently, only one entry may be provided.
                  items:
                    type: string
                  type: array
                policy:
                  description: policy is a set of restrictions applied to the ExternalIP
                    field. If nil, any value is allowed for an ExternalIP. If the
                    empty/zero policy is supplied, then ExternalIP is not allowed
                    to be set.
                  properties:
                    allowedCIDRs:
                      description: allowedCIDRs is the list of allowed CIDRs.
                      items:
                        type: string
                      type: array
                    rejectedCIDRs:
                      description: rejectedCIDRs is the list of disallowed CIDRs.
                        These take precedence over allowedCIDRs.
                      items:
                        type: string
                      type: array
                  type: object
              type: object
            networkType:
              description: 'NetworkType is the plugin that is to be deployed (e.g.
                OpenShiftSDN). This should match a value that the cluster-network-operator
                understands, or else no networking will be installed. Currently supported
                values are: - OpenShiftSDN This field is immutable after installation.'
              type: string
            serviceNetwork:
              description: IP address pool for services. Currently, we only support
                a single entry here. This field is immutable after installation.
              items:
                type: string
              type: array
          type: object
        status:
          description: status holds observed values from the cluster. They may not
            be overridden.
          properties:
            clusterNetwork:
              description: IP address pool to use for pod IPs.
              items:
                description: ClusterNetworkEntry is a contiguous block of IP addresses
                  from which pod IPs are allocated.
                properties:
                  cidr:
                    description: The complete block for pod IPs.
                    type: string
                  hostPrefix:
                    description: The size (prefix) of block to allocate to each node.
                    format: int32
                    type: integer
                type: object
              type: array
            clusterNetworkMTU:
              description: ClusterNetworkMTU is the MTU for inter-pod networking.
              type: integer
            networkType:
              description: NetworkType is the plugin that is deployed (e.g. OpenShiftSDN).
              type: string
            serviceNetwork:
              description: IP address pool for services. Currently, we only support
                a single entry here.
              items:
                type: string
              type: array
          type: object
      required:
      - spec
      type: object
`)

func assetsCrd0000_10_configOperator_01_networkCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_networkCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_networkCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_networkCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_network.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_oauthCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: oauths.config.openshift.io
spec:
  group: config.openshift.io
  names:
    kind: OAuth
    listKind: OAuthList
    plural: oauths
    singular: oauth
  scope: Cluster
  subresources:
    status: {}
  versions:
  - name: v1
    served: true
    storage: true
  validation:
    openAPIV3Schema:
      description: OAuth holds cluster-wide information about OAuth.  The canonical
        name is ` + "`" + `cluster` + "`" + `. It is used to configure the integrated OAuth server. This
        configuration is only honored when the top level Authentication config has
        type set to IntegratedOAuth.
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: ObjectMeta is metadata that all persisted resources must have,
            which includes all objects users must create.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: OAuthSpec contains desired cluster auth configuration
          properties:
            identityProviders:
              description: identityProviders is an ordered list of ways for a user
                to identify themselves. When this list is empty, no identities are
                provisioned for users.
              items:
                description: IdentityProvider provides identities for users authenticating
                  using credentials
                properties:
                  basicAuth:
                    description: basicAuth contains configuration options for the
                      BasicAuth IdP
                    properties:
                      ca:
                        description: ca is an optional reference to a config map by
                          name containing the PEM-encoded CA bundle. It is used as
                          a trust anchor to validate the TLS certificate presented
                          by the remote server. The key "ca.crt" is used to locate
                          the data. If specified and the config map or expected key
                          is not found, the identity provider is not honored. If the
                          specified ca data is not valid, the identity provider is
                          not honored. If empty, the default system roots are used.
                          The namespace for this config map is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              config map
                            type: string
                        required:
                        - name
                        type: object
                      tlsClientCert:
                        description: tlsClientCert is an optional reference to a secret
                          by name that contains the PEM-encoded TLS client certificate
                          to present when connecting to the server. The key "tls.crt"
                          is used to locate the data. If specified and the secret
                          or expected key is not found, the identity provider is not
                          honored. If the specified certificate data is not valid,
                          the identity provider is not honored. The namespace for
                          this secret is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              secret
                            type: string
                        required:
                        - name
                        type: object
                      tlsClientKey:
                        description: tlsClientKey is an optional reference to a secret
                          by name that contains the PEM-encoded TLS private key for
                          the client certificate referenced in tlsClientCert. The
                          key "tls.key" is used to locate the data. If specified and
                          the secret or expected key is not found, the identity provider
                          is not honored. If the specified certificate data is not
                          valid, the identity provider is not honored. The namespace
                          for this secret is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              secret
                            type: string
                        required:
                        - name
                        type: object
                      url:
                        description: url is the remote URL to connect to
                        type: string
                    type: object
                  github:
                    description: github enables user authentication using GitHub credentials
                    properties:
                      ca:
                        description: ca is an optional reference to a config map by
                          name containing the PEM-encoded CA bundle. It is used as
                          a trust anchor to validate the TLS certificate presented
                          by the remote server. The key "ca.crt" is used to locate
                          the data. If specified and the config map or expected key
                          is not found, the identity provider is not honored. If the
                          specified ca data is not valid, the identity provider is
                          not honored. If empty, the default system roots are used.
                          This can only be configured when hostname is set to a non-empty
                          value. The namespace for this config map is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              config map
                            type: string
                        required:
                        - name
                        type: object
                      clientID:
                        description: clientID is the oauth client ID
                        type: string
                      clientSecret:
                        description: clientSecret is a required reference to the secret
                          by name containing the oauth client secret. The key "clientSecret"
                          is used to locate the data. If the secret or expected key
                          is not found, the identity provider is not honored. The
                          namespace for this secret is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              secret
                            type: string
                        required:
                        - name
                        type: object
                      hostname:
                        description: hostname is the optional domain (e.g. "mycompany.com")
                          for use with a hosted instance of GitHub Enterprise. It
                          must match the GitHub Enterprise settings value configured
                          at /setup/settings#hostname.
                        type: string
                      organizations:
                        description: organizations optionally restricts which organizations
                          are allowed to log in
                        items:
                          type: string
                        type: array
                      teams:
                        description: teams optionally restricts which teams are allowed
                          to log in. Format is <org>/<team>.
                        items:
                          type: string
                        type: array
                    type: object
                  gitlab:
                    description: gitlab enables user authentication using GitLab credentials
                    properties:
                      ca:
                        description: ca is an optional reference to a config map by
                          name containing the PEM-encoded CA bundle. It is used as
                          a trust anchor to validate the TLS certificate presented
                          by the remote server. The key "ca.crt" is used to locate
                          the data. If specified and the config map or expected key
                          is not found, the identity provider is not honored. If the
                          specified ca data is not valid, the identity provider is
                          not honored. If empty, the default system roots are used.
                          The namespace for this config map is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              config map
                            type: string
                        required:
                        - name
                        type: object
                      clientID:
                        description: clientID is the oauth client ID
                        type: string
                      clientSecret:
                        description: clientSecret is a required reference to the secret
                          by name containing the oauth client secret. The key "clientSecret"
                          is used to locate the data. If the secret or expected key
                          is not found, the identity provider is not honored. The
                          namespace for this secret is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              secret
                            type: string
                        required:
                        - name
                        type: object
                      url:
                        description: url is the oauth server base URL
                        type: string
                    type: object
                  google:
                    description: google enables user authentication using Google credentials
                    properties:
                      clientID:
                        description: clientID is the oauth client ID
                        type: string
                      clientSecret:
                        description: clientSecret is a required reference to the secret
                          by name containing the oauth client secret. The key "clientSecret"
                          is used to locate the data. If the secret or expected key
                          is not found, the identity provider is not honored. The
                          namespace for this secret is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              secret
                            type: string
                        required:
                        - name
                        type: object
                      hostedDomain:
                        description: hostedDomain is the optional Google App domain
                          (e.g. "mycompany.com") to restrict logins to
                        type: string
                    type: object
                  htpasswd:
                    description: htpasswd enables user authentication using an HTPasswd
                      file to validate credentials
                    properties:
                      fileData:
                        description: fileData is a required reference to a secret
                          by name containing the data to use as the htpasswd file.
                          The key "htpasswd" is used to locate the data. If the secret
                          or expected key is not found, the identity provider is not
                          honored. If the specified htpasswd data is not valid, the
                          identity provider is not honored. The namespace for this
                          secret is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              secret
                            type: string
                        required:
                        - name
                        type: object
                    type: object
                  keystone:
                    description: keystone enables user authentication using keystone
                      password credentials
                    properties:
                      ca:
                        description: ca is an optional reference to a config map by
                          name containing the PEM-encoded CA bundle. It is used as
                          a trust anchor to validate the TLS certificate presented
                          by the remote server. The key "ca.crt" is used to locate
                          the data. If specified and the config map or expected key
                          is not found, the identity provider is not honored. If the
                          specified ca data is not valid, the identity provider is
                          not honored. If empty, the default system roots are used.
                          The namespace for this config map is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              config map
                            type: string
                        required:
                        - name
                        type: object
                      domainName:
                        description: domainName is required for keystone v3
                        type: string
                      tlsClientCert:
                        description: tlsClientCert is an optional reference to a secret
                          by name that contains the PEM-encoded TLS client certificate
                          to present when connecting to the server. The key "tls.crt"
                          is used to locate the data. If specified and the secret
                          or expected key is not found, the identity provider is not
                          honored. If the specified certificate data is not valid,
                          the identity provider is not honored. The namespace for
                          this secret is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              secret
                            type: string
                        required:
                        - name
                        type: object
                      tlsClientKey:
                        description: tlsClientKey is an optional reference to a secret
                          by name that contains the PEM-encoded TLS private key for
                          the client certificate referenced in tlsClientCert. The
                          key "tls.key" is used to locate the data. If specified and
                          the secret or expected key is not found, the identity provider
                          is not honored. If the specified certificate data is not
                          valid, the identity provider is not honored. The namespace
                          for this secret is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              secret
                            type: string
                        required:
                        - name
                        type: object
                      url:
                        description: url is the remote URL to connect to
                        type: string
                    type: object
                  ldap:
                    description: ldap enables user authentication using LDAP credentials
                    properties:
                      attributes:
                        description: attributes maps LDAP attributes to identities
                        properties:
                          email:
                            description: email is the list of attributes whose values
                              should be used as the email address. Optional. If unspecified,
                              no email is set for the identity
                            items:
                              type: string
                            type: array
                          id:
                            description: id is the list of attributes whose values
                              should be used as the user ID. Required. First non-empty
                              attribute is used. At least one attribute is required.
                              If none of the listed attribute have a value, authentication
                              fails. LDAP standard identity attribute is "dn"
                            items:
                              type: string
                            type: array
                          name:
                            description: name is the list of attributes whose values
                              should be used as the display name. Optional. If unspecified,
                              no display name is set for the identity LDAP standard
                              display name attribute is "cn"
                            items:
                              type: string
                            type: array
                          preferredUsername:
                            description: preferredUsername is the list of attributes
                              whose values should be used as the preferred username.
                              LDAP standard login attribute is "uid"
                            items:
                              type: string
                            type: array
                        type: object
                      bindDN:
                        description: bindDN is an optional DN to bind with during
                          the search phase.
                        type: string
                      bindPassword:
                        description: bindPassword is an optional reference to a secret
                          by name containing a password to bind with during the search
                          phase. The key "bindPassword" is used to locate the data.
                          If specified and the secret or expected key is not found,
                          the identity provider is not honored. The namespace for
                          this secret is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              secret
                            type: string
                        required:
                        - name
                        type: object
                      ca:
                        description: ca is an optional reference to a config map by
                          name containing the PEM-encoded CA bundle. It is used as
                          a trust anchor to validate the TLS certificate presented
                          by the remote server. The key "ca.crt" is used to locate
                          the data. If specified and the config map or expected key
                          is not found, the identity provider is not honored. If the
                          specified ca data is not valid, the identity provider is
                          not honored. If empty, the default system roots are used.
                          The namespace for this config map is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              config map
                            type: string
                        required:
                        - name
                        type: object
                      insecure:
                        description: 'insecure, if true, indicates the connection
                          should not use TLS WARNING: Should not be set to ` + "`" + `true` + "`" + `
                          with the URL scheme "ldaps://" as "ldaps://" URLs always          attempt
                          to connect using TLS, even when ` + "`" + `insecure` + "`" + ` is set to ` + "`" + `true` + "`" + `
                          When ` + "`" + `true` + "`" + `, "ldap://" URLS connect insecurely. When ` + "`" + `false` + "`" + `,
                          "ldap://" URLs are upgraded to a TLS connection using StartTLS
                          as specified in https://tools.ietf.org/html/rfc2830.'
                        type: boolean
                      url:
                        description: 'url is an RFC 2255 URL which specifies the LDAP
                          search parameters to use. The syntax of the URL is: ldap://host:port/basedn?attribute?scope?filter'
                        type: string
                    type: object
                  mappingMethod:
                    description: mappingMethod determines how identities from this
                      provider are mapped to users Defaults to "claim"
                    type: string
                  name:
                    description: 'name is used to qualify the identities returned
                      by this provider. - It MUST be unique and not shared by any
                      other identity provider used - It MUST be a valid path segment:
                      name cannot equal "." or ".." or contain "/" or "%" or ":"   Ref:
                      https://godoc.org/github.com/openshift/origin/pkg/user/apis/user/validation#ValidateIdentityProviderName'
                    type: string
                  openID:
                    description: openID enables user authentication using OpenID credentials
                    properties:
                      ca:
                        description: ca is an optional reference to a config map by
                          name containing the PEM-encoded CA bundle. It is used as
                          a trust anchor to validate the TLS certificate presented
                          by the remote server. The key "ca.crt" is used to locate
                          the data. If specified and the config map or expected key
                          is not found, the identity provider is not honored. If the
                          specified ca data is not valid, the identity provider is
                          not honored. If empty, the default system roots are used.
                          The namespace for this config map is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              config map
                            type: string
                        required:
                        - name
                        type: object
                      claims:
                        description: claims mappings
                        properties:
                          email:
                            description: email is the list of claims whose values
                              should be used as the email address. Optional. If unspecified,
                              no email is set for the identity
                            items:
                              type: string
                            type: array
                          name:
                            description: name is the list of claims whose values should
                              be used as the display name. Optional. If unspecified,
                              no display name is set for the identity
                            items:
                              type: string
                            type: array
                          preferredUsername:
                            description: preferredUsername is the list of claims whose
                              values should be used as the preferred username. If
                              unspecified, the preferred username is determined from
                              the value of the sub claim
                            items:
                              type: string
                            type: array
                        type: object
                      clientID:
                        description: clientID is the oauth client ID
                        type: string
                      clientSecret:
                        description: clientSecret is a required reference to the secret
                          by name containing the oauth client secret. The key "clientSecret"
                          is used to locate the data. If the secret or expected key
                          is not found, the identity provider is not honored. The
                          namespace for this secret is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              secret
                            type: string
                        required:
                        - name
                        type: object
                      extraAuthorizeParameters:
                        additionalProperties:
                          type: string
                        description: extraAuthorizeParameters are any custom parameters
                          to add to the authorize request.
                        type: object
                      extraScopes:
                        description: extraScopes are any scopes to request in addition
                          to the standard "openid" scope.
                        items:
                          type: string
                        type: array
                      issuer:
                        description: issuer is the URL that the OpenID Provider asserts
                          as its Issuer Identifier. It must use the https scheme with
                          no query or fragment component.
                        type: string
                    type: object
                  requestHeader:
                    description: requestHeader enables user authentication using request
                      header credentials
                    properties:
                      ca:
                        description: ca is a required reference to a config map by
                          name containing the PEM-encoded CA bundle. It is used as
                          a trust anchor to validate the TLS certificate presented
                          by the remote server. Specifically, it allows verification
                          of incoming requests to prevent header spoofing. The key
                          "ca.crt" is used to locate the data. If the config map or
                          expected key is not found, the identity provider is not
                          honored. If the specified ca data is not valid, the identity
                          provider is not honored. The namespace for this config map
                          is openshift-config.
                        properties:
                          name:
                            description: name is the metadata.name of the referenced
                              config map
                            type: string
                        required:
                        - name
                        type: object
                      challengeURL:
                        description: challengeURL is a URL to redirect unauthenticated
                          /authorize requests to Unauthenticated requests from OAuth
                          clients which expect WWW-Authenticate challenges will be
                          redirected here. ${url} is replaced with the current URL,
                          escaped to be safe in a query parameter   https://www.example.com/sso-login?then=${url}
                          ${query} is replaced with the current query string   https://www.example.com/auth-proxy/oauth/authorize?${query}
                          Required when challenge is set to true.
                        type: string
                      clientCommonNames:
                        description: clientCommonNames is an optional list of common
                          names to require a match from. If empty, any client certificate
                          validated against the clientCA bundle is considered authoritative.
                        items:
                          type: string
                        type: array
                      emailHeaders:
                        description: emailHeaders is the set of headers to check for
                          the email address
                        items:
                          type: string
                        type: array
                      headers:
                        description: headers is the set of headers to check for identity
                          information
                        items:
                          type: string
                        type: array
                      loginURL:
                        description: loginURL is a URL to redirect unauthenticated
                          /authorize requests to Unauthenticated requests from OAuth
                          clients which expect interactive logins will be redirected
                          here ${url} is replaced with the current URL, escaped to
                          be safe in a query parameter   https://www.example.com/sso-login?then=${url}
                          ${query} is replaced with the current query string   https://www.example.com/auth-proxy/oauth/authorize?${query}
                          Required when login is set to true.
                        type: string
                      nameHeaders:
                        description: nameHeaders is the set of headers to check for
                          the display name
                        items:
                          type: string
                        type: array
                      preferredUsernameHeaders:
                        description: preferredUsernameHeaders is the set of headers
                          to check for the preferred username
                        items:
                          type: string
                        type: array
                    type: object
                  type:
                    description: type identifies the identity provider type for this
                      entry.
                    type: string
                type: object
              type: array
            templates:
              description: templates allow you to customize pages like the login page.
              properties:
                error:
                  description: error is the name of a secret that specifies a go template
                    to use to render error pages during the authentication or grant
                    flow. The key "errors.html" is used to locate the template data.
                    If specified and the secret or expected key is not found, the
                    default error page is used. If the specified template is not valid,
                    the default error page is used. If unspecified, the default error
                    page is used. The namespace for this secret is openshift-config.
                  properties:
                    name:
                      description: name is the metadata.name of the referenced secret
                      type: string
                  required:
                  - name
                  type: object
                login:
                  description: login is the name of a secret that specifies a go template
                    to use to render the login page. The key "login.html" is used
                    to locate the template data. If specified and the secret or expected
                    key is not found, the default login page is used. If the specified
                    template is not valid, the default login page is used. If unspecified,
                    the default login page is used. The namespace for this secret
                    is openshift-config.
                  properties:
                    name:
                      description: name is the metadata.name of the referenced secret
                      type: string
                  required:
                  - name
                  type: object
                providerSelection:
                  description: providerSelection is the name of a secret that specifies
                    a go template to use to render the provider selection page. The
                    key "providers.html" is used to locate the template data. If specified
                    and the secret or expected key is not found, the default provider
                    selection page is used. If the specified template is not valid,
                    the default provider selection page is used. If unspecified, the
                    default provider selection page is used. The namespace for this
                    secret is openshift-config.
                  properties:
                    name:
                      description: name is the metadata.name of the referenced secret
                      type: string
                  required:
                  - name
                  type: object
              type: object
            tokenConfig:
              description: tokenConfig contains options for authorization and access
                tokens
              properties:
                accessTokenInactivityTimeoutSeconds:
                  description: 'accessTokenInactivityTimeoutSeconds defines the default
                    token inactivity timeout for tokens granted by any client. The
                    value represents the maximum amount of time that can occur between
                    consecutive uses of the token. Tokens become invalid if they are
                    not used within this temporal window. The user will need to acquire
                    a new token to regain access once a token times out. Valid values
                    are integer values:   x < 0  Tokens time out is enabled but tokens
                    never timeout unless configured per client (e.g. ` + "`" + `-1` + "`" + `)   x = 0  Tokens
                    time out is disabled (default)   x > 0  Tokens time out if there
                    is no activity for x seconds The current minimum allowed value
                    for X is 300 (5 minutes)'
                  format: int32
                  type: integer
                accessTokenMaxAgeSeconds:
                  description: accessTokenMaxAgeSeconds defines the maximum age of
                    access tokens
                  format: int32
                  type: integer
              type: object
          type: object
        status:
          description: OAuthStatus shows current known state of OAuth server in the
            cluster
          type: object
      required:
      - spec
      type: object
`)

func assetsCrd0000_10_configOperator_01_oauthCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_oauthCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_oauthCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_oauthCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_oauth.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_10_configOperator_01_projectCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: projects.config.openshift.io
spec:
  group: config.openshift.io
  scope: Cluster
  versions:
  - name: v1
    served: true
    storage: true
  names:
    kind: Project
    listKind: ProjectList
    plural: projects
    singular: project
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: Project holds cluster-wide information about Project.  The canonical
        name is ` + "`" + `cluster` + "`" + `
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: Standard object's metadata.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: spec holds user settable values for configuration
          properties:
            projectRequestMessage:
              description: projectRequestMessage is the string presented to a user
                if they are unable to request a project via the projectrequest api
                endpoint
              type: string
            projectRequestTemplate:
              description: projectRequestTemplate is the template to use for creating
                projects in response to projectrequest. This must point to a template
                in 'openshift-config' namespace. It is optional. If it is not specified,
                a default template is used.
              properties:
                name:
                  description: name is the metadata.name of the referenced project
                    request template
                  type: string
              type: object
          type: object
        status:
          description: status holds observed values from the cluster. They may not
            be overridden.
          type: object
      required:
      - spec
      type: object
`)

func assetsCrd0000_10_configOperator_01_projectCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_10_configOperator_01_projectCrdYaml, nil
}

func assetsCrd0000_10_configOperator_01_projectCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_10_configOperator_01_projectCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_10_config-operator_01_project.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_11_imageregistryConfigsCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: configs.imageregistry.operator.openshift.io
spec:
  conversion:
    strategy: None
  group: imageregistry.operator.openshift.io
  names:
    kind: Config
    listKind: ConfigList
    plural: configs
    singular: config
  scope: Cluster
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: Config is the configuration object for a registry instance managed
        by the registry operator
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          properties:
            defaultRoute:
              description: DefaultRoute indicates whether an external facing route
                for the registry should be created using the default generated hostname.
              type: boolean
            httpSecret:
              description: HTTPSecret is the value needed by the registry to secure
                uploads, generated by default.
              type: string
            logging:
              description: LogLevel determines the level of logging enabled in the
                registry.
              format: int64
              type: integer
            managementState:
              description: ManagementState indicates whether the registry instance
                represented by this config instance is under operator management or
                not.  Valid values are Managed, Unmanaged, and Removed.
              pattern: ^(Managed|Unmanaged|Force|Removed)$
              type: string
            nodeSelector:
              additionalProperties:
                type: string
              description: NodeSelector defines the node selection constraints for
                the registry pod.
              type: object
            proxy:
              description: Proxy defines the proxy to be used when calling master
                api, upstream registries, etc.
              properties:
                http:
                  type: string
                https:
                  type: string
                noProxy:
                  type: string
              type: object
            readOnly:
              description: ReadOnly indicates whether the registry instance should
                reject attempts to push new images or delete existing ones.
              type: boolean
            replicas:
              description: Replicas determines the number of registry instances to
                run.
              format: int32
              type: integer
            requests:
              description: Requests controls how many parallel requests a given registry
                instance will handle before queuing additional requests.
              properties:
                read:
                  properties:
                    maxInQueue:
                      description: MaxInQueue sets the maximum queued api requests
                        to the registry.
                      type: integer
                    maxRunning:
                      description: MaxRunning sets the maximum in flight api requests
                        to the registry.
                      type: integer
                    maxWaitInQueue:
                      description: MaxWaitInQueue sets the maximum time a request
                        can wait in the queue before being rejected.
                      type: string
                  type: object
                write:
                  properties:
                    maxInQueue:
                      description: MaxInQueue sets the maximum queued api requests
                        to the registry.
                      type: integer
                    maxRunning:
                      description: MaxRunning sets the maximum in flight api requests
                        to the registry.
                      type: integer
                    maxWaitInQueue:
                      description: MaxWaitInQueue sets the maximum time a request
                        can wait in the queue before being rejected.
                      type: string
                  type: object
              type: object
            resources:
              description: Resources defines the resource requests+limits for the
                registry pod.
              properties:
                limits:
                  additionalProperties:
                    type: string
                  description: 'Limits describes the maximum amount of compute resources
                    allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                  type: object
                requests:
                  additionalProperties:
                    type: string
                  description: 'Requests describes the minimum amount of compute resources
                    required. If Requests is omitted for a container, it defaults
                    to Limits if that is explicitly specified, otherwise to an implementation-defined
                    value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                  type: object
              type: object
            routes:
              description: Routes defines additional external facing routes which
                should be created for the registry.
              items:
                properties:
                  hostname:
                    description: Hostname for the route.
                    type: string
                  name:
                    description: Name of the route to be created.
                    type: string
                  secretName:
                    description: SecretName points to secret containing the certificates
                      to be used by the route.
                    type: string
                required:
                - name
                type: object
              type: array
            storage:
              description: Storage details for configuring registry storage, e.g.
                S3 bucket coordinates.
              properties:
                azure:
                  description: Azure represents configuration that uses Azure Blob
                    Storage.
                  properties:
                    accountName:
                      type: string
                    container:
                      type: string
                  type: object
                emptyDir:
                  description: EmptyDir represents ephemeral storage on the pod's
                    host node. This storage cannot be used with more than 1 replica
                    and is not suitable for production use. When the pod is removed
                    from a node for any reason, the data in the emptyDir is deleted
                    forever. This configuration is EXPERIMENTAL and is subject to
                    change without notice.
                  type: object
                gcs:
                  description: GCS represents configuration that uses Google Cloud
                    Storage.
                  properties:
                    bucket:
                      description: Bucket is the bucket name in which you want to
                        store the registry's data. Optional, will be generated if
                        not provided.
                      type: string
                    keyID:
                      description: KeyID is the KMS key ID to use for encryption.
                        Optional, buckets are encrypted by default on GCP. This allows
                        for the use of a custom encryption key.
                      type: string
                    projectID:
                      description: ProjectID is the Project ID of the GCP project
                        that this bucket should be associated with.
                      type: string
                    region:
                      description: Region is the GCS location in which your bucket
                        exists. Optional, will be set based on the installed GCS Region.
                      type: string
                  type: object
                pvc:
                  description: PVC represents configuration that uses a PersistentVolumeClaim.
                  properties:
                    claim:
                      type: string
                  type: object
                s3:
                  description: S3 represents configuration that uses Amazon Simple
                    Storage Service.
                  properties:
                    bucket:
                      description: Bucket is the bucket name in which you want to
                        store the registry's data. Optional, will be generated if
                        not provided.
                      type: string
                    cloudFront:
                      description: CloudFront configures Amazon Cloudfront as the
                        storage middleware in a registry.
                      properties:
                        baseURL:
                          description: BaseURL contains the SCHEME://HOST[/PATH] at
                            which Cloudfront is served.
                          type: string
                        duration:
                          description: Duration is the duration of the Cloudfront
                            session.
                          type: string
                        keypairID:
                          description: KeypairID is key pair ID provided by AWS.
                          type: string
                        privateKey:
                          description: PrivateKey points to secret containing the
                            private key, provided by AWS.
                          properties:
                            key:
                              description: The key of the secret to select from.  Must
                                be a valid secret key.
                              type: string
                            name:
                              description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                TODO: Add other useful fields. apiVersion, kind, uid?'
                              type: string
                            optional:
                              description: Specify whether the Secret or it's key
                                must be defined
                              type: boolean
                          required:
                          - key
                          type: object
                      required:
                      - baseURL
                      - keypairID
                      - privateKey
                      type: object
                    encrypt:
                      description: Encrypt specifies whether the registry stores the
                        image in encrypted format or not. Optional, defaults to false.
                      type: boolean
                    keyID:
                      description: KeyID is the KMS key ID to use for encryption.
                        Optional, Encrypt must be true, or this parameter is ignored.
                      type: string
                    region:
                      description: Region is the AWS region in which your bucket exists.
                        Optional, will be set based on the installed AWS Region.
                      type: string
                    regionEndpoint:
                      description: RegionEndpoint is the endpoint for S3 compatible
                        storage services. Optional, defaults based on the Region that
                        is provided.
                      type: string
                  type: object
                swift:
                  description: Swift represents configuration that uses OpenStack
                    Object Storage. This configuration is EXPERIMENTAL and is subject
                    to change without notice.
                  properties:
                    authURL:
                      type: string
                    authVersion:
                      type: string
                    container:
                      type: string
                    domain:
                      type: string
                    domainID:
                      type: string
                    regionName:
                      type: string
                    tenant:
                      type: string
                    tenantID:
                      type: string
                  type: object
              type: object
            tolerations:
              description: Tolerations defines the tolerations for the registry pod.
              items:
                description: The pod this Toleration is attached to tolerates any
                  taint that matches the triple <key,value,effect> using the matching
                  operator <operator>.
                properties:
                  effect:
                    description: Effect indicates the taint effect to match. Empty
                      means match all taint effects. When specified, allowed values
                      are NoSchedule, PreferNoSchedule and NoExecute.
                    type: string
                  key:
                    description: Key is the taint key that the toleration applies
                      to. Empty means match all taint keys. If the key is empty, operator
                      must be Exists; this combination means to match all values and
                      all keys.
                    type: string
                  operator:
                    description: Operator represents a key's relationship to the value.
                      Valid operators are Exists and Equal. Defaults to Equal. Exists
                      is equivalent to wildcard for value, so that a pod can tolerate
                      all taints of a particular category.
                    type: string
                  tolerationSeconds:
                    description: TolerationSeconds represents the period of time the
                      toleration (which must be of effect NoExecute, otherwise this
                      field is ignored) tolerates the taint. By default, it is not
                      set, which means tolerate the taint forever (do not evict).
                      Zero and negative values will be treated as 0 (evict immediately)
                      by the system.
                    format: int64
                    type: integer
                  value:
                    description: Value is the taint value the toleration matches to.
                      If the operator is Exists, the value should be empty, otherwise
                      just a regular string.
                    type: string
                type: object
              type: array
          required:
          - logging
          - managementState
          - replicas
          type: object
        status:
          properties:
            conditions:
              description: conditions is a list of conditions and their status
              items:
                description: OperatorCondition is just the standard condition fields.
                properties:
                  lastTransitionTime:
                    format: date-time
                    type: string
                  message:
                    type: string
                  reason:
                    type: string
                  status:
                    type: string
                  type:
                    type: string
                type: object
              type: array
            generations:
              description: generations are used to determine when an item needs to
                be reconciled or has changed in a way that needs a reaction.
              items:
                description: GenerationStatus keeps track of the generation for a
                  given resource so that decisions about forced updates can be made.
                properties:
                  group:
                    description: group is the group of the thing you're tracking
                    type: string
                  hash:
                    description: hash is an optional field set for resources without
                      generation that are content sensitive like secrets and configmaps
                    type: string
                  lastGeneration:
                    description: lastGeneration is the last generation of the workload
                      controller involved
                    format: int64
                    type: integer
                  name:
                    description: name is the name of the thing you're tracking
                    type: string
                  namespace:
                    description: namespace is where the thing you're tracking is
                    type: string
                  resource:
                    description: resource is the resource type of the thing you're
                      tracking
                    type: string
                type: object
              type: array
            observedGeneration:
              description: observedGeneration is the last generation change you've
                dealt with
              format: int64
              type: integer
            readyReplicas:
              description: readyReplicas indicates how many replicas are ready and
                at the desired state
              format: int32
              type: integer
            storage:
              description: Storage indicates the current applied storage configuration
                of the registry.
              properties:
                azure:
                  description: Azure represents configuration that uses Azure Blob
                    Storage.
                  properties:
                    accountName:
                      type: string
                    container:
                      type: string
                  type: object
                emptyDir:
                  description: EmptyDir represents ephemeral storage on the pod's
                    host node. This storage cannot be used with more than 1 replica
                    and is not suitable for production use. When the pod is removed
                    from a node for any reason, the data in the emptyDir is deleted
                    forever. This configuration is EXPERIMENTAL and is subject to
                    change without notice.
                  type: object
                gcs:
                  description: GCS represents configuration that uses Google Cloud
                    Storage.
                  properties:
                    bucket:
                      description: Bucket is the bucket name in which you want to
                        store the registry's data. Optional, will be generated if
                        not provided.
                      type: string
                    keyID:
                      description: KeyID is the KMS key ID to use for encryption.
                        Optional, buckets are encrypted by default on GCP. This allows
                        for the use of a custom encryption key.
                      type: string
                    projectID:
                      description: ProjectID is the Project ID of the GCP project
                        that this bucket should be associated with.
                      type: string
                    region:
                      description: Region is the GCS location in which your bucket
                        exists. Optional, will be set based on the installed GCS Region.
                      type: string
                  type: object
                pvc:
                  description: PVC represents configuration that uses a PersistentVolumeClaim.
                  properties:
                    claim:
                      type: string
                  type: object
                s3:
                  description: S3 represents configuration that uses Amazon Simple
                    Storage Service.
                  properties:
                    bucket:
                      description: Bucket is the bucket name in which you want to
                        store the registry's data. Optional, will be generated if
                        not provided.
                      type: string
                    cloudFront:
                      description: CloudFront configures Amazon Cloudfront as the
                        storage middleware in a registry.
                      properties:
                        baseURL:
                          description: BaseURL contains the SCHEME://HOST[/PATH] at
                            which Cloudfront is served.
                          type: string
                        duration:
                          description: Duration is the duration of the Cloudfront
                            session.
                          type: string
                        keypairID:
                          description: KeypairID is key pair ID provided by AWS.
                          type: string
                        privateKey:
                          description: PrivateKey points to secret containing the
                            private key, provided by AWS.
                          properties:
                            key:
                              description: The key of the secret to select from.  Must
                                be a valid secret key.
                              type: string
                            name:
                              description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                                TODO: Add other useful fields. apiVersion, kind, uid?'
                              type: string
                            optional:
                              description: Specify whether the Secret or it's key
                                must be defined
                              type: boolean
                          required:
                          - key
                          type: object
                      required:
                      - baseURL
                      - keypairID
                      - privateKey
                      type: object
                    encrypt:
                      description: Encrypt specifies whether the registry stores the
                        image in encrypted format or not. Optional, defaults to false.
                      type: boolean
                    keyID:
                      description: KeyID is the KMS key ID to use for encryption.
                        Optional, Encrypt must be true, or this parameter is ignored.
                      type: string
                    region:
                      description: Region is the AWS region in which your bucket exists.
                        Optional, will be set based on the installed AWS Region.
                      type: string
                    regionEndpoint:
                      description: RegionEndpoint is the endpoint for S3 compatible
                        storage services. Optional, defaults based on the Region that
                        is provided.
                      type: string
                  type: object
                swift:
                  description: Swift represents configuration that uses OpenStack
                    Object Storage. This configuration is EXPERIMENTAL and is subject
                    to change without notice.
                  properties:
                    authURL:
                      type: string
                    authVersion:
                      type: string
                    container:
                      type: string
                    domain:
                      type: string
                    domainID:
                      type: string
                    regionName:
                      type: string
                    tenant:
                      type: string
                    tenantID:
                      type: string
                  type: object
              type: object
            storageManaged:
              description: StorageManaged is a boolean which denotes whether or not
                we created the registry storage medium (such as an S3 bucket).
              type: boolean
            version:
              description: version is the level this availability applies to
              type: string
          required:
          - storage
          - storageManaged
          type: object
      required:
      - metadata
      - spec
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
  storedVersions:
  - v1
`)

func assetsCrd0000_11_imageregistryConfigsCrdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_11_imageregistryConfigsCrdYaml, nil
}

func assetsCrd0000_11_imageregistryConfigsCrdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_11_imageregistryConfigsCrdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_11_imageregistry-configs.crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_50_serviceCaOperator_02_crdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: servicecas.operator.openshift.io
spec:
  scope: Cluster
  group: operator.openshift.io
  version: v1
  names:
    kind: ServiceCA
    plural: servicecas
    singular: serviceca
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: ServiceCA provides information to configure an operator to manage
        the service cert controllers
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          description: ObjectMeta is metadata that all persisted resources must have,
            which includes all objects users must create.
          properties:
            annotations:
              additionalProperties:
                type: string
              description: 'Annotations is an unstructured key value map stored with
                a resource that may be set by external tools to store and retrieve
                arbitrary metadata. They are not queryable and should be preserved
                when modifying objects. More info: http://kubernetes.io/docs/user-guide/annotations'
              type: object
            clusterName:
              description: The name of the cluster which the object belongs to. This
                is used to distinguish resources with same name and namespace in different
                clusters. This field is not set anywhere right now and apiserver is
                going to ignore it if set in create or update request.
              type: string
            creationTimestamp:
              description: "CreationTimestamp is a timestamp representing the server
                time when this object was created. It is not guaranteed to be set
                in happens-before order across separate operations. Clients may not
                set this value. It is represented in RFC3339 form and is in UTC. \n
                Populated by the system. Read-only. Null for lists. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            deletionGracePeriodSeconds:
              description: Number of seconds allowed for this object to gracefully
                terminate before it will be removed from the system. Only set when
                deletionTimestamp is also set. May only be shortened. Read-only.
              format: int64
              type: integer
            deletionTimestamp:
              description: "DeletionTimestamp is RFC 3339 date and time at which this
                resource will be deleted. This field is set by the server when a graceful
                deletion is requested by the user, and is not directly settable by
                a client. The resource is expected to be deleted (no longer visible
                from resource lists, and not reachable by name) after the time in
                this field, once the finalizers list is empty. As long as the finalizers
                list contains items, deletion is blocked. Once the deletionTimestamp
                is set, this value may not be unset or be set further into the future,
                although it may be shortened or the resource may be deleted prior
                to this time. For example, a user may request that a pod is deleted
                in 30 seconds. The Kubelet will react by sending a graceful termination
                signal to the containers in the pod. After that 30 seconds, the Kubelet
                will send a hard termination signal (SIGKILL) to the container and
                after cleanup, remove the pod from the API. In the presence of network
                partitions, this object may still exist after this timestamp, until
                an administrator or automated process can determine the resource is
                fully terminated. If not set, graceful deletion of the object has
                not been requested. \n Populated by the system when a graceful deletion
                is requested. Read-only. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#metadata"
              format: date-time
              type: string
            finalizers:
              description: Must be empty before the object is deleted from the registry.
                Each entry is an identifier for the responsible component that will
                remove the entry from the list. If the deletionTimestamp of the object
                is non-nil, entries in this list can only be removed.
              items:
                type: string
              type: array
            generateName:
              description: "GenerateName is an optional prefix, used by the server,
                to generate a unique name ONLY IF the Name field has not been provided.
                If this field is used, the name returned to the client will be different
                than the name passed. This value will also be combined with a unique
                suffix. The provided value has the same validation rules as the Name
                field, and may be truncated by the length of the suffix required to
                make the value unique on the server. \n If this field is specified
                and the generated name exists, the server will NOT return a 409 -
                instead, it will either return 201 Created or 500 with Reason ServerTimeout
                indicating a unique name could not be found in the time allotted,
                and the client should retry (optionally after the time indicated in
                the Retry-After header). \n Applied only if Name is not specified.
                More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#idempotency"
              type: string
            generation:
              description: A sequence number representing a specific generation of
                the desired state. Populated by the system. Read-only.
              format: int64
              type: integer
            initializers:
              description: "An initializer is a controller which enforces some system
                invariant at object creation time. This field is a list of initializers
                that have not yet acted on this object. If nil or empty, this object
                has been completely initialized. Otherwise, the object is considered
                uninitialized and is hidden (in list/watch and get calls) from clients
                that haven't explicitly asked to observe uninitialized objects. \n
                When an object is created, the system will populate this list with
                the current set of initializers. Only privileged users may set or
                modify this list. Once it is empty, it may not be modified further
                by any user. \n DEPRECATED - initializers are an alpha field and will
                be removed in v1.15."
              properties:
                pending:
                  description: Pending is a list of initializers that must execute
                    in order before this object is visible. When the last pending
                    initializer is removed, and no failing result is set, the initializers
                    struct will be set to nil and the object is considered as initialized
                    and visible to all clients.
                  items:
                    description: Initializer is information about an initializer that
                      has not yet completed.
                    properties:
                      name:
                        description: name of the process that is responsible for initializing
                          this object.
                        type: string
                    required:
                    - name
                    type: object
                  type: array
                result:
                  description: If result is set with the Failure field, the object
                    will be persisted to storage and then deleted, ensuring that other
                    clients can observe the deletion.
                  properties:
                    apiVersion:
                      description: 'APIVersion defines the versioned schema of this
                        representation of an object. Servers should convert recognized
                        schemas to the latest internal value, and may reject unrecognized
                        values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
                      type: string
                    code:
                      description: Suggested HTTP return code for this status, 0 if
                        not set.
                      format: int32
                      type: integer
                    details:
                      description: Extended data associated with the reason.  Each
                        reason may define its own extended details. This field is
                        optional and the data returned is not guaranteed to conform
                        to any schema except that defined by the reason type.
                      properties:
                        causes:
                          description: The Causes array includes more details associated
                            with the StatusReason failure. Not all StatusReasons may
                            provide detailed causes.
                          items:
                            description: StatusCause provides more information about
                              an api.Status failure, including cases when multiple
                              errors are encountered.
                            properties:
                              field:
                                description: "The field of the resource that has caused
                                  this error, as named by its JSON serialization.
                                  May include dot and postfix notation for nested
                                  attributes. Arrays are zero-indexed.  Fields may
                                  appear more than once in an array of causes due
                                  to fields having multiple errors. Optional. \n Examples:
                                  \  \"name\" - the field \"name\" on the current
                                  resource   \"items[0].name\" - the field \"name\"
                                  on the first array entry in \"items\""
                                type: string
                              message:
                                description: A human-readable description of the cause
                                  of the error.  This field may be presented as-is
                                  to a reader.
                                type: string
                              reason:
                                description: A machine-readable description of the
                                  cause of the error. If this value is empty there
                                  is no information available.
                                type: string
                            type: object
                          type: array
                        group:
                          description: The group attribute of the resource associated
                            with the status StatusReason.
                          type: string
                        kind:
                          description: 'The kind attribute of the resource associated
                            with the status StatusReason. On some operations may differ
                            from the requested resource Kind. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                          type: string
                        name:
                          description: The name attribute of the resource associated
                            with the status StatusReason (when there is a single name
                            which can be described).
                          type: string
                        retryAfterSeconds:
                          description: If specified, the time in seconds before the
                            operation should be retried. Some errors may indicate
                            the client must take an alternate action - for those errors
                            this field may indicate how long to wait before taking
                            the alternate action.
                          format: int32
                          type: integer
                        uid:
                          description: 'UID of the resource. (when there is a single
                            resource which can be described). More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                          type: string
                      type: object
                    kind:
                      description: 'Kind is a string value representing the REST resource
                        this object represents. Servers may infer this from the endpoint
                        the client submits requests to. Cannot be updated. In CamelCase.
                        More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      type: string
                    message:
                      description: A human-readable description of the status of this
                        operation.
                      type: string
                    metadata:
                      description: 'Standard list metadata. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                      properties:
                        continue:
                          description: continue may be set if the user set a limit
                            on the number of items returned, and indicates that the
                            server has more data available. The value is opaque and
                            may be used to issue another request to the endpoint that
                            served this list to retrieve the next set of available
                            objects. Continuing a consistent list may not be possible
                            if the server configuration has changed or more than a
                            few minutes have passed. The resourceVersion field returned
                            when using this continue value will be identical to the
                            value in the first response, unless you have received
                            this token from an error message.
                          type: string
                        resourceVersion:
                          description: 'String that identifies the server''s internal
                            version of this object that can be used by clients to
                            determine when objects have changed. Value must be treated
                            as opaque by clients and passed unmodified back to the
                            server. Populated by the system. Read-only. More info:
                            https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency'
                          type: string
                        selfLink:
                          description: selfLink is a URL representing this object.
                            Populated by the system. Read-only.
                          type: string
                      type: object
                    reason:
                      description: A machine-readable description of why this operation
                        is in the "Failure" status. If this value is empty there is
                        no information available. A Reason clarifies an HTTP status
                        code but does not override it.
                      type: string
                    status:
                      description: 'Status of the operation. One of: "Success" or
                        "Failure". More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#spec-and-status'
                      type: string
                  type: object
              required:
              - pending
              type: object
            labels:
              additionalProperties:
                type: string
              description: 'Map of string keys and values that can be used to organize
                and categorize (scope and select) objects. May match selectors of
                replication controllers and services. More info: http://kubernetes.io/docs/user-guide/labels'
              type: object
            managedFields:
              description: "ManagedFields maps workflow-id and version to the set
                of fields that are managed by that workflow. This is mostly for internal
                housekeeping, and users typically shouldn't need to set or understand
                this field. A workflow can be the user's name, a controller's name,
                or the name of a specific apply path like \"ci-cd\". The set of fields
                is always in the version that the workflow used when modifying the
                object. \n This field is alpha and can be changed or removed without
                notice."
              items:
                description: ManagedFieldsEntry is a workflow-id, a FieldSet and the
                  group version of the resource that the fieldset applies to.
                properties:
                  apiVersion:
                    description: APIVersion defines the version of this resource that
                      this field set applies to. The format is "group/version" just
                      like the top-level APIVersion field. It is necessary to track
                      the version of a field set because it cannot be automatically
                      converted.
                    type: string
                  fields:
                    additionalProperties: true
                    description: Fields identifies a set of fields.
                    type: object
                  manager:
                    description: Manager is an identifier of the workflow managing
                      these fields.
                    type: string
                  operation:
                    description: Operation is the type of operation which lead to
                      this ManagedFieldsEntry being created. The only valid values
                      for this field are 'Apply' and 'Update'.
                    type: string
                  time:
                    description: Time is timestamp of when these fields were set.
                      It should always be empty if Operation is 'Apply'
                    format: date-time
                    type: string
                type: object
              type: array
            name:
              description: 'Name must be unique within a namespace. Is required when
                creating resources, although some resources may allow a client to
                request the generation of an appropriate name automatically. Name
                is primarily intended for creation idempotence and configuration definition.
                Cannot be updated. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
              type: string
            namespace:
              description: "Namespace defines the space within each name must be unique.
                An empty namespace is equivalent to the \"default\" namespace, but
                \"default\" is the canonical representation. Not all objects are required
                to be scoped to a namespace - the value of this field for those objects
                will be empty. \n Must be a DNS_LABEL. Cannot be updated. More info:
                http://kubernetes.io/docs/user-guide/namespaces"
              type: string
            ownerReferences:
              description: List of objects depended by this object. If ALL objects
                in the list have been deleted, this object will be garbage collected.
                If this object is managed by a controller, then an entry in this list
                will point to this controller, with the controller field set to true.
                There cannot be more than one managing controller.
              items:
                description: OwnerReference contains enough information to let you
                  identify an owning object. An owning object must be in the same
                  namespace as the dependent, or be cluster-scoped, so there is no
                  namespace field.
                properties:
                  apiVersion:
                    description: API version of the referent.
                    type: string
                  blockOwnerDeletion:
                    description: If true, AND if the owner has the "foregroundDeletion"
                      finalizer, then the owner cannot be deleted from the key-value
                      store until this reference is removed. Defaults to false. To
                      set this field, a user needs "delete" permission of the owner,
                      otherwise 422 (Unprocessable Entity) will be returned.
                    type: boolean
                  controller:
                    description: If true, this reference points to the managing controller.
                    type: boolean
                  kind:
                    description: 'Kind of the referent. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
                    type: string
                  name:
                    description: 'Name of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#names'
                    type: string
                  uid:
                    description: 'UID of the referent. More info: http://kubernetes.io/docs/user-guide/identifiers#uids'
                    type: string
                required:
                - apiVersion
                - kind
                - name
                - uid
                type: object
              type: array
            resourceVersion:
              description: "An opaque value that represents the internal version of
                this object that can be used by clients to determine when objects
                have changed. May be used for optimistic concurrency, change detection,
                and the watch operation on a resource or set of resources. Clients
                must treat these values as opaque and passed unmodified back to the
                server. They may only be valid for a particular resource or set of
                resources. \n Populated by the system. Read-only. Value must be treated
                as opaque by clients and . More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#concurrency-control-and-consistency"
              type: string
            selfLink:
              description: SelfLink is a URL representing this object. Populated by
                the system. Read-only.
              type: string
            uid:
              description: "UID is the unique in time and space value for this object.
                It is typically generated by the server on successful creation of
                a resource and is not allowed to change on PUT operations. \n Populated
                by the system. Read-only. More info: http://kubernetes.io/docs/user-guide/identifiers#uids"
              type: string
          type: object
        spec:
          description: spec holds user settable values for configuration
          properties:
            logLevel:
              description: logLevel is an intent based logging for an overall component.  It
                does not give fine grained control, but it is a simple way to manage
                coarse grained logging choices that operators have to interpret for
                their operands.
              type: string
            managementState:
              description: managementState indicates whether and how the operator
                should manage the component
              pattern: ^(Managed|Unmanaged|Force|Removed)$
              type: string
            observedConfig:
              description: observedConfig holds a sparse config that controller has
                observed from the cluster state.  It exists in spec because it is
                an input to the level for the operator
              nullable: true
              type: object
            operatorLogLevel:
              description: operatorLogLevel is an intent based logging for the operator
                itself.  It does not give fine grained control, but it is a simple
                way to manage coarse grained logging choices that operators have to
                interpret for themselves.
              type: string
            unsupportedConfigOverrides:
              description: 'unsupportedConfigOverrides holds a sparse config that
                will override any previously set options.  It only needs to be the
                fields to override it will end up overlaying in the following order:
                1. hardcoded defaults 2. observedConfig 3. unsupportedConfigOverrides'
              nullable: true
              type: object
          type: object
        status:
          description: status holds observed values from the cluster. They may not
            be overridden.
          properties:
            conditions:
              description: conditions is a list of conditions and their status
              items:
                description: OperatorCondition is just the standard condition fields.
                properties:
                  lastTransitionTime:
                    format: date-time
                    type: string
                  message:
                    type: string
                  reason:
                    type: string
                  status:
                    type: string
                  type:
                    type: string
                type: object
              type: array
            generations:
              description: generations are used to determine when an item needs to
                be reconciled or has changed in a way that needs a reaction.
              items:
                description: GenerationStatus keeps track of the generation for a
                  given resource so that decisions about forced updates can be made.
                properties:
                  group:
                    description: group is the group of the thing you're tracking
                    type: string
                  hash:
                    description: hash is an optional field set for resources without
                      generation that are content sensitive like secrets and configmaps
                    type: string
                  lastGeneration:
                    description: lastGeneration is the last generation of the workload
                      controller involved
                    format: int64
                    type: integer
                  name:
                    description: name is the name of the thing you're tracking
                    type: string
                  namespace:
                    description: namespace is where the thing you're tracking is
                    type: string
                  resource:
                    description: resource is the resource type of the thing you're
                      tracking
                    type: string
                type: object
              type: array
            observedGeneration:
              description: observedGeneration is the last generation change you've
                dealt with
              format: int64
              type: integer
            readyReplicas:
              description: readyReplicas indicates how many replicas are ready and
                at the desired state
              format: int32
              type: integer
            version:
              description: version is the level this availability applies to
              type: string
          type: object
      required:
      - spec
      type: object`)

func assetsCrd0000_50_serviceCaOperator_02_crdYamlBytes() ([]byte, error) {
	return _assetsCrd0000_50_serviceCaOperator_02_crdYaml, nil
}

func assetsCrd0000_50_serviceCaOperator_02_crdYaml() (*asset, error) {
	bytes, err := assetsCrd0000_50_serviceCaOperator_02_crdYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_50_service-ca-operator_02_crd.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrd0000_70_dnsOperator_00CustomResourceDefinitionYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  creationTimestamp: null
  name: dnses.operator.openshift.io
spec:
  group: operator.openshift.io
  names:
    kind: DNS
    listKind: DNSList
    plural: dnses
    singular: dns
  scope: Cluster
  preserveUnknownFields: false
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: "DNS manages the CoreDNS component to provide a name resolution
        service for pods and services in the cluster. \n This supports the DNS-based
        service discovery specification: https://github.com/kubernetes/dns/blob/master/docs/specification.md
        \n More details: https://kubernetes.io/docs/tasks/administer-cluster/coredns"
      type: object
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: spec is the specification of the desired behavior of the DNS.
          type: object
          properties:
            servers:
              description: "servers is a list of DNS resolvers that provide name query
                delegation for one or more subdomains outside the scope of the cluster
                domain. If servers consists of more than one Server, longest suffix
                match will be used to determine the Server. \n For example, if there
                are two Servers, one for \"foo.com\" and another for \"a.foo.com\",
                and the name query is for \"www.a.foo.com\", it will be routed to
                the Server with Zone \"a.foo.com\". \n If this field is nil, no servers
                are created."
              type: array
              items:
                description: Server defines the schema for a server that runs per
                  instance of CoreDNS.
                type: object
                properties:
                  forwardPlugin:
                    description: forwardPlugin defines a schema for configuring CoreDNS
                      to proxy DNS messages to upstream resolvers.
                    type: object
                    properties:
                      upstreams:
                        description: "upstreams is a list of resolvers to forward
                          name queries for subdomains of Zones. Upstreams are randomized
                          when more than 1 upstream is specified. Each instance of
                          CoreDNS performs health checking of Upstreams. When a healthy
                          upstream returns an error during the exchange, another resolver
                          is tried from Upstreams. Each upstream is represented by
                          an IP address or IP:port if the upstream listens on a port
                          other than 53. \n A maximum of 15 upstreams is allowed per
                          ForwardPlugin."
                        type: array
                        maxItems: 15
                        items:
                          type: string
                  name:
                    description: name is required and specifies a unique name for
                      the server. Name must comply with the Service Name Syntax of
                      rfc6335.
                    type: string
                  zones:
                    description: zones is required and specifies the subdomains that
                      Server is authoritative for. Zones must conform to the rfc1123
                      definition of a subdomain. Specifying the cluster domain (i.e.,
                      "cluster.local") is invalid.
                    type: array
                    items:
                      type: string
        status:
          description: status is the most recently observed status of the DNS.
          type: object
          required:
          - clusterDomain
          - clusterIP
          properties:
            clusterDomain:
              description: "clusterDomain is the local cluster DNS domain suffix for
                DNS services. This will be a subdomain as defined in RFC 1034, section
                3.5: https://tools.ietf.org/html/rfc1034#section-3.5 Example: \"cluster.local\"
                \n More info: https://kubernetes.io/docs/concepts/services-networking/dns-pod-service"
              type: string
            clusterIP:
              description: "clusterIP is the service IP through which this DNS is
                made available. \n In the case of the default DNS, this will be a
                well known IP that is used as the default nameserver for pods that
                are using the default ClusterFirst DNS policy. \n In general, this
                IP can be specified in a pod's spec.dnsConfig.nameservers list or
                used explicitly when performing name resolution from within the cluster.
                Example: dig foo.com @<service IP> \n More info: https://kubernetes.io/docs/concepts/services-networking/service/#virtual-ips-and-service-proxies"
              type: string
            conditions:
              description: "conditions provide information about the state of the
                DNS on the cluster. \n These are the supported DNS conditions: \n
                \  * Available   - True if the following conditions are met:     *
                DNS controller daemonset is available.   - False if any of those conditions
                are unsatisfied."
              type: array
              items:
                description: OperatorCondition is just the standard condition fields.
                type: object
                properties:
                  lastTransitionTime:
                    type: string
                    format: date-time
                  message:
                    type: string
                  reason:
                    type: string
                  status:
                    type: string
                  type:
                    type: string
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`)

func assetsCrd0000_70_dnsOperator_00CustomResourceDefinitionYamlBytes() ([]byte, error) {
	return _assetsCrd0000_70_dnsOperator_00CustomResourceDefinitionYaml, nil
}

func assetsCrd0000_70_dnsOperator_00CustomResourceDefinitionYaml() (*asset, error) {
	bytes, err := assetsCrd0000_70_dnsOperator_00CustomResourceDefinitionYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/0000_70_dns-operator_00-custom-resource-definition.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrdClusterIngress00CustomResourceDefinitionYaml = []byte(`apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  creationTimestamp: null
  name: ingresscontrollers.operator.openshift.io
spec:
  group: operator.openshift.io
  names:
    kind: IngressController
    listKind: IngressControllerList
    plural: ingresscontrollers
    singular: ingresscontroller
  preserveUnknownFields: false
  scope: ""
  subresources:
    scale:
      labelSelectorPath: .status.selector
      specReplicasPath: .spec.replicas
      statusReplicasPath: .status.availableReplicas
    status: {}
  validation:
    openAPIV3Schema:
      description: "IngressController describes a managed ingress controller for the
        cluster. The controller can service OpenShift Route and Kubernetes Ingress
        resources. \n When an IngressController is created, a new ingress controller
        deployment is created to allow external traffic to reach the services that
        expose Ingress or Route resources. Updating this resource may lead to disruption
        for public facing network connections as a new ingress controller revision
        may be rolled out. \n https://kubernetes.io/docs/concepts/services-networking/ingress-controllers
        \n Whenever possible, sensible defaults for the platform are used. See each
        field for more details."
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          description: spec is the specification of the desired behavior of the IngressController.
          properties:
            defaultCertificate:
              description: "defaultCertificate is a reference to a secret containing
                the default certificate served by the ingress controller. When Routes
                don't specify their own certificate, defaultCertificate is used. \n
                The secret must contain the following keys and data: \n   tls.crt:
                certificate file contents   tls.key: key file contents \n If unset,
                a wildcard certificate is automatically generated and used. The certificate
                is valid for the ingress controller domain (and subdomains) and the
                generated certificate's CA will be automatically integrated with the
                cluster's trust store. \n The in-use certificate (whether generated
                or user-specified) will be automatically integrated with OpenShift's
                built-in OAuth server."
              properties:
                name:
                  description: 'Name of the referent. More info: https://kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
                    TODO: Add other useful fields. apiVersion, kind, uid?'
                  type: string
              type: object
            domain:
              description: "domain is a DNS name serviced by the ingress controller
                and is used to configure multiple features: \n * For the LoadBalancerService
                endpoint publishing strategy, domain is   used to configure DNS records.
                See endpointPublishingStrategy. \n * When using a generated default
                certificate, the certificate will be valid   for domain and its subdomains.
                See defaultCertificate. \n * The value is published to individual
                Route statuses so that end-users   know where to target external DNS
                records. \n domain must be unique among all IngressControllers, and
                cannot be updated. \n If empty, defaults to ingress.config.openshift.io/cluster
                .spec.domain."
              type: string
            endpointPublishingStrategy:
              description: "endpointPublishingStrategy is used to publish the ingress
                controller endpoints to other networks, enable load balancer integrations,
                etc. \n If unset, the default is based on infrastructure.config.openshift.io/cluster
                .status.platform: \n   AWS:      LoadBalancerService (with External
                scope)   Azure:    LoadBalancerService (with External scope)   GCP:
                \     LoadBalancerService (with External scope)   IBMCloud: LoadBalancerService
                (with External scope)   Libvirt:  HostNetwork \n Any other platform
                types (including None) default to HostNetwork. \n endpointPublishingStrategy
                cannot be updated."
              properties:
                hostNetwork:
                  description: hostNetwork holds parameters for the HostNetwork endpoint
                    publishing strategy. Present only if type is HostNetwork.
                  type: object
                loadBalancer:
                  description: loadBalancer holds parameters for the load balancer.
                    Present only if type is LoadBalancerService.
                  properties:
                    scope:
                      description: scope indicates the scope at which the load balancer
                        is exposed. Possible values are "External" and "Internal".
                      enum:
                      - Internal
                      - External
                      type: string
                  required:
                  - scope
                  type: object
                nodePort:
                  description: nodePort holds parameters for the NodePortService endpoint
                    publishing strategy. Present only if type is NodePortService.
                  type: object
                private:
                  description: private holds parameters for the Private endpoint publishing
                    strategy. Present only if type is Private.
                  type: object
                type:
                  description: "type is the publishing strategy to use. Valid values
                    are: \n * LoadBalancerService \n Publishes the ingress controller
                    using a Kubernetes LoadBalancer Service. \n In this configuration,
                    the ingress controller deployment uses container networking. A
                    LoadBalancer Service is created to publish the deployment. \n
                    See: https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer
                    \n If domain is set, a wildcard DNS record will be managed to
                    point at the LoadBalancer Service's external name. DNS records
                    are managed only in DNS zones defined by dns.config.openshift.io/cluster
                    .spec.publicZone and .spec.privateZone. \n Wildcard DNS management
                    is currently supported only on the AWS, Azure, and GCP platforms.
                    \n * HostNetwork \n Publishes the ingress controller on node ports
                    where the ingress controller is deployed. \n In this configuration,
                    the ingress controller deployment uses host networking, bound
                    to node ports 80 and 443. The user is responsible for configuring
                    an external load balancer to publish the ingress controller via
                    the node ports. \n * Private \n Does not publish the ingress controller.
                    \n In this configuration, the ingress controller deployment uses
                    container networking, and is not explicitly published. The user
                    must manually publish the ingress controller. \n * NodePortService
                    \n Publishes the ingress controller using a Kubernetes NodePort
                    Service. \n In this configuration, the ingress controller deployment
                    uses container networking. A NodePort Service is created to publish
                    the deployment. The specific node ports are dynamically allocated
                    by OpenShift; however, to support static port allocations, user
                    changes to the node port field of the managed NodePort Service
                    will preserved."
                  enum:
                  - LoadBalancerService
                  - HostNetwork
                  - Private
                  - NodePortService
                  type: string
              required:
              - type
              type: object
            logging:
              description: logging defines parameters for what should be logged where.  If
                this field is empty, operational logs are enabled but access logs
                are disabled.
              properties:
                access:
                  description: "access describes how the client requests should be
                    logged. \n If this field is empty, access logging is disabled."
                  properties:
                    destination:
                      description: destination is where access logs go.
                      properties:
                        container:
                          description: container holds parameters for the Container
                            logging destination. Present only if type is Container.
                          type: object
                        syslog:
                          description: syslog holds parameters for a syslog endpoint.  Present
                            only if type is Syslog.
                          oneOf:
                          - properties:
                              address:
                                format: ipv4
                          - properties:
                              address:
                                format: ipv6
                          properties:
                            address:
                              description: address is the IP address of the syslog
                                endpoint that receives log messages.
                              type: string
                            facility:
                              description: "facility specifies the syslog facility
                                of log messages. \n If this field is empty, the facility
                                is \"local1\"."
                              enum:
                              - kern
                              - user
                              - mail
                              - daemon
                              - auth
                              - syslog
                              - lpr
                              - news
                              - uucp
                              - cron
                              - auth2
                              - ftp
                              - ntp
                              - audit
                              - alert
                              - cron2
                              - local0
                              - local1
                              - local2
                              - local3
                              - local4
                              - local5
                              - local6
                              - local7
                              type: string
                            port:
                              description: port is the UDP port number of the syslog
                                endpoint that receives log messages.
                              format: int32
                              maximum: 65535
                              minimum: 1
                              type: integer
                          required:
                          - address
                          - port
                          type: object
                        type:
                          description: "type is the type of destination for logs.
                            \ It must be one of the following: \n * Container \n The
                            ingress operator configures the sidecar container named
                            \"logs\" on the ingress controller pod and configures
                            the ingress controller to write logs to the sidecar.  The
                            logs are then available as container logs.  The expectation
                            is that the administrator configures a custom logging
                            solution that reads logs from this sidecar.  Note that
                            using container logs means that logs may be dropped if
                            the rate of logs exceeds the container runtime's or the
                            custom logging solution's capacity. \n * Syslog \n Logs
                            are sent to a syslog endpoint.  The administrator must
                            specify an endpoint that can receive syslog messages.
                            \ The expectation is that the administrator has configured
                            a custom syslog instance."
                          enum:
                          - Container
                          - Syslog
                          type: string
                      required:
                      - type
                      type: object
                    httpLogFormat:
                      description: "httpLogFormat specifies the format of the log
                        message for an HTTP request. \n If this field is empty, log
                        messages use the implementation's default HTTP log format.
                        \ For HAProxy's default HTTP log format, see the HAProxy documentation:
                        http://cbonte.github.io/haproxy-dconv/2.0/configuration.html#8.2.3"
                      type: string
                  required:
                  - destination
                  type: object
              type: object
            namespaceSelector:
              description: "namespaceSelector is used to filter the set of namespaces
                serviced by the ingress controller. This is useful for implementing
                shards. \n If unset, the default is no filtering."
              properties:
                matchExpressions:
                  description: matchExpressions is a list of label selector requirements.
                    The requirements are ANDed.
                  items:
                    description: A label selector requirement is a selector that contains
                      values, a key, and an operator that relates the key and values.
                    properties:
                      key:
                        description: key is the label key that the selector applies
                          to.
                        type: string
                      operator:
                        description: operator represents a key's relationship to a
                          set of values. Valid operators are In, NotIn, Exists and
                          DoesNotExist.
                        type: string
                      values:
                        description: values is an array of string values. If the operator
                          is In or NotIn, the values array must be non-empty. If the
                          operator is Exists or DoesNotExist, the values array must
                          be empty. This array is replaced during a strategic merge
                          patch.
                        items:
                          type: string
                        type: array
                    required:
                    - key
                    - operator
                    type: object
                  type: array
                matchLabels:
                  additionalProperties:
                    type: string
                  description: matchLabels is a map of {key,value} pairs. A single
                    {key,value} in the matchLabels map is equivalent to an element
                    of matchExpressions, whose key field is "key", the operator is
                    "In", and the values array contains only "value". The requirements
                    are ANDed.
                  type: object
              type: object
            nodePlacement:
              description: "nodePlacement enables explicit control over the scheduling
                of the ingress controller. \n If unset, defaults are used. See NodePlacement
                for more details."
              properties:
                nodeSelector:
                  description: "nodeSelector is the node selector applied to ingress
                    controller deployments. \n If unset, the default is: \n   beta.kubernetes.io/os:
                    linux   node-role.kubernetes.io/worker: '' \n If set, the specified
                    selector is used and replaces the default."
                  properties:
                    matchExpressions:
                      description: matchExpressions is a list of label selector requirements.
                        The requirements are ANDed.
                      items:
                        description: A label selector requirement is a selector that
                          contains values, a key, and an operator that relates the
                          key and values.
                        properties:
                          key:
                            description: key is the label key that the selector applies
                              to.
                            type: string
                          operator:
                            description: operator represents a key's relationship
                              to a set of values. Valid operators are In, NotIn, Exists
                              and DoesNotExist.
                            type: string
                          values:
                            description: values is an array of string values. If the
                              operator is In or NotIn, the values array must be non-empty.
                              If the operator is Exists or DoesNotExist, the values
                              array must be empty. This array is replaced during a
                              strategic merge patch.
                            items:
                              type: string
                            type: array
                        required:
                        - key
                        - operator
                        type: object
                      type: array
                    matchLabels:
                      additionalProperties:
                        type: string
                      description: matchLabels is a map of {key,value} pairs. A single
                        {key,value} in the matchLabels map is equivalent to an element
                        of matchExpressions, whose key field is "key", the operator
                        is "In", and the values array contains only "value". The requirements
                        are ANDed.
                      type: object
                  type: object
                tolerations:
                  description: "tolerations is a list of tolerations applied to ingress
                    controller deployments. \n The default is an empty list. \n See
                    https://kubernetes.io/docs/concepts/configuration/taint-and-toleration/"
                  items:
                    description: The pod this Toleration is attached to tolerates
                      any taint that matches the triple <key,value,effect> using the
                      matching operator <operator>.
                    properties:
                      effect:
                        description: Effect indicates the taint effect to match. Empty
                          means match all taint effects. When specified, allowed values
                          are NoSchedule, PreferNoSchedule and NoExecute.
                        type: string
                      key:
                        description: Key is the taint key that the toleration applies
                          to. Empty means match all taint keys. If the key is empty,
                          operator must be Exists; this combination means to match
                          all values and all keys.
                        type: string
                      operator:
                        description: Operator represents a key's relationship to the
                          value. Valid operators are Exists and Equal. Defaults to
                          Equal. Exists is equivalent to wildcard for value, so that
                          a pod can tolerate all taints of a particular category.
                        type: string
                      tolerationSeconds:
                        description: TolerationSeconds represents the period of time
                          the toleration (which must be of effect NoExecute, otherwise
                          this field is ignored) tolerates the taint. By default,
                          it is not set, which means tolerate the taint forever (do
                          not evict). Zero and negative values will be treated as
                          0 (evict immediately) by the system.
                        format: int64
                        type: integer
                      value:
                        description: Value is the taint value the toleration matches
                          to. If the operator is Exists, the value should be empty,
                          otherwise just a regular string.
                        type: string
                    type: object
                  type: array
              type: object
            replicas:
              description: replicas is the desired number of ingress controller replicas.
                If unset, defaults to 2.
              format: int32
              type: integer
            routeAdmission:
              description: "routeAdmission defines a policy for handling new route
                claims (for example, to allow or deny claims across namespaces). \n
                If empty, defaults will be applied. See specific routeAdmission fields
                for details about their defaults."
              properties:
                namespaceOwnership:
                  description: "namespaceOwnership describes how host name claims
                    across namespaces should be handled. \n Value must be one of:
                    \n - Strict: Do not allow routes in different namespaces to claim
                    the same host. \n - InterNamespaceAllowed: Allow routes to claim
                    different paths of the same   host name across namespaces. \n
                    If empty, the default is Strict."
                  enum:
                  - InterNamespaceAllowed
                  - Strict
                  type: string
                wildcardPolicy:
                  description: "wildcardPolicy describes how routes with wildcard
                    policies should be handled for the ingress controller. WildcardPolicy
                    controls use of routes [1] exposed by the ingress controller based
                    on the route's wildcard policy. \n [1] https://github.com/openshift/api/blob/master/route/v1/types.go
                    \n Note: Updating WildcardPolicy from WildcardsAllowed to WildcardsDisallowed
                    will cause admitted routes with a wildcard policy of Subdomain
                    to stop working. These routes must be updated to a wildcard policy
                    of None to be readmitted by the ingress controller. \n WildcardPolicy
                    supports WildcardsAllowed and WildcardsDisallowed values. \n If
                    empty, defaults to \"WildcardsDisallowed\"."
                  enum:
                  - WildcardsAllowed
                  - WildcardsDisallowed
                  type: string
              type: object
            routeSelector:
              description: "routeSelector is used to filter the set of Routes serviced
                by the ingress controller. This is useful for implementing shards.
                \n If unset, the default is no filtering."
              properties:
                matchExpressions:
                  description: matchExpressions is a list of label selector requirements.
                    The requirements are ANDed.
                  items:
                    description: A label selector requirement is a selector that contains
                      values, a key, and an operator that relates the key and values.
                    properties:
                      key:
                        description: key is the label key that the selector applies
                          to.
                        type: string
                      operator:
                        description: operator represents a key's relationship to a
                          set of values. Valid operators are In, NotIn, Exists and
                          DoesNotExist.
                        type: string
                      values:
                        description: values is an array of string values. If the operator
                          is In or NotIn, the values array must be non-empty. If the
                          operator is Exists or DoesNotExist, the values array must
                          be empty. This array is replaced during a strategic merge
                          patch.
                        items:
                          type: string
                        type: array
                    required:
                    - key
                    - operator
                    type: object
                  type: array
                matchLabels:
                  additionalProperties:
                    type: string
                  description: matchLabels is a map of {key,value} pairs. A single
                    {key,value} in the matchLabels map is equivalent to an element
                    of matchExpressions, whose key field is "key", the operator is
                    "In", and the values array contains only "value". The requirements
                    are ANDed.
                  type: object
              type: object
            tlsSecurityProfile:
              description: "tlsSecurityProfile specifies settings for TLS connections
                for ingresscontrollers. \n If unset, the default is based on the apiservers.config.openshift.io/cluster
                resource. \n Note that when using the Old, Intermediate, and Modern
                profile types, the effective profile configuration is subject to change
                between releases. For example, given a specification to use the Intermediate
                profile deployed on release X.Y.Z, an upgrade to release X.Y.Z+1 may
                cause a new profile configuration to be applied to the ingress controller,
                resulting in a rollout. \n Note that the minimum TLS version for ingress
                controllers is 1.1, and the maximum TLS version is 1.2.  An implication
                of this restriction is that the Modern TLS profile type cannot be
                used because it requires TLS 1.3."
              properties:
                custom:
                  description: "custom is a user-defined TLS security profile. Be
                    extremely careful using a custom profile as invalid configurations
                    can be catastrophic. An example custom profile looks like this:
                    \n   ciphers:     - ECDHE-ECDSA-CHACHA20-POLY1305     - ECDHE-RSA-CHACHA20-POLY1305
                    \    - ECDHE-RSA-AES128-GCM-SHA256     - ECDHE-ECDSA-AES128-GCM-SHA256
                    \  minTLSVersion: TLSv1.1"
                  nullable: true
                  properties:
                    ciphers:
                      description: "ciphers is used to specify the cipher algorithms
                        that are negotiated during the TLS handshake.  Operators may
                        remove entries their operands do not support.  For example,
                        to use DES-CBC3-SHA  (yaml): \n   ciphers:     - DES-CBC3-SHA"
                      items:
                        type: string
                      type: array
                    minTLSVersion:
                      description: "minTLSVersion is used to specify the minimal version
                        of the TLS protocol that is negotiated during the TLS handshake.
                        For example, to use TLS versions 1.1, 1.2 and 1.3 (yaml):
                        \n   minTLSVersion: TLSv1.1 \n NOTE: currently the highest
                        minTLSVersion allowed is VersionTLS12"
                      enum:
                      - VersionTLS10
                      - VersionTLS11
                      - VersionTLS12
                      - VersionTLS13
                      type: string
                  type: object
                intermediate:
                  description: "intermediate is a TLS security profile based on: \n
                    https://wiki.mozilla.org/Security/Server_Side_TLS#Intermediate_compatibility_.28recommended.29
                    \n and looks like this (yaml): \n   ciphers:     - TLS_AES_128_GCM_SHA256
                    \    - TLS_AES_256_GCM_SHA384     - TLS_CHACHA20_POLY1305_SHA256
                    \    - ECDHE-ECDSA-AES128-GCM-SHA256     - ECDHE-RSA-AES128-GCM-SHA256
                    \    - ECDHE-ECDSA-AES256-GCM-SHA384     - ECDHE-RSA-AES256-GCM-SHA384
                    \    - ECDHE-ECDSA-CHACHA20-POLY1305     - ECDHE-RSA-CHACHA20-POLY1305
                    \    - DHE-RSA-AES128-GCM-SHA256     - DHE-RSA-AES256-GCM-SHA384
                    \  minTLSVersion: TLSv1.2"
                  nullable: true
                  type: object
                modern:
                  description: "modern is a TLS security profile based on: \n https://wiki.mozilla.org/Security/Server_Side_TLS#Modern_compatibility
                    \n and looks like this (yaml): \n   ciphers:     - TLS_AES_128_GCM_SHA256
                    \    - TLS_AES_256_GCM_SHA384     - TLS_CHACHA20_POLY1305_SHA256
                    \  minTLSVersion: TLSv1.3 \n NOTE: Currently unsupported."
                  nullable: true
                  type: object
                old:
                  description: "old is a TLS security profile based on: \n https://wiki.mozilla.org/Security/Server_Side_TLS#Old_backward_compatibility
                    \n and looks like this (yaml): \n   ciphers:     - TLS_AES_128_GCM_SHA256
                    \    - TLS_AES_256_GCM_SHA384     - TLS_CHACHA20_POLY1305_SHA256
                    \    - ECDHE-ECDSA-AES128-GCM-SHA256     - ECDHE-RSA-AES128-GCM-SHA256
                    \    - ECDHE-ECDSA-AES256-GCM-SHA384     - ECDHE-RSA-AES256-GCM-SHA384
                    \    - ECDHE-ECDSA-CHACHA20-POLY1305     - ECDHE-RSA-CHACHA20-POLY1305
                    \    - DHE-RSA-AES128-GCM-SHA256     - DHE-RSA-AES256-GCM-SHA384
                    \    - DHE-RSA-CHACHA20-POLY1305     - ECDHE-ECDSA-AES128-SHA256
                    \    - ECDHE-RSA-AES128-SHA256     - ECDHE-ECDSA-AES128-SHA     -
                    ECDHE-RSA-AES128-SHA     - ECDHE-ECDSA-AES256-SHA384     - ECDHE-RSA-AES256-SHA384
                    \    - ECDHE-ECDSA-AES256-SHA     - ECDHE-RSA-AES256-SHA     -
                    DHE-RSA-AES128-SHA256     - DHE-RSA-AES256-SHA256     - AES128-GCM-SHA256
                    \    - AES256-GCM-SHA384     - AES128-SHA256     - AES256-SHA256
                    \    - AES128-SHA     - AES256-SHA     - DES-CBC3-SHA   minTLSVersion:
                    TLSv1.0"
                  nullable: true
                  type: object
                type:
                  description: "type is one of Old, Intermediate, Modern or Custom.
                    Custom provides the ability to specify individual TLS security
                    profile parameters. Old, Intermediate and Modern are TLS security
                    profiles based on: \n https://wiki.mozilla.org/Security/Server_Side_TLS#Recommended_configurations
                    \n The profiles are intent based, so they may change over time
                    as new ciphers are developed and existing ciphers are found to
                    be insecure.  Depending on precisely which ciphers are available
                    to a process, the list may be reduced. \n Note that the Modern
                    profile is currently not supported because it is not yet well
                    adopted by common software libraries."
                  enum:
                  - Old
                  - Intermediate
                  - Modern
                  - Custom
                  type: string
              type: object
          type: object
        status:
          description: status is the most recently observed status of the IngressController.
          properties:
            availableReplicas:
              description: availableReplicas is number of observed available replicas
                according to the ingress controller deployment.
              format: int32
              type: integer
            conditions:
              description: "conditions is a list of conditions and their status. \n
                Available means the ingress controller deployment is available and
                servicing route and ingress resources (i.e, .status.availableReplicas
                equals .spec.replicas) \n There are additional conditions which indicate
                the status of other ingress controller features and capabilities.
                \n   * LoadBalancerManaged   - True if the following conditions are
                met:     * The endpoint publishing strategy requires a service load
                balancer.   - False if any of those conditions are unsatisfied. \n
                \  * LoadBalancerReady   - True if the following conditions are met:
                \    * A load balancer is managed.     * The load balancer is ready.
                \  - False if any of those conditions are unsatisfied. \n   * DNSManaged
                \  - True if the following conditions are met:     * The endpoint
                publishing strategy and platform support DNS.     * The ingress controller
                domain is set.     * dns.config.openshift.io/cluster configures DNS
                zones.   - False if any of those conditions are unsatisfied. \n   *
                DNSReady   - True if the following conditions are met:     * DNS is
                managed.     * DNS records have been successfully created.   - False
                if any of those conditions are unsatisfied."
              items:
                description: OperatorCondition is just the standard condition fields.
                properties:
                  lastTransitionTime:
                    format: date-time
                    type: string
                  message:
                    type: string
                  reason:
                    type: string
                  status:
                    type: string
                  type:
                    type: string
                type: object
              type: array
            domain:
              description: domain is the actual domain in use.
              type: string
            endpointPublishingStrategy:
              description: endpointPublishingStrategy is the actual strategy in use.
              properties:
                hostNetwork:
                  description: hostNetwork holds parameters for the HostNetwork endpoint
                    publishing strategy. Present only if type is HostNetwork.
                  type: object
                loadBalancer:
                  description: loadBalancer holds parameters for the load balancer.
                    Present only if type is LoadBalancerService.
                  properties:
                    scope:
                      description: scope indicates the scope at which the load balancer
                        is exposed. Possible values are "External" and "Internal".
                      enum:
                      - Internal
                      - External
                      type: string
                  required:
                  - scope
                  type: object
                nodePort:
                  description: nodePort holds parameters for the NodePortService endpoint
                    publishing strategy. Present only if type is NodePortService.
                  type: object
                private:
                  description: private holds parameters for the Private endpoint publishing
                    strategy. Present only if type is Private.
                  type: object
                type:
                  description: "type is the publishing strategy to use. Valid values
                    are: \n * LoadBalancerService \n Publishes the ingress controller
                    using a Kubernetes LoadBalancer Service. \n In this configuration,
                    the ingress controller deployment uses container networking. A
                    LoadBalancer Service is created to publish the deployment. \n
                    See: https://kubernetes.io/docs/concepts/services-networking/service/#loadbalancer
                    \n If domain is set, a wildcard DNS record will be managed to
                    point at the LoadBalancer Service's external name. DNS records
                    are managed only in DNS zones defined by dns.config.openshift.io/cluster
                    .spec.publicZone and .spec.privateZone. \n Wildcard DNS management
                    is currently supported only on the AWS, Azure, and GCP platforms.
                    \n * HostNetwork \n Publishes the ingress controller on node ports
                    where the ingress controller is deployed. \n In this configuration,
                    the ingress controller deployment uses host networking, bound
                    to node ports 80 and 443. The user is responsible for configuring
                    an external load balancer to publish the ingress controller via
                    the node ports. \n * Private \n Does not publish the ingress controller.
                    \n In this configuration, the ingress controller deployment uses
                    container networking, and is not explicitly published. The user
                    must manually publish the ingress controller. \n * NodePortService
                    \n Publishes the ingress controller using a Kubernetes NodePort
                    Service. \n In this configuration, the ingress controller deployment
                    uses container networking. A NodePort Service is created to publish
                    the deployment. The specific node ports are dynamically allocated
                    by OpenShift; however, to support static port allocations, user
                    changes to the node port field of the managed NodePort Service
                    will preserved."
                  enum:
                  - LoadBalancerService
                  - HostNetwork
                  - Private
                  - NodePortService
                  type: string
              required:
              - type
              type: object
            observedGeneration:
              description: observedGeneration is the most recent generation observed.
              format: int64
              type: integer
            selector:
              description: selector is a label selector, in string format, for ingress
                controller pods corresponding to the IngressController. The number
                of matching pods should equal the value of availableReplicas.
              type: string
            tlsProfile:
              description: tlsProfile is the TLS connection configuration that is
                in effect.
              properties:
                ciphers:
                  description: "ciphers is used to specify the cipher algorithms that
                    are negotiated during the TLS handshake.  Operators may remove
                    entries their operands do not support.  For example, to use DES-CBC3-SHA
                    \ (yaml): \n   ciphers:     - DES-CBC3-SHA"
                  items:
                    type: string
                  type: array
                minTLSVersion:
                  description: "minTLSVersion is used to specify the minimal version
                    of the TLS protocol that is negotiated during the TLS handshake.
                    For example, to use TLS versions 1.1, 1.2 and 1.3 (yaml): \n   minTLSVersion:
                    TLSv1.1 \n NOTE: currently the highest minTLSVersion allowed is
                    VersionTLS12"
                  enum:
                  - VersionTLS10
                  - VersionTLS11
                  - VersionTLS12
                  - VersionTLS13
                  type: string
              type: object
          type: object
      type: object
  version: v1
  versions:
  - name: v1
    served: true
    storage: true
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []
`)

func assetsCrdClusterIngress00CustomResourceDefinitionYamlBytes() ([]byte, error) {
	return _assetsCrdClusterIngress00CustomResourceDefinitionYaml, nil
}

func assetsCrdClusterIngress00CustomResourceDefinitionYaml() (*asset, error) {
	bytes, err := assetsCrdClusterIngress00CustomResourceDefinitionYamlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/cluster-ingress-00-custom-resource-definition.yaml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

var _assetsCrdClusterNetwork01CrdYml = []byte(`
---
# This is the advanced network configuration CRD
# Only necessary if you need to tweak certain settings.
# See https://github.com/openshift/cluster-network-operator#configuring
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: networks.operator.openshift.io
spec:
  group: operator.openshift.io
  names:
    kind: Network
    listKind: NetworkList
    plural: networks
    singular: network
  scope: Cluster
  versions:
  - name: v1
    served: true
    storage: true
`)

func assetsCrdClusterNetwork01CrdYmlBytes() ([]byte, error) {
	return _assetsCrdClusterNetwork01CrdYml, nil
}

func assetsCrdClusterNetwork01CrdYml() (*asset, error) {
	bytes, err := assetsCrdClusterNetwork01CrdYmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "assets/crd/cluster-network-01-crd.yml", size: 0, mode: os.FileMode(0), modTime: time.Unix(0, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"assets/crd/0000_00_cluster-version-operator_01_clusteroperator.crd.yaml":       assetsCrd0000_00_clusterVersionOperator_01_clusteroperatorCrdYaml,
	"assets/crd/0000_00_cluster-version-operator_01_clusterversion.crd.yaml":        assetsCrd0000_00_clusterVersionOperator_01_clusterversionCrdYaml,
	"assets/crd/0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml": assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml,
	"assets/crd/0000_03_config-operator_01_operatorhub.crd.yaml":                    assetsCrd0000_03_configOperator_01_operatorhubCrdYaml,
	"assets/crd/0000_03_config-operator_01_proxy.crd.yaml":                          assetsCrd0000_03_configOperator_01_proxyCrdYaml,
	"assets/crd/0000_03_quota-openshift_01_clusterresourcequota.crd.yaml":           assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYaml,
	"assets/crd/0000_03_security-openshift_01_scc.crd.yaml":                         assetsCrd0000_03_securityOpenshift_01_sccCrdYaml,
	"assets/crd/0000_10_config-operator_01_apiserver.crd.yaml":                      assetsCrd0000_10_configOperator_01_apiserverCrdYaml,
	"assets/crd/0000_10_config-operator_01_authentication.crd.yaml":                 assetsCrd0000_10_configOperator_01_authenticationCrdYaml,
	"assets/crd/0000_10_config-operator_01_build.crd.yaml":                          assetsCrd0000_10_configOperator_01_buildCrdYaml,
	"assets/crd/0000_10_config-operator_01_console.crd.yaml":                        assetsCrd0000_10_configOperator_01_consoleCrdYaml,
	"assets/crd/0000_10_config-operator_01_dns.crd.yaml":                            assetsCrd0000_10_configOperator_01_dnsCrdYaml,
	"assets/crd/0000_10_config-operator_01_featuregate.crd.yaml":                    assetsCrd0000_10_configOperator_01_featuregateCrdYaml,
	"assets/crd/0000_10_config-operator_01_image.crd.yaml":                          assetsCrd0000_10_configOperator_01_imageCrdYaml,
	"assets/crd/0000_10_config-operator_01_imagecontentsourcepolicy.crd.yaml":       assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYaml,
	"assets/crd/0000_10_config-operator_01_infrastructure.crd.yaml":                 assetsCrd0000_10_configOperator_01_infrastructureCrdYaml,
	"assets/crd/0000_10_config-operator_01_ingress.crd.yaml":                        assetsCrd0000_10_configOperator_01_ingressCrdYaml,
	"assets/crd/0000_10_config-operator_01_network.crd.yaml":                        assetsCrd0000_10_configOperator_01_networkCrdYaml,
	"assets/crd/0000_10_config-operator_01_oauth.crd.yaml":                          assetsCrd0000_10_configOperator_01_oauthCrdYaml,
	"assets/crd/0000_10_config-operator_01_project.crd.yaml":                        assetsCrd0000_10_configOperator_01_projectCrdYaml,
	"assets/crd/0000_11_imageregistry-configs.crd.yaml":                             assetsCrd0000_11_imageregistryConfigsCrdYaml,
	"assets/crd/0000_50_service-ca-operator_02_crd.yaml":                            assetsCrd0000_50_serviceCaOperator_02_crdYaml,
	"assets/crd/0000_70_dns-operator_00-custom-resource-definition.yaml":            assetsCrd0000_70_dnsOperator_00CustomResourceDefinitionYaml,
	"assets/crd/cluster-ingress-00-custom-resource-definition.yaml":                 assetsCrdClusterIngress00CustomResourceDefinitionYaml,
	"assets/crd/cluster-network-01-crd.yml":                                         assetsCrdClusterNetwork01CrdYml,
}

// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}

type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{nil, map[string]*bintree{
	"assets": {nil, map[string]*bintree{
		"crd": {nil, map[string]*bintree{
			"0000_00_cluster-version-operator_01_clusteroperator.crd.yaml":       {assetsCrd0000_00_clusterVersionOperator_01_clusteroperatorCrdYaml, map[string]*bintree{}},
			"0000_00_cluster-version-operator_01_clusterversion.crd.yaml":        {assetsCrd0000_00_clusterVersionOperator_01_clusterversionCrdYaml, map[string]*bintree{}},
			"0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml": {assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml, map[string]*bintree{}},
			"0000_03_config-operator_01_operatorhub.crd.yaml":                    {assetsCrd0000_03_configOperator_01_operatorhubCrdYaml, map[string]*bintree{}},
			"0000_03_config-operator_01_proxy.crd.yaml":                          {assetsCrd0000_03_configOperator_01_proxyCrdYaml, map[string]*bintree{}},
			"0000_03_quota-openshift_01_clusterresourcequota.crd.yaml":           {assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYaml, map[string]*bintree{}},
			"0000_03_security-openshift_01_scc.crd.yaml":                         {assetsCrd0000_03_securityOpenshift_01_sccCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_apiserver.crd.yaml":                      {assetsCrd0000_10_configOperator_01_apiserverCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_authentication.crd.yaml":                 {assetsCrd0000_10_configOperator_01_authenticationCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_build.crd.yaml":                          {assetsCrd0000_10_configOperator_01_buildCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_console.crd.yaml":                        {assetsCrd0000_10_configOperator_01_consoleCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_dns.crd.yaml":                            {assetsCrd0000_10_configOperator_01_dnsCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_featuregate.crd.yaml":                    {assetsCrd0000_10_configOperator_01_featuregateCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_image.crd.yaml":                          {assetsCrd0000_10_configOperator_01_imageCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_imagecontentsourcepolicy.crd.yaml":       {assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_infrastructure.crd.yaml":                 {assetsCrd0000_10_configOperator_01_infrastructureCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_ingress.crd.yaml":                        {assetsCrd0000_10_configOperator_01_ingressCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_network.crd.yaml":                        {assetsCrd0000_10_configOperator_01_networkCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_oauth.crd.yaml":                          {assetsCrd0000_10_configOperator_01_oauthCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_project.crd.yaml":                        {assetsCrd0000_10_configOperator_01_projectCrdYaml, map[string]*bintree{}},
			"0000_11_imageregistry-configs.crd.yaml":                             {assetsCrd0000_11_imageregistryConfigsCrdYaml, map[string]*bintree{}},
			"0000_50_service-ca-operator_02_crd.yaml":                            {assetsCrd0000_50_serviceCaOperator_02_crdYaml, map[string]*bintree{}},
			"0000_70_dns-operator_00-custom-resource-definition.yaml":            {assetsCrd0000_70_dnsOperator_00CustomResourceDefinitionYaml, map[string]*bintree{}},
			"cluster-ingress-00-custom-resource-definition.yaml":                 {assetsCrdClusterIngress00CustomResourceDefinitionYaml, map[string]*bintree{}},
			"cluster-network-01-crd.yml":                                         {assetsCrdClusterNetwork01CrdYml, map[string]*bintree{}},
		}},
	}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}
