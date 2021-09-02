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

var _assetsCrd0000_03_authorizationOpenshift_01_rolebindingrestrictionCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: rolebindingrestrictions.authorization.openshift.io
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
spec:
  group: authorization.openshift.io
  scope: Namespaced
  names:
    kind: RoleBindingRestriction
    listKind: RoleBindingRestrictionList
    plural: rolebindingrestrictions
    singular: rolebindingrestriction
  versions:
  - name: v1
    served: true
    storage: true
    schema:
      openAPIV3Schema:
        description: RoleBindingRestriction is an object that can be matched against
          a subject (user, group, or service account) to determine whether rolebindings
          on that subject are allowed in the namespace to which the RoleBindingRestriction
          belongs.  If any one of those RoleBindingRestriction objects matches a subject,
          rolebindings on that subject in the namespace are allowed.
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
            description: Spec defines the matcher.
            type: object
            properties:
              grouprestriction:
                description: GroupRestriction matches against group subjects.
                type: object
                properties:
                  groups:
                    description: Groups is a list of groups used to match against
                      an individual user's groups. If the user is a member of one
                      of the whitelisted groups, the user is allowed to be bound to
                      a role.
                    type: array
                    items:
                      type: string
                    nullable: true
                  labels:
                    description: Selectors specifies a list of label selectors over
                      group labels.
                    type: array
                    items:
                      description: A label selector is a label query over a set of
                        resources. The result of matchLabels and matchExpressions
                        are ANDed. An empty label selector matches all objects. A
                        null label selector matches no objects.
                      type: object
                      properties:
                        matchExpressions:
                          description: matchExpressions is a list of label selector
                            requirements. The requirements are ANDed.
                          type: array
                          items:
                            description: A label selector requirement is a selector
                              that contains values, a key, and an operator that relates
                              the key and values.
                            type: object
                            required:
                            - key
                            - operator
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
                                description: values is an array of string values.
                                  If the operator is In or NotIn, the values array
                                  must be non-empty. If the operator is Exists or
                                  DoesNotExist, the values array must be empty. This
                                  array is replaced during a strategic merge patch.
                                type: array
                                items:
                                  type: string
                        matchLabels:
                          description: matchLabels is a map of {key,value} pairs.
                            A single {key,value} in the matchLabels map is equivalent
                            to an element of matchExpressions, whose key field is
                            "key", the operator is "In", and the values array contains
                            only "value". The requirements are ANDed.
                          type: object
                          additionalProperties:
                            type: string
                    nullable: true
                nullable: true
              serviceaccountrestriction:
                description: ServiceAccountRestriction matches against service-account
                  subjects.
                type: object
                properties:
                  namespaces:
                    description: Namespaces specifies a list of literal namespace
                      names.
                    type: array
                    items:
                      type: string
                  serviceaccounts:
                    description: ServiceAccounts specifies a list of literal service-account
                      names.
                    type: array
                    items:
                      description: ServiceAccountReference specifies a service account
                        and namespace by their names.
                      type: object
                      properties:
                        name:
                          description: Name is the name of the service account.
                          type: string
                        namespace:
                          description: Namespace is the namespace of the service account.  Service
                            accounts from inside the whitelisted namespaces are allowed
                            to be bound to roles.  If Namespace is empty, then the
                            namespace of the RoleBindingRestriction in which the ServiceAccountReference
                            is embedded is used.
                          type: string
                nullable: true
              userrestriction:
                description: UserRestriction matches against user subjects.
                type: object
                properties:
                  groups:
                    description: Groups specifies a list of literal group names.
                    type: array
                    items:
                      type: string
                    nullable: true
                  labels:
                    description: Selectors specifies a list of label selectors over
                      user labels.
                    type: array
                    items:
                      description: A label selector is a label query over a set of
                        resources. The result of matchLabels and matchExpressions
                        are ANDed. An empty label selector matches all objects. A
                        null label selector matches no objects.
                      type: object
                      properties:
                        matchExpressions:
                          description: matchExpressions is a list of label selector
                            requirements. The requirements are ANDed.
                          type: array
                          items:
                            description: A label selector requirement is a selector
                              that contains values, a key, and an operator that relates
                              the key and values.
                            type: object
                            required:
                            - key
                            - operator
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
                                description: values is an array of string values.
                                  If the operator is In or NotIn, the values array
                                  must be non-empty. If the operator is Exists or
                                  DoesNotExist, the values array must be empty. This
                                  array is replaced during a strategic merge patch.
                                type: array
                                items:
                                  type: string
                        matchLabels:
                          description: matchLabels is a map of {key,value} pairs.
                            A single {key,value} in the matchLabels map is equivalent
                            to an element of matchExpressions, whose key field is
                            "key", the operator is "In", and the values array contains
                            only "value". The requirements are ANDed.
                          type: object
                          additionalProperties:
                            type: string
                    nullable: true
                  users:
                    description: Users specifies a list of literal user names.
                    type: array
                    items:
                      type: string
                nullable: true
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

var _assetsCrd0000_03_configOperator_01_proxyCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: proxies.config.openshift.io
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
spec:
  group: config.openshift.io
  scope: Cluster
  names:
    kind: Proxy
    listKind: ProxyList
    plural: proxies
    singular: proxy
  versions:
  - name: v1
    served: true
    storage: true
    subresources:
      status: {}
    schema:
      openAPIV3Schema:
        description: Proxy holds cluster-wide information on how to configure default
          proxies for the cluster. The canonical name is ` + "`" + `cluster` + "`" + `
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
            description: Spec holds user-settable values for the proxy configuration
            type: object
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
                description: noProxy is a comma-separated list of hostnames and/or
                  CIDRs for which the proxy should not be used. Empty means unset
                  and will not result in an env var.
                type: string
              readinessEndpoints:
                description: readinessEndpoints is a list of endpoints used to verify
                  readiness of the proxy.
                type: array
                items:
                  type: string
              trustedCA:
                description: "trustedCA is a reference to a ConfigMap containing a
                  CA certificate bundle. The trustedCA field should only be consumed
                  by a proxy validator. The validator is responsible for reading the
                  certificate bundle from the required key \"ca-bundle.crt\", merging
                  it with the system default trust bundle, and writing the merged
                  trust bundle to a ConfigMap named \"trusted-ca-bundle\" in the \"openshift-config-managed\"
                  namespace. Clients that expect to make proxy connections must use
                  the trusted-ca-bundle for all HTTPS requests to the proxy, and may
                  use the trusted-ca-bundle for non-proxy HTTPS requests as well.
                  \n The namespace for the ConfigMap referenced by trustedCA is \"openshift-config\".
                  Here is an example ConfigMap (in yaml): \n apiVersion: v1 kind:
                  ConfigMap metadata:  name: user-ca-bundle  namespace: openshift-config
                  \ data:    ca-bundle.crt: |      -----BEGIN CERTIFICATE-----      Custom
                  CA certificate bundle.      -----END CERTIFICATE-----"
                type: object
                required:
                - name
                properties:
                  name:
                    description: name is the metadata.name of the referenced config
                      map
                    type: string
          status:
            description: status holds observed values from the cluster. They may not
              be overridden.
            type: object
            properties:
              httpProxy:
                description: httpProxy is the URL of the proxy for HTTP requests.
                type: string
              httpsProxy:
                description: httpsProxy is the URL of the proxy for HTTPS requests.
                type: string
              noProxy:
                description: noProxy is a comma-separated list of hostnames and/or
                  CIDRs for which the proxy should not be used.
                type: string
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
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
  name: clusterresourcequotas.quota.openshift.io
spec:
  group: quota.openshift.io
  names:
    kind: ClusterResourceQuota
    listKind: ClusterResourceQuotaList
    plural: clusterresourcequotas
    singular: clusterresourcequota
  preserveUnknownFields: false
  scope: Cluster
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      description: ClusterResourceQuota mirrors ResourceQuota at a cluster scope.  This
        object is easily convertible to synthetic ResourceQuota object to allow quota
        evaluation re-use.
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
          description: Spec defines the desired quota
          properties:
            quota:
              description: Quota defines the desired quota
              properties:
                hard:
                  additionalProperties:
                    anyOf:
                    - type: integer
                    - type: string
                    pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                    type: ""
                    x-kubernetes-int-or-string: true
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
                          anyOf:
                          - type: integer
                          - type: string
                          pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                          x-kubernetes-int-or-string: true
                        description: 'Hard is the set of enforced hard limits for
                          each named resource. More info: https://kubernetes.io/docs/concepts/policy/resource-quotas/'
                        type: object
                      used:
                        additionalProperties:
                          anyOf:
                          - type: integer
                          - type: string
                          pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                          x-kubernetes-int-or-string: true
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
                    anyOf:
                    - type: integer
                    - type: string
                    pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                    x-kubernetes-int-or-string: true
                  description: 'Hard is the set of enforced hard limits for each named
                    resource. More info: https://kubernetes.io/docs/concepts/policy/resource-quotas/'
                  type: object
                used:
                  additionalProperties:
                    anyOf:
                    - type: integer
                    - type: string
                    pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                    x-kubernetes-int-or-string: true
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
  versions:
  - name: v1
    served: true
    storage: true
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

