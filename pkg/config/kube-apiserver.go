/*
Copyright © 2021 Microshift Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package config

import (
	"io/ioutil"
	"os"
	"path/filepath"
	//	"text/template"
)

func kubeAPIAuditPolicyFile(path string) error {
	data := []byte(`
apiVersion: audit.k8s.io/v1
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

func kubeAPIOAuthMetadataFile(path string) error {
	data := []byte(`
  {
    "issuer": "https://oauth-openshift.cluster.local",
    "authorization_endpoint": "https://oauth-openshift.cluster.local/oauth/authorize",
    "token_endpoint": "https://oauth-openshift.cluster.local/oauth/token",
    "scopes_supported": [
      "user:check-access",
      "user:full",
      "user:info",
      "user:list-projects",
      "user:list-scoped-projects"
    ],
    "response_types_supported": [
      "code",
      "token"
    ],
    "grant_types_supported": [
      "authorization_code",
      "implicit"
    ],
    "code_challenge_methods_supported": [
      "plain",
      "S256"
    ]
  }
`)
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	return ioutil.WriteFile(path, data, 0644)
}

// KubeAPIServerConfig creates a config for kube-apiserver to use in --openshift-config option
func KubeAPIServerConfig(cfg *MicroshiftConfig) error {
	/*
			// based on https://github.com/openshift/cluster-kube-apiserver-operator/blob/master/bindata/v4.1.0/config/defaultconfig.yaml
			configTemplate := template.Must(template.New("config").Parse(`
		apiVersion: kubecontrolplane.config.openshift.io/v1
		kind: KubeAPIServerConfig
		serviceAccountPublicKeyFiles:
		  - ` + cfg.DataDir + `/resources/kube-apiserver/sa-public-key/serving-ca.pub
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
		  authorization-mode:
		    - Scope
		    - SystemMasters
		    - RBAC
		    - Node
		  audit-log-path:
		    - ` + cfg.LogDir + `/kube-apiserver/audit.log
		  audit-policy-file:
		    - ` + cfg.DataDir + `/resources/kube-apiserver-audit-policies/default.yaml
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
		    - ` + cfg.DataDir + `/certs/ca-bundle/ca-bundle.crt
		  kubelet-client-certificate:
		    - ` + cfg.DataDir + `/resources/kube-apiserver/secrets/kubelet-client/tls.crt
		  kubelet-client-key:
		    - ` + cfg.DataDir + `/resources/kube-apiserver/secrets/kubelet-client/tls.key
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
		    - ` + cfg.DataDir + `/certs/kube-apiserver/secrets/aggregator-client/tls.crt
		  proxy-client-key-file:
		    - ` + cfg.DataDir + `/certs/kube-apiserver/secrets/aggregator-client/tls.key
		  requestheader-allowed-names:
		    - system:admin
		    - aggregator
		    - system:aggregator
		    - openshift-apiserver
		    - system:openshift-apiserver
		    - kube-apiserver
		    - system:kube-apiserver
		    - system:openshift-aggregator
		    - kube-apiserver-proxy
		    - system:kube-apiserver-proxy
		    - system:openshift-aggregator
		    - openshift-aggregator
		  requestheader-client-ca-file:
		    - ` + cfg.DataDir + `/certs/ca-bundle/ca-bundle.crt
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
		    - ` + cfg.DataDir + `/certs/kube-apiserver/secrets/service-network-serving-certkey/tls.crt
		  tls-private-key-file:
		    - ` + cfg.DataDir + `/certs/kube-apiserver/secrets/service-network-serving-certkey/tls.key
		  service-account-issuer:
		    - "https://kubernetes.svc"
		  service-account-signing-key-file:
		    - ` + cfg.DataDir + `/resources/kube-apiserver/secrets/service-account-signing-key/service-account.key
		  etcd-cafile:
		    - ` + cfg.DataDir + `/certs/ca-bundle/ca-bundle.crt
		  etcd-certfile:
		    - ` + cfg.DataDir + `/resources/kube-apiserver/secrets/etcd-client/tls.crt
		  etcd-keyfile:
		    - ` + cfg.DataDir + `/resources/kube-apiserver/secrets/etcd-client/tls.key
		  etcd-prefix:
		    - kubernetes.io
		  etcd-servers:
		    - https://127.0.0.1:2379
		auditConfig:
		  auditFilePath: "` + cfg.LogDir + `/kube-apiserver/audit.log"
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
		      - "/readyz"
		      userGroups:
		      - system:authenticated
		      - system:unauthenticated
		    - level: Metadata
		      omitStages:
		      - RequestReceived
		authConfig:
		  oauthMetadataFile: "` + cfg.DataDir + `/resources/kube-apiserver/oauthMetadata"
		  requestHeader:
		    clientCA: "` + cfg.DataDir + `/certs/ca-bundle/ca-bundle.crt"
		    clientCommonNames:
		    - kube-apiserver
		    - system:kube-apiserver
		    - kube-apiserver-proxy
		    - system:kube-apiserver-proxy
		    - system:openshift-aggregator
		    extraHeaderPrefixes:
		    - X-Remote-Extra-
		    groupHeaders:
		    - X-Remote-Group
		    usernameHeaders:
		    - X-Remote-User
		  webhookTokenAuthenticators:
		consolePublicURL: ""
		projectConfig:
		  defaultNodeSelector: ""
		servicesSubnet: {{.ServiceCIDR}} # 10.3.0.0/16 # ServiceCIDR # set by observe_network.go
		servingInfo:
		  bindAddress: 0.0.0.0:6443 # set by observe_network.go
		  bindNetwork: tcp4 # set by observe_network.go
		`))

			data := struct {
				ServiceCIDR string
			}{
				ServiceCIDR: cfg.Cluster.ServiceCIDR,
			}
			os.MkdirAll(filepath.Dir(cfg.DataDir+"/resources/kube-apiserver/config/config.yaml"), os.FileMode(0755))
			output, err := os.Create(cfg.DataDir + "/resources/kube-apiserver/config/config.yaml")
			if err != nil {
				return err
			}
			defer output.Close()

			if err := configTemplate.Execute(output, &data); err != nil {
		    return err
		  }
	*/
	if err := kubeAPIOAuthMetadataFile(cfg.DataDir + "/resources/kube-apiserver/oauthMetadata"); err != nil {
		return err
	}
	if err := kubeAPIAuditPolicyFile(cfg.DataDir + "/resources/kube-apiserver-audit-policies/default.yaml"); err != nil {
		return err
	}
	return nil
}
