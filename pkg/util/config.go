package util

import (
	"bytes"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"text/template"
)

// KubeAPIServerConfig creates a config for kube-apiserver to use in --openshift-config option
func KubeAPIServerConfig() error {
	// based on https://github.com/openshift/cluster-kube-apiserver-operator/blob/master/bindata/v4.1.0/config/defaultconfig.yaml
	apiConfigTemplate := template.Must(template.New("config").Parse(`
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
  authorization-mode:
    - Scope
    - SystemMasters
    - RBAC
    - Node
  audit-log-format:
    - json
  audit-log-maxbackup:
    - "10"
  audit-log-maxsize:
    - "100"
  audit-log-path:
    - /var/log/kube-apiserver/audit.log
  audit-policy-file:
    - /etc/kubernetes/static-pod-resources/configmaps/kube-apiserver-audit-policies/default.yaml
  client-ca-file:
    - /etc/kubernetes/static-pod-certs/configmaps/client-ca/ca-bundle.crt
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
    - RuntimeClass
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
  etcd-cafile:
    - /etc/kubernetes/static-pod-resources/configmaps/etcd-serving-ca/ca-bundle.crt
  etcd-certfile:
    - /etc/kubernetes/static-pod-resources/secrets/etcd-client/tls.crt
  etcd-keyfile:
    - /etc/kubernetes/static-pod-resources/secrets/etcd-client/tls.key
  etcd-prefix:
    - kubernetes.io
  event-ttl:
    - 3h
  goaway-chance:
    - "0"
  http2-max-streams-per-connection:
    - "2000"  # recommended is 1000, but we need to mitigate https://github.com/kubernetes/kubernetes/issues/74412
  insecure-port:
    - "0"
  kubelet-certificate-authority:
    - /etc/kubernetes/static-pod-resources/configmaps/kubelet-serving-ca/ca-bundle.crt
  kubelet-client-certificate:
    - /etc/kubernetes/static-pod-resources/secrets/kubelet-client/tls.crt
  kubelet-client-key:
    - /etc/kubernetes/static-pod-resources/secrets/kubelet-client/tls.key
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
    - /etc/kubernetes/static-pod-certs/secrets/aggregator-client/tls.crt
  proxy-client-key-file:
    - /etc/kubernetes/static-pod-certs/secrets/aggregator-client/tls.key
  requestheader-allowed-names:
    - kube-apiserver-proxy
    - system:kube-apiserver-proxy
    - system:openshift-aggregator
  requestheader-client-ca-file:
    - /etc/kubernetes/static-pod-certs/configmaps/aggregator-client-ca/ca-bundle.crt
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
    - /etc/kubernetes/static-pod-certs/secrets/service-network-serving-certkey/tls.crt
  tls-private-key-file:
    - /etc/kubernetes/static-pod-certs/secrets/service-network-serving-certkey/tls.key
authConfig:
  oauthMetadataFile: ""
consolePublicURL: ""
projectConfig:
  defaultNodeSelector: ""
servicesSubnet: 10.3.0.0/16 # ServiceCIDR # set by observe_network.go
servingInfo:
  bindAddress: 0.0.0.0:6443 # set by observe_network.go
  bindNetwork: tcp4 # set by observe_network.go
  namedCertificates: null # set by observe_apiserver.go
	`))
	return nil
}

// KubeControllerManagerConfig creates a config for kube-controller-manager in option --openshift-config
func KubeControllerManagerConfig() error {
	cmConfigTemplate := template.Must(template.New("config").Parse(`
apiVersion: kubecontrolplane.config.openshift.io/v1
kind: KubeControllerManagerConfig  
extendedArguments:
  enable-dynamic-provisioning:
  - "true"
  allocate-node-cidrs:
  - "false"
  configure-cloud-routes:
  - "false"
  use-service-account-credentials:
  - "true"
  flex-volume-plugin-dir:
  - "/etc/kubernetes/kubelet-plugins/volume/exec" # created by machine-config-operator, owned by storage team/hekumar@redhat.com
  pv-recycler-pod-template-filepath-nfs: # owned by storage team/fbertina@redhat.com
  - "/etc/kubernetes/static-pod-resources/configmaps/recycler-config/recycler-pod.yaml"
  pv-recycler-pod-template-filepath-hostpath: # owned by storage team/fbertina@redhat.com
  - "/etc/kubernetes/static-pod-resources/configmaps/recycler-config/recycler-pod.yaml"
  leader-elect:
  - "true"
  leader-elect-retry-period:
  - "3s"
  leader-elect-resource-lock:
  - "configmaps"
  controllers:
  - "*"
  - "-ttl" # TODO: this is excluded in kube-core, but not in #21092
  - "-bootstrapsigner"
  - "-tokencleaner"
  experimental-cluster-signing-duration:
  - "720h"
  secure-port:
  - "10257"
  port:
  - "0"
  cert-dir:
  - "/var/run/kubernetes"
  root-ca-file:
  - "/etc/kubernetes/static-pod-resources/configmaps/serviceaccount-ca/ca-bundle.crt"
  service-account-private-key-file:
  - "/etc/kubernetes/static-pod-resources/secrets/service-account-private-key/service-account.key"
  cluster-signing-cert-file:
  - "/etc/kubernetes/static-pod-certs/secrets/csr-signer/tls.crt"
  cluster-signing-key-file:
  - "/etc/kubernetes/static-pod-certs/secrets/csr-signer/tls.key"
  kube-api-qps:
  - "150" # this is a historical values
  kube-api-burst:
  - "300" # this is a historical values`))
	return nil
}

