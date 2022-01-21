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
package controllers

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/microshift/pkg/config"
	openshift_apiserver "github.com/openshift/openshift-apiserver/pkg/cmd/openshift-apiserver"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/util/wait"
	genericapiserveroptions "k8s.io/apiserver/pkg/server/options"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type OCPAPIServer struct {
	options openshift_apiserver.OpenShiftAPIServer
	cfg     *config.MicroshiftConfig
}

func NewOpenShiftAPIServer(cfg *config.MicroshiftConfig) *OCPAPIServer {
	s := &OCPAPIServer{}
	s.configure(cfg)
	return s
}

func (s *OCPAPIServer) Name() string { return "ocp-apiserver" }
func (s *OCPAPIServer) Dependencies() []string {
	return []string{"kube-apiserver", "openshift-crd-manager"}
}

func (s *OCPAPIServer) configure(cfg *config.MicroshiftConfig) error {
	var configFilePath = cfg.DataDir + "/resources/openshift-apiserver/config/config.yaml"

	if err := OpenShiftAPIServerConfig(cfg); err != nil {
		klog.Infof("Failed to create a new ocp-apiserver configuration: %v", err)
		return err
	}
	args := []string{
		"--config=" + configFilePath,
	}

	options := openshift_apiserver.OpenShiftAPIServer{
		Output:         os.Stdout,
		Authentication: genericapiserveroptions.NewDelegatingAuthenticationOptions(),
		Authorization:  genericapiserveroptions.NewDelegatingAuthorizationOptions().WithAlwaysAllowPaths("/healthz", "/healthz/").WithAlwaysAllowGroups("system:masters"),
	}
	options.Authentication.RemoteKubeConfigFile = cfg.DataDir + "/resources/kubeadmin/kubeconfig"
	options.Authorization.RemoteKubeConfigFile = cfg.DataDir + "/resources/kubeadmin/kubeconfig"

	cmd := &cobra.Command{
		Use:          "ocp-apiserver",
		Long:         "ocp-apiserver",
		SilenceUsage: true,
		RunE:         func(cmd *cobra.Command, args []string) error { return nil },
	}

	cmd.SetArgs(args)
	cmd.MarkFlagFilename("config", "yaml", "yml")
	cmd.MarkFlagRequired("config")
	cmd.ParseFlags(args)

	s.cfg = cfg
	s.options = options
	s.options.ConfigFile = configFilePath

	klog.Infof("starting openshift-apiserver %s, args: %v", cfg.NodeIP, args)
	return nil

}

func (s *OCPAPIServer) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	go func() error {
		// probe ocp api services
		restConfig, err := clientcmd.BuildConfigFromFlags("", s.cfg.DataDir+"/resources/kubeadmin/kubeconfig")
		if err != nil {
			return err
		}
		client, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			return err
		}
		err = waitForOCPAPIServer(client, 10*time.Second)
		if err != nil {
			klog.Warningf("Failed to wait for ocp apiserver: %v", err)
			return err
		}
		klog.Infof("ocp apiserver is ready")
		close(ready)
		return nil
	}()

	err := s.prepareOCPComponents(s.cfg)
	if err != nil {
		klog.Errorf("Failed to prepare ocp-components %v", err)
		return err
	}

	stopCh := make(chan struct{})
	if err := s.options.RunAPIServer(stopCh); err != nil {
		klog.Fatalf("Failed to start ocp-apiserver %v", err)
	}

	return ctx.Err()
}

func (s *OCPAPIServer) prepareOCPComponents(cfg *config.MicroshiftConfig) error {

	// ocp api service registration
	if err := createAPIHeadlessSvc(cfg, "openshift-apiserver", 8444); err != nil {
		klog.Warningf("failed to apply headless svc %v", err)
		return err
	}
	if err := createAPIHeadlessSvc(cfg, "openshift-oauth-apiserver", 8443); err != nil {
		klog.Warningf("failed to apply headless svc %v", err)
		return err
	}
	if err := createAPIRegistration(cfg); err != nil {
		klog.Warningf("failed to register api %v", err)
		return err
	}

	return nil
}

func waitForOCPAPIServer(client kubernetes.Interface, timeout time.Duration) error {
	var lastErr error

	err := wait.PollImmediate(time.Second, timeout, func() (bool, error) {
		for _, apiSvc := range []string{
			"apps.openshift.io",
			"authorization.openshift.io",
			"build.openshift.io",
			"image.openshift.io",
			"project.openshift.io",
			"quota.openshift.io",
			"route.openshift.io",
			"security.openshift.io",
			"template.openshift.io", //TODO missing templateinstances
		} {
			status := 0
			url := "/apis/" + apiSvc
			result := client.Discovery().RESTClient().Get().AbsPath(url).Do(context.TODO()).StatusCode(&status)
			if result.Error() != nil {
				lastErr = fmt.Errorf("failed to get apiserver %s status: %v", url, result.Error())
				return false, nil
			}
			if status != http.StatusOK {
				content, _ := result.Raw()
				lastErr = fmt.Errorf("APIServer isn't available: %v", string(content))
				klog.Warningf("APIServer isn't available yet: %v. Waiting a little while.", string(content))
				return false, nil
			}
		}
		return true, nil
	})

	if err != nil {
		return fmt.Errorf("%v: %v", err, lastErr)
	}

	return nil
}

func OpenShiftAPIServerConfig(cfg *config.MicroshiftConfig) error {
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
  kubeConfig:  ` + cfg.DataDir + `/resources/kubeadmin/kubeconfig
apiServerArguments:
  minimal-shutdown-duration:
  - 30s
  anonymous-auth:
  - "false"
  authorization-kubeconfig:
  - ` + cfg.DataDir + `/resources/kubeadmin/kubeconfig
  authentication-kubeconfig:
  - ` + cfg.DataDir + `/resources/kubeadmin/kubeconfig
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
  - Node`)

	if cfg.AuditLogDir != "" {
		data = append(data, `
auditConfig:
  auditFilePath: `+cfg.AuditLogDir+`openshift-apiserver-audit.log
  enabled: true
  logFormat: json
  maximumFileSizeMegabytes: 100
  maximumRetainedFiles: 10
  policyFile: "`+cfg.DataDir+`/resources/openshift-apiserver/config/policy.yaml"
  policyConfiguration:
    apiVersion: audit.k8s.io/v1
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
      - RequestReceived`...)
	}

	data = append(data, `
imagePolicyConfig:
  internalRegistryHostname: image-registry.openshift-image-registry.svc:5000
projectConfig:
  projectRequestMessage: ''
routingConfig:
  subdomain: `+cfg.Cluster.Domain+`
servingInfo:
  bindAddress: "0.0.0.0:8444"
  certFile: `+cfg.DataDir+`/resources/openshift-apiserver/secrets/tls.crt
  keyFile: `+cfg.DataDir+`/resources/openshift-apiserver/secrets/tls.key
  ca: `+cfg.DataDir+`/certs/ca-bundle/ca-bundle.crt
storageConfig:
  urls:
  - https://127.0.0.1:2379
  certFile: `+cfg.DataDir+`/resources/kube-apiserver/secrets/etcd-client/tls.crt
  keyFile: `+cfg.DataDir+`/resources/kube-apiserver/secrets/etcd-client/tls.key
  ca: `+cfg.DataDir+`/certs/ca-bundle/ca-bundle.crt
  `...)

	path := filepath.Join(cfg.DataDir, "resources", "openshift-apiserver", "config", "config.yaml")
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	return ioutil.WriteFile(path, data, 0644)
}
