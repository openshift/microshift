package config

import (
	"os"
	"path/filepath"
	"strconv"
	"text/template"
)

// OpenShiftAPIServerConfig creates a config for openshift-apiserver to use
func OpenShiftAPIServerConfig(path string) error {
	configTemplate := template.Must(template.New("apiserver-config.yaml").Parse(`apiVersion: openshiftcontrolplane.config.openshift.io/v1
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
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	output, err := os.Create(path)
	if err != nil {
		return err
	}
	defer output.Close()

	return configTemplate.Execute(output, &data)
}

func OpenShiftControllerManagerConfig(path string) error {
	configTemplate := template.Must(template.New("controller-manager-config.yaml").Parse(`
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
kubeClientConfig:
  kubeConfig: {{.KubeConfig}}  
ingress:
	ingressIPNetworkCIDR: ''
	`))

	data := struct {
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