var _assetsCrd0000_03_securityOpenshift_01_sccCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: securitycontextconstraints.security.openshift.io
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
spec:
  group: security.openshift.io
  scope: Cluster
  names:
    kind: SecurityContextConstraints
    listKind: SecurityContextConstraintsList
    plural: securitycontextconstraints
    singular: securitycontextconstraints
  versions:
  - name: v1
    served: true
    storage: true
    additionalPrinterColumns:
    - name: Priv
      type: string
      jsonPath: .allowPrivilegedContainer
      description: Determines if a container can request to be run as privileged
    - name: Caps
      type: string
      jsonPath: .allowedCapabilities
      description: A list of capabilities that can be requested to add to the container
    - name: SELinux
      type: string
      jsonPath: .seLinuxContext.type
      description: Strategy that will dictate what labels will be set in the SecurityContext
    - name: RunAsUser
      type: string
      jsonPath: .runAsUser.type
      description: Strategy that will dictate what RunAsUser is used in the SecurityContext
    - name: FSGroup
      type: string
      jsonPath: .fsGroup.type
      description: Strategy that will dictate what fs group is used by the SecurityContext
    - name: SupGroup
      type: string
      jsonPath: .supplementalGroups.type
      description: Strategy that will dictate what supplemental groups are used by
        the SecurityContext
    - name: Priority
      type: string
      jsonPath: .priority
      description: Sort order of SCCs
    - name: ReadOnlyRootFS
      type: string
      jsonPath: .readOnlyRootFilesystem
      description: Force containers to run with a read only root file system
    - name: Volumes
      type: string
      jsonPath: .volumes
      description: White list of allowed volume plugins
    schema:
      openAPIV3Schema:
        description: SecurityContextConstraints governs the ability to make requests
          that affect the SecurityContext that will be applied to a container. For
          historical reasons SCC was exposed under the core Kubernetes API group.
          That exposure is deprecated and will be removed in a future release - users
          should instead use the security.openshift.io group to manage SecurityContextConstraints.
        type: object
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
        properties:
          allowHostDirVolumePlugin:
            description: AllowHostDirVolumePlugin determines if the policy allow containers
              to use the HostDir volume plugin
            type: boolean
          allowHostIPC:
            description: AllowHostIPC determines if the policy allows host ipc in
              the containers.
            type: boolean
          allowHostNetwork:
            description: AllowHostNetwork determines if the policy allows the use
              of HostNetwork in the pod spec.
            type: boolean
          allowHostPID:
            description: AllowHostPID determines if the policy allows host pid in
              the containers.
            type: boolean
          allowHostPorts:
            description: AllowHostPorts determines if the policy allows host ports
              in the containers.
            type: boolean
          allowPrivilegeEscalation:
            description: AllowPrivilegeEscalation determines if a pod can request
              to allow privilege escalation. If unspecified, defaults to true.
            type: boolean
            nullable: true
          allowPrivilegedContainer:
            description: AllowPrivilegedContainer determines if a container can request
              to be run as privileged.
            type: boolean
          allowedCapabilities:
            description: AllowedCapabilities is a list of capabilities that can be
              requested to add to the container. Capabilities in this field maybe
              added at the pod author's discretion. You must not list a capability
              in both AllowedCapabilities and RequiredDropCapabilities. To allow all
              capabilities you may use '*'.
            type: array
            items:
              description: Capability represent POSIX capabilities type
              type: string
            nullable: true
          allowedFlexVolumes:
            description: AllowedFlexVolumes is a whitelist of allowed Flexvolumes.  Empty
              or nil indicates that all Flexvolumes may be used.  This parameter is
              effective only when the usage of the Flexvolumes is allowed in the "Volumes"
              field.
            type: array
            items:
              description: AllowedFlexVolume represents a single Flexvolume that is
                allowed to be used.
              type: object
              required:
              - driver
              properties:
                driver:
                  description: Driver is the name of the Flexvolume driver.
                  type: string
            nullable: true
          allowedUnsafeSysctls:
            description: "AllowedUnsafeSysctls is a list of explicitly allowed unsafe
              sysctls, defaults to none. Each entry is either a plain sysctl name
              or ends in \"*\" in which case it is considered as a prefix of allowed
              sysctls. Single * means all unsafe sysctls are allowed. Kubelet has
              to whitelist all allowed unsafe sysctls explicitly to avoid rejection.
              \n Examples: e.g. \"foo/*\" allows \"foo/bar\", \"foo/baz\", etc. e.g.
              \"foo.*\" allows \"foo.bar\", \"foo.baz\", etc."
            type: array
            items:
              type: string
            nullable: true
          apiVersion:
            description: 'APIVersion defines the versioned schema of this representation
              of an object. Servers should convert recognized schemas to the latest
              internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#resources'
            type: string
          defaultAddCapabilities:
            description: DefaultAddCapabilities is the default set of capabilities
              that will be added to the container unless the pod spec specifically
              drops the capability.  You may not list a capabiility in both DefaultAddCapabilities
              and RequiredDropCapabilities.
            type: array
            items:
              description: Capability represent POSIX capabilities type
              type: string
            nullable: true
          defaultAllowPrivilegeEscalation:
            description: DefaultAllowPrivilegeEscalation controls the default setting
              for whether a process can gain more privileges than its parent process.
            type: boolean
            nullable: true
          forbiddenSysctls:
            description: "ForbiddenSysctls is a list of explicitly forbidden sysctls,
              defaults to none. Each entry is either a plain sysctl name or ends in
              \"*\" in which case it is considered as a prefix of forbidden sysctls.
              Single * means all sysctls are forbidden. \n Examples: e.g. \"foo/*\"
              forbids \"foo/bar\", \"foo/baz\", etc. e.g. \"foo.*\" forbids \"foo.bar\",
              \"foo.baz\", etc."
            type: array
            items:
              type: string
            nullable: true
          fsGroup:
            description: FSGroup is the strategy that will dictate what fs group is
              used by the SecurityContext.
            type: object
            properties:
              ranges:
                description: Ranges are the allowed ranges of fs groups.  If you would
                  like to force a single fs group then supply a single range with
                  the same start and end.
                type: array
                items:
                  description: 'IDRange provides a min/max of an allowed range of
                    IDs. TODO: this could be reused for UIDs.'
                  type: object
                  properties:
                    max:
                      description: Max is the end of the range, inclusive.
                      type: integer
                      format: int64
                    min:
                      description: Min is the start of the range, inclusive.
                      type: integer
                      format: int64
              type:
                description: Type is the strategy that will dictate what FSGroup is
                  used in the SecurityContext.
                type: string
            nullable: true
          groups:
            description: The groups that have permission to use this security context
              constraints
            type: array
            items:
              type: string
            nullable: true
          kind:
            description: 'Kind is a string value representing the REST resource this
              object represents. Servers may infer this from the endpoint the client
              submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/sig-architecture/api-conventions.md#types-kinds'
            type: string
          metadata:
            type: object
          priority:
            description: Priority influences the sort order of SCCs when evaluating
              which SCCs to try first for a given pod request based on access in the
              Users and Groups fields.  The higher the int, the higher priority. An
              unset value is considered a 0 priority. If scores for multiple SCCs
              are equal they will be sorted from most restrictive to least restrictive.
              If both priorities and restrictions are equal the SCCs will be sorted
              by name.
            type: integer
            format: int32
            nullable: true
          readOnlyRootFilesystem:
            description: ReadOnlyRootFilesystem when set to true will force containers
              to run with a read only root file system.  If the container specifically
              requests to run with a non-read only root file system the SCC should
              deny the pod. If set to false the container may run with a read only
              root file system if it wishes but it will not be forced to.
            type: boolean
          requiredDropCapabilities:
            description: RequiredDropCapabilities are the capabilities that will be
              dropped from the container.  These are required to be dropped and cannot
              be added.
            type: array
            items:
              description: Capability represent POSIX capabilities type
              type: string
            nullable: true
          runAsUser:
            description: RunAsUser is the strategy that will dictate what RunAsUser
              is used in the SecurityContext.
            type: object
            properties:
              type:
                description: Type is the strategy that will dictate what RunAsUser
                  is used in the SecurityContext.
                type: string
              uid:
                description: UID is the user id that containers must run as.  Required
                  for the MustRunAs strategy if not using namespace/service account
                  allocated uids.
                type: integer
                format: int64
              uidRangeMax:
                description: UIDRangeMax defines the max value for a strategy that
                  allocates by range.
                type: integer
                format: int64
              uidRangeMin:
                description: UIDRangeMin defines the min value for a strategy that
                  allocates by range.
                type: integer
                format: int64
            nullable: true
          seLinuxContext:
            description: SELinuxContext is the strategy that will dictate what labels
              will be set in the SecurityContext.
            type: object
            properties:
              seLinuxOptions:
                description: seLinuxOptions required to run as; required for MustRunAs
                type: object
                properties:
                  level:
                    description: Level is SELinux level label that applies to the
                      container.
                    type: string
                  role:
                    description: Role is a SELinux role label that applies to the
                      container.
                    type: string
                  type:
                    description: Type is a SELinux type label that applies to the
                      container.
                    type: string
                  user:
                    description: User is a SELinux user label that applies to the
                      container.
                    type: string
              type:
                description: Type is the strategy that will dictate what SELinux context
                  is used in the SecurityContext.
                type: string
            nullable: true
          seccompProfiles:
            description: "SeccompProfiles lists the allowed profiles that may be set
              for the pod or container's seccomp annotations.  An unset (nil) or empty
              value means that no profiles may be specifid by the pod or container.\tThe
              wildcard '*' may be used to allow all profiles.  When used to generate
              a value for a pod the first non-wildcard profile will be used as the
              default."
            type: array
            items:
              type: string
            nullable: true
          supplementalGroups:
            description: SupplementalGroups is the strategy that will dictate what
              supplemental groups are used by the SecurityContext.
            type: object
            properties:
              ranges:
                description: Ranges are the allowed ranges of supplemental groups.  If
                  you would like to force a single supplemental group then supply
                  a single range with the same start and end.
                type: array
                items:
                  description: 'IDRange provides a min/max of an allowed range of
                    IDs. TODO: this could be reused for UIDs.'
                  type: object
                  properties:
                    max:
                      description: Max is the end of the range, inclusive.
                      type: integer
                      format: int64
                    min:
                      description: Min is the start of the range, inclusive.
                      type: integer
                      format: int64
              type:
                description: Type is the strategy that will dictate what supplemental
                  groups is used in the SecurityContext.
                type: string
            nullable: true
          users:
            description: The users who have permissions to use this security context
              constraints
            type: array
            items:
              type: string
            nullable: true
          volumes:
            description: Volumes is a white list of allowed volume plugins.  FSType
              corresponds directly with the field names of a VolumeSource (azureFile,
              configMap, emptyDir).  To allow all volumes you may use "*". To allow
              no volumes, set to ["none"].
            type: array
            items:
              description: FS Type gives strong typing to different file systems that
                are used by volumes.
              type: string
            nullable: true
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
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
spec:
  group: config.openshift.io
  scope: Cluster
  preserveUnknownFields: false
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
  "validation":
    "openAPIV3Schema":
      description: "Build configures the behavior of OpenShift builds for the entire
        cluster. This includes default settings that can be overridden in BuildConfig
        objects, and overrides which are applied to all builds. \n The canonical name
        is \"cluster\""
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
          description: Spec holds user-settable values for the build controller configuration
          type: object
          properties:
            additionalTrustedCA:
              description: "AdditionalTrustedCA is a reference to a ConfigMap containing
                additional CAs that should be trusted for image pushes and pulls during
                builds. The namespace for this config map is openshift-config. \n
                DEPRECATED: Additional CAs for image pull and push should be set on
                image.config.openshift.io/cluster instead."
              type: object
              required:
              - name
              properties:
                name:
                  description: name is the metadata.name of the referenced config
                    map
                  type: string
            buildDefaults:
              description: BuildDefaults controls the default information for Builds
              type: object
              properties:
                defaultProxy:
                  description: "DefaultProxy contains the default proxy settings for
                    all build operations, including image pull/push and source download.
                    \n Values can be overrode by setting the ` + "`" + `HTTP_PROXY` + "`" + `, ` + "`" + `HTTPS_PROXY` + "`" + `,
                    and ` + "`" + `NO_PROXY` + "`" + ` environment variables in the build config's strategy."
                  type: object
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
                      type: array
                      items:
                        type: string
                    trustedCA:
                      description: "trustedCA is a reference to a ConfigMap containing
                        a CA certificate bundle. The trustedCA field should only be
                        consumed by a proxy validator. The validator is responsible
                        for reading the certificate bundle from the required key \"ca-bundle.crt\",
                        merging it with the system default trust bundle, and writing
                        the merged trust bundle to a ConfigMap named \"trusted-ca-bundle\"
                        in the \"openshift-config-managed\" namespace. Clients that
                        expect to make proxy connections must use the trusted-ca-bundle
                        for all HTTPS requests to the proxy, and may use the trusted-ca-bundle
                        for non-proxy HTTPS requests as well. \n The namespace for
                        the ConfigMap referenced by trustedCA is \"openshift-config\".
                        Here is an example ConfigMap (in yaml): \n apiVersion: v1
                        kind: ConfigMap metadata:  name: user-ca-bundle  namespace:
                        openshift-config  data:    ca-bundle.crt: |      -----BEGIN
                        CERTIFICATE-----      Custom CA certificate bundle.      -----END
                        CERTIFICATE-----"
                      type: object
                      required:
                      - name
                      properties:
                        name:
                          description: name is the metadata.name of the referenced
                            config map
                          type: string
                env:
                  description: Env is a set of default environment variables that
                    will be applied to the build if the specified variables do not
                    exist on the build
                  type: array
                  items:
                    description: EnvVar represents an environment variable present
                      in a Container.
                    type: object
                    required:
                    - name
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
                        type: object
                        properties:
                          configMapKeyRef:
                            description: Selects a key of a ConfigMap.
                            type: object
                            required:
                            - key
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
                                description: Specify whether the ConfigMap or its
                                  key must be defined
                                type: boolean
                          fieldRef:
                            description: 'Selects a field of the pod: supports metadata.name,
                              metadata.namespace, ` + "`" + `metadata.labels[''<KEY>'']` + "`" + `, ` + "`" + `metadata.annotations[''<KEY>'']` + "`" + `,
                              spec.nodeName, spec.serviceAccountName, status.hostIP,
                              status.podIP, status.podIPs.'
                            type: object
                            required:
                            - fieldPath
                            properties:
                              apiVersion:
                                description: Version of the schema the FieldPath is
                                  written in terms of, defaults to "v1".
                                type: string
                              fieldPath:
                                description: Path of the field to select in the specified
                                  API version.
                                type: string
                          resourceFieldRef:
                            description: 'Selects a resource of the container: only
                              resources limits and requests (limits.cpu, limits.memory,
                              limits.ephemeral-storage, requests.cpu, requests.memory
                              and requests.ephemeral-storage) are currently supported.'
                            type: object
                            required:
                            - resource
                            properties:
                              containerName:
                                description: 'Container name: required for volumes,
                                  optional for env vars'
                                type: string
                              divisor:
                                description: Specifies the output format of the exposed
                                  resources, defaults to "1"
                                pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                                anyOf:
                                - type: integer
                                - type: string
                                x-kubernetes-int-or-string: true
                              resource:
                                description: 'Required: resource to select'
                                type: string
                          secretKeyRef:
                            description: Selects a key of a secret in the pod's namespace
                            type: object
                            required:
                            - key
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
                                description: Specify whether the Secret or its key
                                  must be defined
                                type: boolean
                gitProxy:
                  description: "GitProxy contains the proxy settings for git operations
                    only. If set, this will override any Proxy settings for all git
                    commands, such as git clone. \n Values that are not set here will
                    be inherited from DefaultProxy."
                  type: object
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
                      type: array
                      items:
                        type: string
                    trustedCA:
                      description: "trustedCA is a reference to a ConfigMap containing
                        a CA certificate bundle. The trustedCA field should only be
                        consumed by a proxy validator. The validator is responsible
                        for reading the certificate bundle from the required key \"ca-bundle.crt\",
                        merging it with the system default trust bundle, and writing
                        the merged trust bundle to a ConfigMap named \"trusted-ca-bundle\"
                        in the \"openshift-config-managed\" namespace. Clients that
                        expect to make proxy connections must use the trusted-ca-bundle
                        for all HTTPS requests to the proxy, and may use the trusted-ca-bundle
                        for non-proxy HTTPS requests as well. \n The namespace for
                        the ConfigMap referenced by trustedCA is \"openshift-config\".
                        Here is an example ConfigMap (in yaml): \n apiVersion: v1
                        kind: ConfigMap metadata:  name: user-ca-bundle  namespace:
                        openshift-config  data:    ca-bundle.crt: |      -----BEGIN
                        CERTIFICATE-----      Custom CA certificate bundle.      -----END
                        CERTIFICATE-----"
                      type: object
                      required:
                      - name
                      properties:
                        name:
                          description: name is the metadata.name of the referenced
                            config map
                          type: string
                imageLabels:
                  description: ImageLabels is a list of docker labels that are applied
                    to the resulting image. User can override a default label by providing
                    a label with the same name in their Build/BuildConfig.
                  type: array
                  items:
                    type: object
                    properties:
                      name:
                        description: Name defines the name of the label. It must have
                          non-zero length.
                        type: string
                      value:
                        description: Value defines the literal value of the label.
                        type: string
                resources:
                  description: Resources defines resource requirements to execute
                    the build.
                  type: object
                  properties:
                    limits:
                      description: 'Limits describes the maximum amount of compute
                        resources allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                      type: object
                      additionalProperties:
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        anyOf:
                        - type: integer
                        - type: string
                        x-kubernetes-int-or-string: true
                    requests:
                      description: 'Requests describes the minimum amount of compute
                        resources required. If Requests is omitted for a container,
                        it defaults to Limits if that is explicitly specified, otherwise
                        to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                      type: object
                      additionalProperties:
                        pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                        anyOf:
                        - type: integer
                        - type: string
                        x-kubernetes-int-or-string: true
            buildOverrides:
              description: BuildOverrides controls override settings for builds
              type: object
              properties:
                forcePull:
                  description: ForcePull overrides, if set, the equivalent value in
                    the builds, i.e. false disables force pull for all builds, true
                    enables force pull for all builds, independently of what each
                    build specifies itself
                  type: boolean
                imageLabels:
                  description: ImageLabels is a list of docker labels that are applied
                    to the resulting image. If user provided a label in their Build/BuildConfig
                    with the same name as one in this list, the user's label will
                    be overwritten.
                  type: array
                  items:
                    type: object
                    properties:
                      name:
                        description: Name defines the name of the label. It must have
                          non-zero length.
                        type: string
                      value:
                        description: Value defines the literal value of the label.
                        type: string
                nodeSelector:
                  description: NodeSelector is a selector which must be true for the
                    build pod to fit on a node
                  type: object
                  additionalProperties:
                    type: string
                tolerations:
                  description: Tolerations is a list of Tolerations that will override
                    any existing tolerations set on a build pod.
                  type: array
                  items:
                    description: The pod this Toleration is attached to tolerates
                      any taint that matches the triple <key,value,effect> using the
                      matching operator <operator>.
                    type: object
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
                        type: integer
                        format: int64
                      value:
                        description: Value is the taint value the toleration matches
                          to. If the operator is Exists, the value should be empty,
                          otherwise just a regular string.
                        type: string
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

