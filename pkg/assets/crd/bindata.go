// Code generated for package assets by go-bindata DO NOT EDIT. (@generated)
// sources:
// assets/crd/0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml
// assets/crd/0000_03_config-operator_01_proxy.crd.yaml
// assets/crd/0000_03_quota-openshift_01_clusterresourcequota.crd.yaml
// assets/crd/0000_03_security-openshift_01_scc.crd.yaml
// assets/crd/0000_10_config-operator_01_build.crd.yaml
// assets/crd/0000_10_config-operator_01_featuregate.crd.yaml
// assets/crd/0000_10_config-operator_01_image.crd.yaml
// assets/crd/0000_10_config-operator_01_imagecontentsourcepolicy.crd.yaml
// assets/crd/0000_11_imageregistry-configs.crd.yaml
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
	"assets/crd/0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml": assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml,
	"assets/crd/0000_03_config-operator_01_proxy.crd.yaml":                          assetsCrd0000_03_configOperator_01_proxyCrdYaml,
	"assets/crd/0000_03_quota-openshift_01_clusterresourcequota.crd.yaml":           assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYaml,
	"assets/crd/0000_03_security-openshift_01_scc.crd.yaml":                         assetsCrd0000_03_securityOpenshift_01_sccCrdYaml,
	"assets/crd/0000_10_config-operator_01_build.crd.yaml":                          assetsCrd0000_10_configOperator_01_buildCrdYaml,
	"assets/crd/0000_10_config-operator_01_featuregate.crd.yaml":                    assetsCrd0000_10_configOperator_01_featuregateCrdYaml,
	"assets/crd/0000_10_config-operator_01_image.crd.yaml":                          assetsCrd0000_10_configOperator_01_imageCrdYaml,
	"assets/crd/0000_10_config-operator_01_imagecontentsourcepolicy.crd.yaml":       assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYaml,
	"assets/crd/0000_11_imageregistry-configs.crd.yaml":                             assetsCrd0000_11_imageregistryConfigsCrdYaml,
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
			"0000_03_authorization-openshift_01_rolebindingrestriction.crd.yaml": {assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml, map[string]*bintree{}},
			"0000_03_config-operator_01_proxy.crd.yaml":                          {assetsCrd0000_03_configOperator_01_proxyCrdYaml, map[string]*bintree{}},
			"0000_03_quota-openshift_01_clusterresourcequota.crd.yaml":           {assetsCrd0000_03_quotaOpenshift_01_clusterresourcequotaCrdYaml, map[string]*bintree{}},
			"0000_03_security-openshift_01_scc.crd.yaml":                         {assetsCrd0000_03_securityOpenshift_01_sccCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_build.crd.yaml":                          {assetsCrd0000_10_configOperator_01_buildCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_featuregate.crd.yaml":                    {assetsCrd0000_10_configOperator_01_featuregateCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_image.crd.yaml":                          {assetsCrd0000_10_configOperator_01_imageCrdYaml, map[string]*bintree{}},
			"0000_10_config-operator_01_imagecontentsourcepolicy.crd.yaml":       {assetsCrd0000_10_configOperator_01_imagecontentsourcepolicyCrdYaml, map[string]*bintree{}},
			"0000_11_imageregistry-configs.crd.yaml":                             {assetsCrd0000_11_imageregistryConfigsCrdYaml, map[string]*bintree{}},
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