// OpenShiftAPIServerConfig creates a config for openshift-apiserver to use
func OpenShiftAPIServerConfig() error {
	ocpAPIConfigTemplate = template.Must(template.New("apiserver-config.yaml").Parse(`apiVersion: openshiftcontrolplane.config.openshift.io/v1
kind: OpenShiftAPIServerConfig
aggregatorConfig:
  allowedNames:
  - kube-apiserver
  - system:kube-apiserver
  - kube-apiserver-proxy
  - system:kube-apiserver-proxy
  - system:openshift-aggregator
  - system:admin
  clientCA: {{.ClientCACert}}
  extraHeaderPrefixes:
  - X-Remote-Extra-
  groupHeaders:
  - X-Remote-Group
  usernameHeaders:
  - X-Remote-User
apiServerArguments:
  minimal-shutdown-duration:
  - 30s
auditConfig:
  auditFilePath: "/var/log/openshift-apiserver/audit.log"
  enabled: true
  logFormat: json
  maximumFileSizeMegabytes: 100
  maximumRetainedFiles: 10
  policyConfiguration:
    apiVersion: audit.k8s.io/v1beta1
    kind: Policy
    omitStages:
    - RequestReceived
    rules:
    - level: None
      resources:
      - group: ''
        resources:
        - events
    - level: None
      resources:
      - group: oauth.openshift.io
        resources:
        - oauthaccesstokens
        - oauthauthorizetokens
    - level: None
      nonResourceURLs:
      - "/api*"
      - "/version"
      - "/healthz"
      userGroups:
      - system:authenticated
      - system:unauthenticated
    - level: Metadata
      omitStages:
      - RequestReceived
kubeClientConfig:
  kubeConfig: {{.KubeConfig}}
servingInfo:
  bindAddress: "0.0.0.0:` + strconv.Itoa(port) + `" 
  certFile: {{.ServingCert}}
  keyFile: {{.ServingKey}}
  clientCA: {{.ServingClientCert}}
imagePolicyConfig:
  internalRegistryHostname: image-registry.openshift-image-registry.svc:5000
projectConfig:
  projectRequestMessage: ''
routingConfig:
  subdomain: {{.IngressDomain}}
storageConfig:
  urls:
  - {{.EtcdUrl}}
  certFile: {{.EtcdCert}}
  keyFile: {{.EtcdKey}}
  ca: {{.EtcdCA}}
  `))

	data := struct {
		ClientCACert, KubeConfig, ServingCert, ServingKey, ServingClientCert,
		IngressDomain, EtcdUrl, EtcdCert, EtcdKey, EtcdCA string
	}{
		/*
			ClientCACert:      ,
			KubeConfig:        ,
			ServingCert:       ,
			ServingKey:        ,
			ServingClientCert: ,
			IngressDomain:     ,
			EtcdUrl:           ,
			EtcdCA:            ,
			EtcdCert:          ,
			EtcdKey:           ,
		*/
	}

	output, err := os.Create()
	if err != nil {
		return err
	}
	defer output.Close()

	return ocpAPIConfigTemplate.Execute(output, &data)
}

func OpenShiftControllerManagerConfig() error {
	ocpControllManagerConfigTemplate = template.Must(template.New("controller-manager-config.yaml").Parse(`
apiVersion: openshiftcontrolplane.config.openshift.io/v1
kind: OpenShiftControllerManagerConfig
build:
	buildDefaults:
	resources: {}
	imageTemplateFormat:
	format: {{.BuilderImage}}
deployer:
	imageTemplateFormat:
	format: {{.DeployerName}}
dockerPullSecret:
	internalRegistryHostname: {{.ImageRegistryUrl}}
ingress:
	ingressIPNetworkCIDR: ''
kubeClientConfig:
	kubeConfig: {{.KubeConfig}}
servingInfo:
	bindAddress: "0.0.0.0:8445"
	certFile: {{.ServingCert}}
	keyFile: {{.ServingKey}}
	clientCA: {{.ServingClientCert}}
	`))
	data := struct {
		ClientCACert, KubeConfig, ServingCert, ServingKey, ServingClientCert,
		IngressDomain, EtcdUrl, EtcdCert, EtcdKey, EtcdCA string
	}{
		/*
			ClientCACert:      ,
			KubeConfig:        ,
			ServingCert:       ,
			ServingKey:        ,
			ServingClientCert: ,
			IngressDomain:     ,
			EtcdUrl:           ,
			EtcdCA:            ,
			EtcdCert:          ,
			EtcdKey:           ,
		*/
	}

	output, err := os.Create()
	if err != nil {
		return err
	}
	defer output.Close()

	return ocpAPIConfigTemplate.Execute(output, &data)

}
