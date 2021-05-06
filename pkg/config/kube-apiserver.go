package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

const (
	port = 32444
)

func kubeAPIAuditPolicyFile(path string) error {
	data := []byte(`
apiVersion: audit.k8s.io/v1beta1
kind: Policy
metadata:
  name: Default
# Don't generate audit events for all requests in RequestReceived stage.
omitStages:
- "RequestReceived"
rules:
# Don't log requests for events
- level: None
  resources:
  - group: ""
    resources: ["events"]
# Don't log oauth tokens as metadata.name is the secret
- level: None
  resources:
  - group: "oauth.openshift.io"
    resources: ["oauthaccesstokens", "oauthauthorizetokens"]
# Don't log authenticated requests to certain non-resource URL paths.
- level: None
  userGroups: ["system:authenticated", "system:unauthenticated"]
  nonResourceURLs:
  - "/api*" # Wildcard matching.
  - "/version"
  - "/healthz"
  - "/readyz"
# A catch-all rule to log all other requests at the Metadata level.
- level: Metadata
  # Long-running requests like watches that fall under this rule will not
  # generate an audit event in RequestReceived.
  omitStages:
  - "RequestReceived"`)
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	return ioutil.WriteFile(path, data, 0644)
}

// KubeAPIServerConfig creates a config for kube-apiserver to use in --openshift-config option
func KubeAPIServerConfig(path, svcCIDR string) error {
	// based on https://github.com/openshift/cluster-kube-apiserver-operator/blob/master/bindata/v4.1.0/config/defaultconfig.yaml
	configTemplate := template.Must(template.New("config").Parse(`
apiVersion: kubecontrolplane.config.openshift.io/v1  
kind: KubeAPIServerConfig
serviceAccountPublicKeyFiles:
  - /etc/kubernetes/ushift-resources/kube-apiserver/sa-public-key/serving-ca.pub
admission:
  pluginConfig:
    network.openshift.io/ExternalIPRanger:
      configuration:
        allowIngressIP: true
        apiVersion: network.openshift.io/v1
        externalIPNetworkCIDRs: null
        kind: ExternalIPRangerAdmissionConfig
      location: ""
apiServerArguments:
  allow-privileged:
    - "true"
  anonymous-auth:
    - "true"
  audit-log-format:
    - json
  audit-log-maxbackup:
    - "10"
  audit-log-maxsize:
    - "100"
  audit-log-path:
    - /var/log/kube-apiserver/audit.log
  audit-policy-file:
    - /etc/kubernetes/ushift-resources/kube-apiserver-audit-policies/default.yaml
  enable-admission-plugins:
    - CertificateApproval
    - CertificateSigning
    - CertificateSubjectRestriction
    - DefaultIngressClass
    - DefaultStorageClass
    - DefaultTolerationSeconds
    - LimitRanger
    - MutatingAdmissionWebhook
    - NamespaceLifecycle
    - NodeRestriction
    - OwnerReferencesPermissionEnforcement
    - PersistentVolumeClaimResize
    - PersistentVolumeLabel
    - PodNodeSelector
    - PodTolerationRestriction
    - Priority
    - ResourceQuota
    - ServiceAccount
    - StorageObjectInUseProtection
    - TaintNodesByCondition
    - ValidatingAdmissionWebhook
    - authorization.openshift.io/RestrictSubjectBindings
    - authorization.openshift.io/ValidateRoleBindingRestriction
    - config.openshift.io/DenyDeleteClusterConfiguration
    - config.openshift.io/ValidateAPIServer
    - config.openshift.io/ValidateAuthentication
    - config.openshift.io/ValidateConsole
    - config.openshift.io/ValidateFeatureGate
    - config.openshift.io/ValidateImage
    - config.openshift.io/ValidateOAuth
    - config.openshift.io/ValidateProject
    - config.openshift.io/ValidateScheduler
    - image.openshift.io/ImagePolicy
    - network.openshift.io/ExternalIPRanger
    - network.openshift.io/RestrictedEndpointsAdmission
    - quota.openshift.io/ClusterResourceQuota
    - quota.openshift.io/ValidateClusterResourceQuota
    - route.openshift.io/IngressAdmission
    - scheduling.openshift.io/OriginPodNodeEnvironment
    - security.openshift.io/DefaultSecurityContextConstraints
    - security.openshift.io/SCCExecRestrictions
    - security.openshift.io/SecurityContextConstraint
    - security.openshift.io/ValidateSecurityContextConstraints
  # switch to direct pod IP routing for aggregated apiservers to avoid service IPs as on source of instability
  enable-aggregator-routing:
    - "true"
  enable-logs-handler:
    - "false"
  enable-swagger-ui:
    - "true"
  endpoint-reconciler-type:
    - "lease"
  event-ttl:
    - 3h
  goaway-chance:
    - "0"
  http2-max-streams-per-connection:
    - "2000"  # recommended is 1000, but we need to mitigate https://github.com/kubernetes/kubernetes/issues/74412
  insecure-port:
    - "0"
  kubelet-certificate-authority:
    - /etc/kubernetes/ushift-certs/ca-bundle/ca-bundle.crt
  kubelet-client-certificate:
    - /etc/kubernetes/ushift-resources/kube-apiserver/secrets/kubelet-client/tls.crt
  kubelet-client-key:
    - /etc/kubernetes/ushift-resources/kube-apiserver/secrets/kubelet-client/tls.key
  kubelet-https:
    - "true"
  kubelet-preferred-address-types:
    - InternalIP # all of our kubelets have internal IPs and we *only* support communicating with them via that internal IP so that NO_PROXY always works and is lightweight
  kubelet-read-only-port:
    - "0"
  kubernetes-service-node-port:
    - "0"
  # value should logically scale with max-requests-inflight
  max-mutating-requests-inflight:
    - "1000"
  # value needed to be bumped for scale tests.  The kube-apiserver did ok here
  max-requests-inflight:
    - "3000"
  min-request-timeout:
    - "3600"
  proxy-client-cert-file:
    - /etc/kubernetes/ushift-certs/kube-apiserver/secrets/aggregator-client/tls.crt
  proxy-client-key-file:
    - /etc/kubernetes/ushift-certs/kube-apiserver/secrets/aggregator-client/tls.key
  requestheader-allowed-names:
    - kube-apiserver-proxy
    - system:kube-apiserver-proxy
    - system:openshift-aggregator
  requestheader-client-ca-file:
    - /etc/kubernetes/ushift-certs/ca-bundle/ca-bundle.crt
  requestheader-extra-headers-prefix:
    - X-Remote-Extra-
  requestheader-group-headers:
    - X-Remote-Group
  requestheader-username-headers:
    - X-Remote-User
  # need to enable alpha APIs for the priority and fairness feature
  service-account-lookup:
    - "true"
  service-node-port-range:
    - 30000-32767
  shutdown-delay-duration:
    - 70s # give SDN some time to converge: 30s for iptable lock contention, 25s for the second try and some seconds for AWS to update ELBs
  storage-backend:
    - etcd3
  storage-media-type:
    - application/vnd.kubernetes.protobuf
  tls-cert-file:
    - /etc/kubernetes/ushift-certs/kube-apiserver/secrets/service-network-serving-certkey/tls.crt
  tls-private-key-file:
    - /etc/kubernetes/ushift-certs/kube-apiserver/secrets/service-network-serving-certkey/tls.key
  service-account-issuer:
    - https://kubernetes.default.svc
  service-account-signing-key-file:
    - /etc/kubernetes/ushift-resources/kube-apiserver/secrets/service-account-signing-key/service-account.key
  etcd-cafile:
    - /etc/kubernetes/ushift-certs/ca-bundle/ca-bundle.crt
  etcd-certfile:
    - /etc/kubernetes/ushift-resources/kube-apiserver/secrets/etcd-client/tls.crt
  etcd-keyfile:
    - /etc/kubernetes/ushift-resources/kube-apiserver/secrets/etcd-client/tls.key
  etcd-prefix:
    - kubernetes.io
  etcd-servers:
    - https://127.0.0.1:2379
authConfig:
  oauthMetadataFile: ""
consolePublicURL: ""
projectConfig:
  defaultNodeSelector: ""
servicesSubnet: {{.ServiceCIDR}} # 10.3.0.0/16 # ServiceCIDR # set by observe_network.go
servingInfo:
  bindAddress: 0.0.0.0:6443 # set by observe_network.go
  bindNetwork: tcp4 # set by observe_network.go
  namedCertificates: null # set by observe_apiserver.go`))

	if err := kubeAPIAuditPolicyFile("/etc/kubernetes/ushift-resources/kube-apiserver-audit-policies/default.yaml"); err != nil {
		return err
	}
	data := struct {
		ServiceCIDR string
	}{
		ServiceCIDR: svcCIDR,
	}
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()

	return configTemplate.Execute(output, &data)

}
