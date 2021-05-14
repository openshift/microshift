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
	"text/template"

	"github.com/openshift/microshift/pkg/constant"
)

// OpenShiftAPIServerConfig creates a config for openshift-apiserver to use
func OpenShiftAPIServerConfig(path string) error {
	data := []byte(`apiVersion: openshiftcontrolplane.config.openshift.io/v1
kind: OpenShiftAPIServerConfig
aggregatorConfig:
  allowedNames:
  - kube-apiserver
  - system:kube-apiserver
  - kube-apiserver-proxy
  - system:kube-apiserver-proxy
  - system:openshift-aggregator
  - system:admin
  extraHeaderPrefixes:
  - X-Remote-Extra-
  groupHeaders:
  - X-Remote-Group
  usernameHeaders:
  - X-Remote-User
kubeClientConfig:
  kubeConfig:  ` + constant.AdminKubeconfigPath + `
apiServerArguments:
  minimal-shutdown-duration:
  - 30s
  anonymous-auth:
  - "false"
  authorization-kubeconfig:
  - ` + constant.AdminKubeconfigPath + `
  authentication-kubeconfig:
  - ` + constant.AdminKubeconfigPath + `
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
imagePolicyConfig:
  internalRegistryHostname: image-registry.openshift-image-registry.svc:5000
projectConfig:
  projectRequestMessage: ''
routingConfig:
  subdomain: ` + constant.DomainName + `
servingInfo:
  bindAddress: "0.0.0.0:8444"
  certFile: /etc/kubernetes/ushift-resources/ocp-apiserver/secrets/tls.crt
  keyFile: /etc/kubernetes/ushift-resources/ocp-apiserver/secrets/tls.key
  ca: /etc/kubernetes/ushift-certs/ca-bundle/ca-bundle.crt
storageConfig:
  urls:
  - https://127.0.0.1:2379
  certFile: /etc/kubernetes/ushift-resources/kube-apiserver/secrets/etcd-client/tls.crt
  keyFile: /etc/kubernetes/ushift-resources/kube-apiserver/secrets/etcd-client/tls.key
  ca: /etc/kubernetes/ushift-certs/ca-bundle/ca-bundle.crt
  `)
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	return ioutil.WriteFile(path, data, 0644)
}

func OpenShiftControllerManagerConfig(path string) error {
	configTemplate := template.Must(template.New("controller-manager-config.yaml").Parse(`
apiVersion: openshiftcontrolplane.config.openshift.io/v1
kind: OpenShiftControllerManagerConfig
kubeClientConfig:
  kubeConfig: ` + constant.AdminKubeconfigPath + `
servingInfo:
  bindAddress: "0.0.0.0:8445"
  certFile: /etc/kubernetes/ushift-resources/ocp-controller-manager/secrets/tls.crt
  keyFile:  /etc/kubernetes/ushift-resources/ocp-controller-manager/secrets/tls.key
  clientCA: /etc/kubernetes/ushift-certs/ca-bundle/ca-bundle.crt`))

	data := struct { //TODO
		KubeConfig, BuilderImage, DeployerName, ImageRegistryUrl string
	}{
		//KubeConfig: ,
		BuilderImage:     "docker-build",
		DeployerName:     "docker-build",
		ImageRegistryUrl: "image-registry.openshift-image-registry.svc:5000",
	}
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()

	return configTemplate.Execute(output, &data)
}