var _assetsCrd0000_10_configOperator_01_featuregateCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: featuregates.config.openshift.io
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
spec:
  group: config.openshift.io
  scope: Cluster
  names:
    kind: FeatureGate
    listKind: FeatureGateList
    plural: featuregates
    singular: featuregate
  versions:
  - name: v1
    served: true
    storage: true
    subresources:
      status: {}
    schema:
      openAPIV3Schema:
        description: Feature holds cluster-wide information about feature gates.  The
          canonical name is ` + "`" + `cluster` + "`" + `
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
            description: spec holds user settable values for configuration
            type: object
            properties:
              customNoUpgrade:
                description: customNoUpgrade allows the enabling or disabling of any
                  feature. Turning this feature set on IS NOT SUPPORTED, CANNOT BE
                  UNDONE, and PREVENTS UPGRADES. Because of its nature, this setting
                  cannot be validated.  If you have any typos or accidentally apply
                  invalid combinations your cluster may fail in an unrecoverable way.  featureSet
                  must equal "CustomNoUpgrade" must be set to use this field.
                type: object
                properties:
                  disabled:
                    description: disabled is a list of all feature gates that you
                      want to force off
                    type: array
                    items:
                      type: string
                  enabled:
                    description: enabled is a list of all feature gates that you want
                      to force on
                    type: array
                    items:
                      type: string
                nullable: true
              featureSet:
                description: featureSet changes the list of features in the cluster.  The
                  default is empty.  Be very careful adjusting this setting. Turning
                  on or off features may cause irreversible changes in your cluster
                  which cannot be undone.
                type: string
          status:
            description: status holds observed values from the cluster. They may not
              be overridden.
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
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
spec:
  group: config.openshift.io
  scope: Cluster
  preserveUnknownFields: false
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
  "validation":
    "openAPIV3Schema":
      description: Image governs policies related to imagestream imports and runtime
        configuration for external registries. It allows cluster admins to configure
        which registries OpenShift is allowed to import images from, extra CA trust
        bundles for external registries, and policies to block or allow registry hostnames.
        When exposing OpenShift's image registry to the public, this also lets cluster
        admins specify the external hostname.
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
          description: spec holds user settable values for configuration
          type: object
          properties:
            additionalTrustedCA:
              description: additionalTrustedCA is a reference to a ConfigMap containing
                additional CAs that should be trusted during imagestream import, pod
                image pull, build image pull, and imageregistry pullthrough. The namespace
                for this config map is openshift-config.
              type: object
              required:
              - name
              properties:
                name:
                  description: name is the metadata.name of the referenced config
                    map
                  type: string
            allowedRegistriesForImport:
              description: allowedRegistriesForImport limits the container image registries
                that normal users may import images from. Set this list to the registries
                that you trust to contain valid Docker images and that you want applications
                to be able to import from. Users with permission to create Images
                or ImageStreamMappings via the API are not affected by this policy
                - typically only administrators or system integrations will have those
                permissions.
              type: array
              items:
                description: RegistryLocation contains a location of the registry
                  specified by the registry domain name. The domain name might include
                  wildcards, like '*' or '??'.
                type: object
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
            externalRegistryHostnames:
              description: externalRegistryHostnames provides the hostnames for the
                default external image registry. The external hostname should be set
                only when the image registry is exposed externally. The first value
                is used in 'publicDockerImageRepository' field in ImageStreams. The
                value must be in "hostname[:port]" format.
              type: array
              items:
                type: string
            registrySources:
              description: registrySources contains configuration that determines
                how the container runtime should treat individual registries when
                accessing images for builds+pods. (e.g. whether or not to allow insecure
                access).  It does not contain configuration for the internal cluster
                registry.
              type: object
              properties:
                allowedRegistries:
                  description: "allowedRegistries are the only registries permitted
                    for image pull and push actions. All other registries are denied.
                    \n Only one of BlockedRegistries or AllowedRegistries may be set."
                  type: array
                  items:
                    type: string
                blockedRegistries:
                  description: "blockedRegistries cannot be used for image pull and
                    push actions. All other registries are permitted. \n Only one
                    of BlockedRegistries or AllowedRegistries may be set."
                  type: array
                  items:
                    type: string
                containerRuntimeSearchRegistries:
                  description: 'containerRuntimeSearchRegistries are registries that
                    will be searched when pulling images that do not have fully qualified
                    domains in their pull specs. Registries will be searched in the
                    order provided in the list. Note: this search list only works
                    with the container runtime, i.e CRI-O. Will NOT work with builds
                    or imagestream imports.'
                  type: array
                  format: hostname
                  minItems: 1
                  items:
                    type: string
                  x-kubernetes-list-type: set
                insecureRegistries:
                  description: insecureRegistries are registries which do not have
                    a valid TLS certificates or only support HTTP connections.
                  type: array
                  items:
                    type: string
        status:
          description: status holds observed values from the cluster. They may not
            be overridden.
          type: object
          properties:
            externalRegistryHostnames:
              description: externalRegistryHostnames provides the hostnames for the
                default external image registry. The external hostname should be set
                only when the image registry is exposed externally. The first value
                is used in 'publicDockerImageRepository' field in ImageStreams. The
                value must be in "hostname[:port]" format.
              type: array
              items:
                type: string
            internalRegistryHostname:
              description: internalRegistryHostname sets the hostname for the default
                internal image registry. The value must be in "hostname[:port]" format.
                This value is set by the image registry operator which controls the
                internal registry hostname. For backward compatibility, users can
                still use OPENSHIFT_DEFAULT_REGISTRY environment variable but this
                setting overrides the environment variable.
              type: string
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
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
spec:
  group: operator.openshift.io
  scope: Cluster
  preserveUnknownFields: false
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
  "validation":
    "openAPIV3Schema":
      description: ImageContentSourcePolicy holds cluster-wide information about how
        to handle registry mirror rules. When multiple policies are defined, the outcome
        of the behavior is defined on each field.
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
          description: spec holds user settable values for configuration
          type: object
          properties:
            repositoryDigestMirrors:
              description: "repositoryDigestMirrors allows images referenced by image
                digests in pods to be pulled from alternative mirrored repository
                locations. The image pull specification provided to the pod will be
                compared to the source locations described in RepositoryDigestMirrors
                and the image may be pulled down from any of the mirrors in the list
                instead of the specified repository allowing administrators to choose
                a potentially faster mirror. Only image pull specifications that have
                an image digest will have this behavior applied to them - tags will
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
              type: array
              items:
                description: 'RepositoryDigestMirrors holds cluster-wide information
                  about how to handle mirros in the registries config. Note: the mirrors
                  only work when pulling the images that are referenced by their digests.'
                type: object
                required:
                - source
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
                    type: array
                    items:
                      type: string
                  source:
                    description: source is the repository that users refer to, e.g.
                      in image pull specifications.
                    type: string
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

var _assetsCrd0000_11_imageregistryConfigsCrdYaml = []byte(`apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: configs.imageregistry.operator.openshift.io
  annotations:
    include.release.openshift.io/ibm-cloud-managed: "true"
    include.release.openshift.io/self-managed-high-availability: "true"
    include.release.openshift.io/single-node-developer: "true"
spec:
  group: imageregistry.operator.openshift.io
  scope: Cluster
  versions:
  - name: v1
    served: true
    storage: true
    subresources:
      status: {}
    "schema":
      "openAPIV3Schema":
        description: Config is the configuration object for a registry instance managed
          by the registry operator
        type: object
        required:
        - metadata
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
            description: ImageRegistrySpec defines the specs for the running registry.
            type: object
            required:
            - managementState
            - replicas
            properties:
              affinity:
                description: affinity is a group of node affinity scheduling rules
                  for the image registry pod(s).
                type: object
                properties:
                  nodeAffinity:
                    description: Describes node affinity scheduling rules for the
                      pod.
                    type: object
                    properties:
                      preferredDuringSchedulingIgnoredDuringExecution:
                        description: The scheduler will prefer to schedule pods to
                          nodes that satisfy the affinity expressions specified by
                          this field, but it may choose a node that violates one or
                          more of the expressions. The node that is most preferred
                          is the one with the greatest sum of weights, i.e. for each
                          node that meets all of the scheduling requirements (resource
                          request, requiredDuringScheduling affinity expressions,
                          etc.), compute a sum by iterating through the elements of
                          this field and adding "weight" to the sum if the node matches
                          the corresponding matchExpressions; the node(s) with the
                          highest sum are the most preferred.
                        type: array
                        items:
                          description: An empty preferred scheduling term matches
                            all objects with implicit weight 0 (i.e. it's a no-op).
                            A null preferred scheduling term matches no objects (i.e.
                            is also a no-op).
                          type: object
                          required:
                          - preference
                          - weight
                          properties:
                            preference:
                              description: A node selector term, associated with the
                                corresponding weight.
                              type: object
                              properties:
                                matchExpressions:
                                  description: A list of node selector requirements
                                    by node's labels.
                                  type: array
                                  items:
                                    description: A node selector requirement is a
                                      selector that contains values, a key, and an
                                      operator that relates the key and values.
                                    type: object
                                    required:
                                    - key
                                    - operator
                                    properties:
                                      key:
                                        description: The label key that the selector
                                          applies to.
                                        type: string
                                      operator:
                                        description: Represents a key's relationship
                                          to a set of values. Valid operators are
                                          In, NotIn, Exists, DoesNotExist. Gt, and
                                          Lt.
                                        type: string
                                      values:
                                        description: An array of string values. If
                                          the operator is In or NotIn, the values
                                          array must be non-empty. If the operator
                                          is Exists or DoesNotExist, the values array
                                          must be empty. If the operator is Gt or
                                          Lt, the values array must have a single
                                          element, which will be interpreted as an
                                          integer. This array is replaced during a
                                          strategic merge patch.
                                        type: array
                                        items:
                                          type: string
                                matchFields:
                                  description: A list of node selector requirements
                                    by node's fields.
                                  type: array
                                  items:
                                    description: A node selector requirement is a
                                      selector that contains values, a key, and an
                                      operator that relates the key and values.
                                    type: object
                                    required:
                                    - key
                                    - operator
                                    properties:
                                      key:
                                        description: The label key that the selector
                                          applies to.
                                        type: string
                                      operator:
                                        description: Represents a key's relationship
                                          to a set of values. Valid operators are
                                          In, NotIn, Exists, DoesNotExist. Gt, and
                                          Lt.
                                        type: string
                                      values:
                                        description: An array of string values. If
                                          the operator is In or NotIn, the values
                                          array must be non-empty. If the operator
                                          is Exists or DoesNotExist, the values array
                                          must be empty. If the operator is Gt or
                                          Lt, the values array must have a single
                                          element, which will be interpreted as an
                                          integer. This array is replaced during a
                                          strategic merge patch.
                                        type: array
                                        items:
                                          type: string
                            weight:
                              description: Weight associated with matching the corresponding
                                nodeSelectorTerm, in the range 1-100.
                              type: integer
                              format: int32
                      requiredDuringSchedulingIgnoredDuringExecution:
                        description: If the affinity requirements specified by this
                          field are not met at scheduling time, the pod will not be
                          scheduled onto the node. If the affinity requirements specified
                          by this field cease to be met at some point during pod execution
                          (e.g. due to an update), the system may or may not try to
                          eventually evict the pod from its node.
                        type: object
                        required:
                        - nodeSelectorTerms
                        properties:
                          nodeSelectorTerms:
                            description: Required. A list of node selector terms.
                              The terms are ORed.
                            type: array
                            items:
                              description: A null or empty node selector term matches
                                no objects. The requirements of them are ANDed. The
                                TopologySelectorTerm type implements a subset of the
                                NodeSelectorTerm.
                              type: object
                              properties:
                                matchExpressions:
                                  description: A list of node selector requirements
                                    by node's labels.
                                  type: array
                                  items:
                                    description: A node selector requirement is a
                                      selector that contains values, a key, and an
                                      operator that relates the key and values.
                                    type: object
                                    required:
                                    - key
                                    - operator
                                    properties:
                                      key:
                                        description: The label key that the selector
                                          applies to.
                                        type: string
                                      operator:
                                        description: Represents a key's relationship
                                          to a set of values. Valid operators are
                                          In, NotIn, Exists, DoesNotExist. Gt, and
                                          Lt.
                                        type: string
                                      values:
                                        description: An array of string values. If
                                          the operator is In or NotIn, the values
                                          array must be non-empty. If the operator
                                          is Exists or DoesNotExist, the values array
                                          must be empty. If the operator is Gt or
                                          Lt, the values array must have a single
                                          element, which will be interpreted as an
                                          integer. This array is replaced during a
                                          strategic merge patch.
                                        type: array
                                        items:
                                          type: string
                                matchFields:
                                  description: A list of node selector requirements
                                    by node's fields.
                                  type: array
                                  items:
                                    description: A node selector requirement is a
                                      selector that contains values, a key, and an
                                      operator that relates the key and values.
                                    type: object
                                    required:
                                    - key
                                    - operator
                                    properties:
                                      key:
                                        description: The label key that the selector
                                          applies to.
                                        type: string
                                      operator:
                                        description: Represents a key's relationship
                                          to a set of values. Valid operators are
                                          In, NotIn, Exists, DoesNotExist. Gt, and
                                          Lt.
                                        type: string
                                      values:
                                        description: An array of string values. If
                                          the operator is In or NotIn, the values
                                          array must be non-empty. If the operator
                                          is Exists or DoesNotExist, the values array
                                          must be empty. If the operator is Gt or
                                          Lt, the values array must have a single
                                          element, which will be interpreted as an
                                          integer. This array is replaced during a
                                          strategic merge patch.
                                        type: array
                                        items:
                                          type: string
                  podAffinity:
                    description: Describes pod affinity scheduling rules (e.g. co-locate
                      this pod in the same node, zone, etc. as some other pod(s)).
                    type: object
                    properties:
                      preferredDuringSchedulingIgnoredDuringExecution:
                        description: The scheduler will prefer to schedule pods to
                          nodes that satisfy the affinity expressions specified by
                          this field, but it may choose a node that violates one or
                          more of the expressions. The node that is most preferred
                          is the one with the greatest sum of weights, i.e. for each
                          node that meets all of the scheduling requirements (resource
                          request, requiredDuringScheduling affinity expressions,
                          etc.), compute a sum by iterating through the elements of
                          this field and adding "weight" to the sum if the node has
                          pods which matches the corresponding podAffinityTerm; the
                          node(s) with the highest sum are the most preferred.
                        type: array
                        items:
                          description: The weights of all of the matched WeightedPodAffinityTerm
                            fields are added per-node to find the most preferred node(s)
                          type: object
                          required:
                          - podAffinityTerm
                          - weight
                          properties:
                            podAffinityTerm:
                              description: Required. A pod affinity term, associated
                                with the corresponding weight.
                              type: object
                              required:
                              - topologyKey
                              properties:
                                labelSelector:
                                  description: A label query over a set of resources,
                                    in this case pods.
                                  type: object
                                  properties:
                                    matchExpressions:
                                      description: matchExpressions is a list of label
                                        selector requirements. The requirements are
                                        ANDed.
                                      type: array
                                      items:
                                        description: A label selector requirement
                                          is a selector that contains values, a key,
                                          and an operator that relates the key and
                                          values.
                                        type: object
                                        required:
                                        - key
                                        - operator
                                        properties:
                                          key:
                                            description: key is the label key that
                                              the selector applies to.
                                            type: string
                                          operator:
                                            description: operator represents a key's
                                              relationship to a set of values. Valid
                                              operators are In, NotIn, Exists and
                                              DoesNotExist.
                                            type: string
                                          values:
                                            description: values is an array of string
                                              values. If the operator is In or NotIn,
                                              the values array must be non-empty.
                                              If the operator is Exists or DoesNotExist,
                                              the values array must be empty. This
                                              array is replaced during a strategic
                                              merge patch.
                                            type: array
                                            items:
                                              type: string
                                    matchLabels:
                                      description: matchLabels is a map of {key,value}
                                        pairs. A single {key,value} in the matchLabels
                                        map is equivalent to an element of matchExpressions,
                                        whose key field is "key", the operator is
                                        "In", and the values array contains only "value".
                                        The requirements are ANDed.
                                      type: object
                                      additionalProperties:
                                        type: string
                                namespaces:
                                  description: namespaces specifies which namespaces
                                    the labelSelector applies to (matches against);
                                    null or empty list means "this pod's namespace"
                                  type: array
                                  items:
                                    type: string
                                topologyKey:
                                  description: This pod should be co-located (affinity)
                                    or not co-located (anti-affinity) with the pods
                                    matching the labelSelector in the specified namespaces,
                                    where co-located is defined as running on a node
                                    whose value of the label with key topologyKey
                                    matches that of any node on which any of the selected
                                    pods is running. Empty topologyKey is not allowed.
                                  type: string
                            weight:
                              description: weight associated with matching the corresponding
                                podAffinityTerm, in the range 1-100.
                              type: integer
                              format: int32
                      requiredDuringSchedulingIgnoredDuringExecution:
                        description: If the affinity requirements specified by this
                          field are not met at scheduling time, the pod will not be
                          scheduled onto the node. If the affinity requirements specified
                          by this field cease to be met at some point during pod execution
                          (e.g. due to a pod label update), the system may or may
                          not try to eventually evict the pod from its node. When
                          there are multiple elements, the lists of nodes corresponding
                          to each podAffinityTerm are intersected, i.e. all terms
                          must be satisfied.
                        type: array
                        items:
                          description: Defines a set of pods (namely those matching
                            the labelSelector relative to the given namespace(s))
                            that this pod should be co-located (affinity) or not co-located
                            (anti-affinity) with, where co-located is defined as running
                            on a node whose value of the label with key <topologyKey>
                            matches that of any node on which a pod of the set of
                            pods is running
                          type: object
                          required:
                          - topologyKey
                          properties:
                            labelSelector:
                              description: A label query over a set of resources,
                                in this case pods.
                              type: object
                              properties:
                                matchExpressions:
                                  description: matchExpressions is a list of label
                                    selector requirements. The requirements are ANDed.
                                  type: array
                                  items:
                                    description: A label selector requirement is a
                                      selector that contains values, a key, and an
                                      operator that relates the key and values.
                                    type: object
                                    required:
                                    - key
                                    - operator
                                    properties:
                                      key:
                                        description: key is the label key that the
                                          selector applies to.
                                        type: string
                                      operator:
                                        description: operator represents a key's relationship
                                          to a set of values. Valid operators are
                                          In, NotIn, Exists and DoesNotExist.
                                        type: string
                                      values:
                                        description: values is an array of string
                                          values. If the operator is In or NotIn,
                                          the values array must be non-empty. If the
                                          operator is Exists or DoesNotExist, the
                                          values array must be empty. This array is
                                          replaced during a strategic merge patch.
                                        type: array
                                        items:
                                          type: string
                                matchLabels:
                                  description: matchLabels is a map of {key,value}
                                    pairs. A single {key,value} in the matchLabels
                                    map is equivalent to an element of matchExpressions,
                                    whose key field is "key", the operator is "In",
                                    and the values array contains only "value". The
                                    requirements are ANDed.
                                  type: object
                                  additionalProperties:
                                    type: string
                            namespaces:
                              description: namespaces specifies which namespaces the
                                labelSelector applies to (matches against); null or
                                empty list means "this pod's namespace"
                              type: array
                              items:
                                type: string
                            topologyKey:
                              description: This pod should be co-located (affinity)
                                or not co-located (anti-affinity) with the pods matching
                                the labelSelector in the specified namespaces, where
                                co-located is defined as running on a node whose value
                                of the label with key topologyKey matches that of
                                any node on which any of the selected pods is running.
                                Empty topologyKey is not allowed.
                              type: string
                  podAntiAffinity:
                    description: Describes pod anti-affinity scheduling rules (e.g.
                      avoid putting this pod in the same node, zone, etc. as some
                      other pod(s)).
                    type: object
                    properties:
                      preferredDuringSchedulingIgnoredDuringExecution:
                        description: The scheduler will prefer to schedule pods to
                          nodes that satisfy the anti-affinity expressions specified
                          by this field, but it may choose a node that violates one
                          or more of the expressions. The node that is most preferred
                          is the one with the greatest sum of weights, i.e. for each
                          node that meets all of the scheduling requirements (resource
                          request, requiredDuringScheduling anti-affinity expressions,
                          etc.), compute a sum by iterating through the elements of
                          this field and adding "weight" to the sum if the node has
                          pods which matches the corresponding podAffinityTerm; the
                          node(s) with the highest sum are the most preferred.
                        type: array
                        items:
                          description: The weights of all of the matched WeightedPodAffinityTerm
                            fields are added per-node to find the most preferred node(s)
                          type: object
                          required:
                          - podAffinityTerm
                          - weight
                          properties:
                            podAffinityTerm:
                              description: Required. A pod affinity term, associated
                                with the corresponding weight.
                              type: object
                              required:
                              - topologyKey
                              properties:
                                labelSelector:
                                  description: A label query over a set of resources,
                                    in this case pods.
                                  type: object
                                  properties:
                                    matchExpressions:
                                      description: matchExpressions is a list of label
                                        selector requirements. The requirements are
                                        ANDed.
                                      type: array
                                      items:
                                        description: A label selector requirement
                                          is a selector that contains values, a key,
                                          and an operator that relates the key and
                                          values.
                                        type: object
                                        required:
                                        - key
                                        - operator
                                        properties:
                                          key:
                                            description: key is the label key that
                                              the selector applies to.
                                            type: string
                                          operator:
                                            description: operator represents a key's
                                              relationship to a set of values. Valid
                                              operators are In, NotIn, Exists and
                                              DoesNotExist.
                                            type: string
                                          values:
                                            description: values is an array of string
                                              values. If the operator is In or NotIn,
                                              the values array must be non-empty.
                                              If the operator is Exists or DoesNotExist,
                                              the values array must be empty. This
                                              array is replaced during a strategic
                                              merge patch.
                                            type: array
                                            items:
                                              type: string
                                    matchLabels:
                                      description: matchLabels is a map of {key,value}
                                        pairs. A single {key,value} in the matchLabels
                                        map is equivalent to an element of matchExpressions,
                                        whose key field is "key", the operator is
                                        "In", and the values array contains only "value".
                                        The requirements are ANDed.
                                      type: object
                                      additionalProperties:
                                        type: string
                                namespaces:
                                  description: namespaces specifies which namespaces
                                    the labelSelector applies to (matches against);
                                    null or empty list means "this pod's namespace"
                                  type: array
                                  items:
                                    type: string
                                topologyKey:
                                  description: This pod should be co-located (affinity)
                                    or not co-located (anti-affinity) with the pods
                                    matching the labelSelector in the specified namespaces,
                                    where co-located is defined as running on a node
                                    whose value of the label with key topologyKey
                                    matches that of any node on which any of the selected
                                    pods is running. Empty topologyKey is not allowed.
                                  type: string
                            weight:
                              description: weight associated with matching the corresponding
                                podAffinityTerm, in the range 1-100.
                              type: integer
                              format: int32
                      requiredDuringSchedulingIgnoredDuringExecution:
                        description: If the anti-affinity requirements specified by
                          this field are not met at scheduling time, the pod will
                          not be scheduled onto the node. If the anti-affinity requirements
                          specified by this field cease to be met at some point during
                          pod execution (e.g. due to a pod label update), the system
                          may or may not try to eventually evict the pod from its
                          node. When there are multiple elements, the lists of nodes
                          corresponding to each podAffinityTerm are intersected, i.e.
                          all terms must be satisfied.
                        type: array
                        items:
                          description: Defines a set of pods (namely those matching
                            the labelSelector relative to the given namespace(s))
                            that this pod should be co-located (affinity) or not co-located
                            (anti-affinity) with, where co-located is defined as running
                            on a node whose value of the label with key <topologyKey>
                            matches that of any node on which a pod of the set of
                            pods is running
                          type: object
                          required:
                          - topologyKey
                          properties:
                            labelSelector:
                              description: A label query over a set of resources,
                                in this case pods.
                              type: object
                              properties:
                                matchExpressions:
                                  description: matchExpressions is a list of label
                                    selector requirements. The requirements are ANDed.
                                  type: array
                                  items:
                                    description: A label selector requirement is a
                                      selector that contains values, a key, and an
                                      operator that relates the key and values.
                                    type: object
                                    required:
                                    - key
                                    - operator
                                    properties:
                                      key:
                                        description: key is the label key that the
                                          selector applies to.
                                        type: string
                                      operator:
                                        description: operator represents a key's relationship
                                          to a set of values. Valid operators are
                                          In, NotIn, Exists and DoesNotExist.
                                        type: string
                                      values:
                                        description: values is an array of string
                                          values. If the operator is In or NotIn,
                                          the values array must be non-empty. If the
                                          operator is Exists or DoesNotExist, the
                                          values array must be empty. This array is
                                          replaced during a strategic merge patch.
                                        type: array
                                        items:
                                          type: string
                                matchLabels:
                                  description: matchLabels is a map of {key,value}
                                    pairs. A single {key,value} in the matchLabels
                                    map is equivalent to an element of matchExpressions,
                                    whose key field is "key", the operator is "In",
                                    and the values array contains only "value". The
                                    requirements are ANDed.
                                  type: object
                                  additionalProperties:
                                    type: string
                            namespaces:
                              description: namespaces specifies which namespaces the
                                labelSelector applies to (matches against); null or
                                empty list means "this pod's namespace"
                              type: array
                              items:
                                type: string
                            topologyKey:
                              description: This pod should be co-located (affinity)
                                or not co-located (anti-affinity) with the pods matching
                                the labelSelector in the specified namespaces, where
                                co-located is defined as running on a node whose value
                                of the label with key topologyKey matches that of
                                any node on which any of the selected pods is running.
                                Empty topologyKey is not allowed.
                              type: string
              defaultRoute:
                description: defaultRoute indicates whether an external facing route
                  for the registry should be created using the default generated hostname.
                type: boolean
              disableRedirect:
                description: disableRedirect controls whether to route all data through
                  the Registry, rather than redirecting to the backend.
                type: boolean
              httpSecret:
                description: httpSecret is the value needed by the registry to secure
                  uploads, generated by default.
                type: string
              logLevel:
                description: "logLevel is an intent based logging for an overall component.
                  \ It does not give fine grained control, but it is a simple way
                  to manage coarse grained logging choices that operators have to
                  interpret for their operands. \n Valid values are: \"Normal\", \"Debug\",
                  \"Trace\", \"TraceAll\". Defaults to \"Normal\"."
                type: string
                default: Normal
                enum:
                - ""
                - Normal
                - Debug
                - Trace
                - TraceAll
              logging:
                description: logging is deprecated, use logLevel instead.
                type: integer
                format: int64
              managementState:
                description: managementState indicates whether and how the operator
                  should manage the component
                type: string
                pattern: ^(Managed|Unmanaged|Force|Removed)$
              nodeSelector:
                description: nodeSelector defines the node selection constraints for
                  the registry pod.
                type: object
                additionalProperties:
                  type: string
              observedConfig:
                description: observedConfig holds a sparse config that controller
                  has observed from the cluster state.  It exists in spec because
                  it is an input to the level for the operator
                type: object
                nullable: true
                x-kubernetes-preserve-unknown-fields: true
              operatorLogLevel:
                description: "operatorLogLevel is an intent based logging for the
                  operator itself.  It does not give fine grained control, but it
                  is a simple way to manage coarse grained logging choices that operators
                  have to interpret for themselves. \n Valid values are: \"Normal\",
                  \"Debug\", \"Trace\", \"TraceAll\". Defaults to \"Normal\"."
                type: string
                default: Normal
                enum:
                - ""
                - Normal
                - Debug
                - Trace
                - TraceAll
              proxy:
                description: proxy defines the proxy to be used when calling master
                  api, upstream registries, etc.
                type: object
                properties:
                  http:
                    description: http defines the proxy to be used by the image registry
                      when accessing HTTP endpoints.
                    type: string
                  https:
                    description: https defines the proxy to be used by the image registry
                      when accessing HTTPS endpoints.
                    type: string
                  noProxy:
                    description: noProxy defines a comma-separated list of host names
                      that shouldn't go through any proxy.
                    type: string
              readOnly:
                description: readOnly indicates whether the registry instance should
                  reject attempts to push new images or delete existing ones.
                type: boolean
              replicas:
                description: replicas determines the number of registry instances
                  to run.
                type: integer
                format: int32
              requests:
                description: requests controls how many parallel requests a given
                  registry instance will handle before queuing additional requests.
                type: object
                properties:
                  read:
                    description: read defines limits for image registry's reads.
                    type: object
                    properties:
                      maxInQueue:
                        description: maxInQueue sets the maximum queued api requests
                          to the registry.
                        type: integer
                      maxRunning:
                        description: maxRunning sets the maximum in flight api requests
                          to the registry.
                        type: integer
                      maxWaitInQueue:
                        description: maxWaitInQueue sets the maximum time a request
                          can wait in the queue before being rejected.
                        type: string
                        format: duration
                  write:
                    description: write defines limits for image registry's writes.
                    type: object
                    properties:
                      maxInQueue:
                        description: maxInQueue sets the maximum queued api requests
                          to the registry.
                        type: integer
                      maxRunning:
                        description: maxRunning sets the maximum in flight api requests
                          to the registry.
                        type: integer
                      maxWaitInQueue:
                        description: maxWaitInQueue sets the maximum time a request
                          can wait in the queue before being rejected.
                        type: string
                        format: duration
              resources:
                description: resources defines the resource requests+limits for the
                  registry pod.
                type: object
                properties:
                  limits:
                    description: 'Limits describes the maximum amount of compute resources
                      allowed. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                    type: object
                    additionalProperties:
                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                      anyOf:
                      - type: integer
                      - type: string
                      x-kubernetes-int-or-string: true
                  requests:
                    description: 'Requests describes the minimum amount of compute
                      resources required. If Requests is omitted for a container,
                      it defaults to Limits if that is explicitly specified, otherwise
                      to an implementation-defined value. More info: https://kubernetes.io/docs/concepts/configuration/manage-compute-resources-container/'
                    type: object
                    additionalProperties:
                      pattern: ^(\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))(([KMGTPE]i)|[numkMGTPE]|([eE](\+|-)?(([0-9]+(\.[0-9]*)?)|(\.[0-9]+))))?$
                      anyOf:
                      - type: integer
                      - type: string
                      x-kubernetes-int-or-string: true
              rolloutStrategy:
                description: rolloutStrategy defines rollout strategy for the image
                  registry deployment.
                type: string
                pattern: ^(RollingUpdate|Recreate)$
              routes:
                description: routes defines additional external facing routes which
                  should be created for the registry.
                type: array
                items:
                  description: ImageRegistryConfigRoute holds information on external
                    route access to image registry.
                  type: object
                  required:
                  - name
                  properties:
                    hostname:
                      description: hostname for the route.
                      type: string
                    name:
                      description: name of the route to be created.
                      type: string
                    secretName:
                      description: secretName points to secret containing the certificates
                        to be used by the route.
                      type: string
              storage:
                description: storage details for configuring registry storage, e.g.
                  S3 bucket coordinates.
                type: object
                properties:
                  azure:
                    description: azure represents configuration that uses Azure Blob
                      Storage.
                    type: object
                    properties:
                      accountName:
                        description: accountName defines the account to be used by
                          the registry.
                        type: string
                      cloudName:
                        description: cloudName is the name of the Azure cloud environment
                          to be used by the registry. If empty, the operator will
                          set it based on the infrastructure object.
                        type: string
                      container:
                        description: container defines Azure's container to be used
                          by registry.
                        type: string
                        maxLength: 63
                        minLength: 3
                        pattern: ^[0-9a-z]+(-[0-9a-z]+)*$
                  emptyDir:
                    description: 'emptyDir represents ephemeral storage on the pod''s
                      host node. WARNING: this storage cannot be used with more than
                      1 replica and is not suitable for production use. When the pod
                      is removed from a node for any reason, the data in the emptyDir
                      is deleted forever.'
                    type: object
                  gcs:
                    description: gcs represents configuration that uses Google Cloud
                      Storage.
                    type: object
                    properties:
                      bucket:
                        description: bucket is the bucket name in which you want to
                          store the registry's data. Optional, will be generated if
                          not provided.
                        type: string
                      keyID:
                        description: keyID is the KMS key ID to use for encryption.
                          Optional, buckets are encrypted by default on GCP. This
                          allows for the use of a custom encryption key.
                        type: string
                      projectID:
                        description: projectID is the Project ID of the GCP project
                          that this bucket should be associated with.
                        type: string
                      region:
                        description: region is the GCS location in which your bucket
                          exists. Optional, will be set based on the installed GCS
                          Region.
                        type: string
                  managementState:
                    description: managementState indicates if the operator manages
                      the underlying storage unit. If Managed the operator will remove
                      the storage when this operator gets Removed.
                    type: string
                    pattern: ^(Managed|Unmanaged)$
                  pvc:
                    description: pvc represents configuration that uses a PersistentVolumeClaim.
                    type: object
                    properties:
                      claim:
                        description: claim defines the Persisent Volume Claim's name
                          to be used.
                        type: string
                  s3:
                    description: s3 represents configuration that uses Amazon Simple
                      Storage Service.
                    type: object
                    properties:
                      bucket:
                        description: bucket is the bucket name in which you want to
                          store the registry's data. Optional, will be generated if
                          not provided.
                        type: string
                      cloudFront:
                        description: cloudFront configures Amazon Cloudfront as the
                          storage middleware in a registry.
                        type: object
                        required:
                        - baseURL
                        - keypairID
                        - privateKey
                        properties:
                          baseURL:
                            description: baseURL contains the SCHEME://HOST[/PATH]
                              at which Cloudfront is served.
                            type: string
                          duration:
                            description: duration is the duration of the Cloudfront
                              session.
                            type: string
                            format: duration
                          keypairID:
                            description: keypairID is key pair ID provided by AWS.
                            type: string
                          privateKey:
                            description: privateKey points to secret containing the
                              private key, provided by AWS.
                            type: object
                            required:
                            - key
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
                                description: Specify whether the Secret or its key
                                  must be defined
                                type: boolean
                      encrypt:
                        description: encrypt specifies whether the registry stores
                          the image in encrypted format or not. Optional, defaults
                          to false.
                        type: boolean
                      keyID:
                        description: keyID is the KMS key ID to use for encryption.
                          Optional, Encrypt must be true, or this parameter is ignored.
                        type: string
                      region:
                        description: region is the AWS region in which your bucket
                          exists. Optional, will be set based on the installed AWS
                          Region.
                        type: string
                      regionEndpoint:
                        description: regionEndpoint is the endpoint for S3 compatible
                          storage services. Optional, defaults based on the Region
                          that is provided.
                        type: string
                      virtualHostedStyle:
                        description: virtualHostedStyle enables using S3 virtual hosted
                          style bucket paths with a custom RegionEndpoint Optional,
                          defaults to false.
                        type: boolean
                  swift:
                    description: swift represents configuration that uses OpenStack
                      Object Storage.
                    type: object
                    properties:
                      authURL:
                        description: authURL defines the URL for obtaining an authentication
                          token.
                        type: string
                      authVersion:
                        description: authVersion specifies the OpenStack Auth's version.
                        type: string
                      container:
                        description: container defines the name of Swift container
                          where to store the registry's data.
                        type: string
                      domain:
                        description: domain specifies Openstack's domain name for
                          Identity v3 API.
                        type: string
                      domainID:
                        description: domainID specifies Openstack's domain id for
                          Identity v3 API.
                        type: string
                      regionName:
                        description: regionName defines Openstack's region in which
                          container exists.
                        type: string
                      tenant:
                        description: tenant defines Openstack tenant name to be used
                          by registry.
                        type: string
                      tenantID:
                        description: tenant defines Openstack tenant id to be used
                          by registry.
                        type: string
              tolerations:
                description: tolerations defines the tolerations for the registry
                  pod.
                type: array
                items:
                  description: The pod this Toleration is attached to tolerates any
                    taint that matches the triple <key,value,effect> using the matching
                    operator <operator>.
                  type: object
                  properties:
                    effect:
                      description: Effect indicates the taint effect to match. Empty
                        means match all taint effects. When specified, allowed values
                        are NoSchedule, PreferNoSchedule and NoExecute.
                      type: string
                    key:
                      description: Key is the taint key that the toleration applies
                        to. Empty means match all taint keys. If the key is empty,
                        operator must be Exists; this combination means to match all
                        values and all keys.
                      type: string
                    operator:
                      description: Operator represents a key's relationship to the
                        value. Valid operators are Exists and Equal. Defaults to Equal.
                        Exists is equivalent to wildcard for value, so that a pod
                        can tolerate all taints of a particular category.
                      type: string
                    tolerationSeconds:
                      description: TolerationSeconds represents the period of time
                        the toleration (which must be of effect NoExecute, otherwise
                        this field is ignored) tolerates the taint. By default, it
                        is not set, which means tolerate the taint forever (do not
                        evict). Zero and negative values will be treated as 0 (evict
                        immediately) by the system.
                      type: integer
                      format: int64
                    value:
                      description: Value is the taint value the toleration matches
                        to. If the operator is Exists, the value should be empty,
                        otherwise just a regular string.
                      type: string
              unsupportedConfigOverrides:
                description: 'unsupportedConfigOverrides holds a sparse config that
                  will override any previously set options.  It only needs to be the
                  fields to override it will end up overlaying in the following order:
                  1. hardcoded defaults 2. observedConfig 3. unsupportedConfigOverrides'
                type: object
                nullable: true
                x-kubernetes-preserve-unknown-fields: true
          status:
            description: ImageRegistryStatus reports image registry operational status.
            type: object
            required:
            - storage
            - storageManaged
            properties:
              conditions:
                description: conditions is a list of conditions and their status
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
              generations:
                description: generations are used to determine when an item needs
                  to be reconciled or has changed in a way that needs a reaction.
                type: array
                items:
                  description: GenerationStatus keeps track of the generation for
                    a given resource so that decisions about forced updates can be
                    made.
                  type: object
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
                      type: integer
                      format: int64
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
              observedGeneration:
                description: observedGeneration is the last generation change you've
                  dealt with
                type: integer
                format: int64
              readyReplicas:
                description: readyReplicas indicates how many replicas are ready and
                  at the desired state
                type: integer
                format: int32
              storage:
                description: storage indicates the current applied storage configuration
                  of the registry.
                type: object
                properties:
                  azure:
                    description: azure represents configuration that uses Azure Blob
                      Storage.
                    type: object
                    properties:
                      accountName:
                        description: accountName defines the account to be used by
                          the registry.
                        type: string
                      cloudName:
                        description: cloudName is the name of the Azure cloud environment
                          to be used by the registry. If empty, the operator will
                          set it based on the infrastructure object.
                        type: string
                      container:
                        description: container defines Azure's container to be used
                          by registry.
                        type: string
                        maxLength: 63
                        minLength: 3
                        pattern: ^[0-9a-z]+(-[0-9a-z]+)*$
                  emptyDir:
                    description: 'emptyDir represents ephemeral storage on the pod''s
                      host node. WARNING: this storage cannot be used with more than
                      1 replica and is not suitable for production use. When the pod
                      is removed from a node for any reason, the data in the emptyDir
                      is deleted forever.'
                    type: object
                  gcs:
                    description: gcs represents configuration that uses Google Cloud
                      Storage.
                    type: object
                    properties:
                      bucket:
                        description: bucket is the bucket name in which you want to
                          store the registry's data. Optional, will be generated if
                          not provided.
                        type: string
                      keyID:
                        description: keyID is the KMS key ID to use for encryption.
                          Optional, buckets are encrypted by default on GCP. This
                          allows for the use of a custom encryption key.
                        type: string
                      projectID:
                        description: projectID is the Project ID of the GCP project
                          that this bucket should be associated with.
                        type: string
                      region:
                        description: region is the GCS location in which your bucket
                          exists. Optional, will be set based on the installed GCS
                          Region.
                        type: string
                  managementState:
                    description: managementState indicates if the operator manages
                      the underlying storage unit. If Managed the operator will remove
                      the storage when this operator gets Removed.
                    type: string
                    pattern: ^(Managed|Unmanaged)$
                  pvc:
                    description: pvc represents configuration that uses a PersistentVolumeClaim.
                    type: object
                    properties:
                      claim:
                        description: claim defines the Persisent Volume Claim's name
                          to be used.
                        type: string
                  s3:
                    description: s3 represents configuration that uses Amazon Simple
                      Storage Service.
                    type: object
                    properties:
                      bucket:
                        description: bucket is the bucket name in which you want to
                          store the registry's data. Optional, will be generated if
                          not provided.
                        type: string
                      cloudFront:
                        description: cloudFront configures Amazon Cloudfront as the
                          storage middleware in a registry.
                        type: object
                        required:
                        - baseURL
                        - keypairID
                        - privateKey
                        properties:
                          baseURL:
                            description: baseURL contains the SCHEME://HOST[/PATH]
                              at which Cloudfront is served.
                            type: string
                          duration:
                            description: duration is the duration of the Cloudfront
                              session.
                            type: string
                            format: duration
                          keypairID:
                            description: keypairID is key pair ID provided by AWS.
                            type: string
                          privateKey:
                            description: privateKey points to secret containing the
                              private key, provided by AWS.
                            type: object
                            required:
                            - key
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
                                description: Specify whether the Secret or its key
                                  must be defined
                                type: boolean
                      encrypt:
                        description: encrypt specifies whether the registry stores
                          the image in encrypted format or not. Optional, defaults
                          to false.
                        type: boolean
                      keyID:
                        description: keyID is the KMS key ID to use for encryption.
                          Optional, Encrypt must be true, or this parameter is ignored.
                        type: string
                      region:
                        description: region is the AWS region in which your bucket
                          exists. Optional, will be set based on the installed AWS
                          Region.
                        type: string
                      regionEndpoint:
                        description: regionEndpoint is the endpoint for S3 compatible
                          storage services. Optional, defaults based on the Region
                          that is provided.
                        type: string
                      virtualHostedStyle:
                        description: virtualHostedStyle enables using S3 virtual hosted
                          style bucket paths with a custom RegionEndpoint Optional,
                          defaults to false.
                        type: boolean
                  swift:
                    description: swift represents configuration that uses OpenStack
                      Object Storage.
                    type: object
                    properties:
                      authURL:
                        description: authURL defines the URL for obtaining an authentication
                          token.
                        type: string
                      authVersion:
                        description: authVersion specifies the OpenStack Auth's version.
                        type: string
                      container:
                        description: container defines the name of Swift container
                          where to store the registry's data.
                        type: string
                      domain:
                        description: domain specifies Openstack's domain name for
                          Identity v3 API.
                        type: string
                      domainID:
                        description: domainID specifies Openstack's domain id for
                          Identity v3 API.
                        type: string
                      regionName:
                        description: regionName defines Openstack's region in which
                          container exists.
                        type: string
                      tenant:
                        description: tenant defines Openstack tenant name to be used
                          by registry.
                        type: string
                      tenantID:
                        description: tenant defines Openstack tenant id to be used
                          by registry.
                        type: string
              storageManaged:
                description: storageManaged is deprecated, please refer to Storage.managementState
                type: boolean
              version:
                description: version is the level this availability applies to
                type: string
  names:
    kind: Config
    listKind: ConfigList
    plural: configs
    singular: config
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
