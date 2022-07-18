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
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/openshift/microshift/pkg/config"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/runtime"
	serveropts "k8s.io/apiserver/pkg/server/options"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	genericcontrollermanager "k8s.io/controller-manager/app"
	"k8s.io/klog/v2"
	kubeapiserver "k8s.io/kubernetes/cmd/kube-apiserver/app"
	"k8s.io/kubernetes/cmd/kube-apiserver/app/options"
)

const (
	kubeAPIStartupTimeout = 60
)

type KubeAPIServer struct {
	serverOptions *options.ServerRunOptions

	kubeconfig string
}

func NewKubeAPIServer(cfg *config.MicroshiftConfig) *KubeAPIServer {
	s := &KubeAPIServer{}
	s.configure(cfg)
	return s
}

func (s *KubeAPIServer) Name() string           { return "kube-apiserver" }
func (s *KubeAPIServer) Dependencies() []string { return []string{"etcd"} }

func (s *KubeAPIServer) configure(cfg *config.MicroshiftConfig) {
	runtime.Must(utilfeature.DefaultMutableFeatureGate.Set("APIServerTracing=true"))
	caCertFile := filepath.Join(cfg.DataDir, "certs", "ca-bundle", "ca-bundle.crt")
	// certDir := filepath.Join(cfg.DataDir, "certs", s.Name())
	// dataDir := filepath.Join(cfg.DataDir, s.Name())

	if err := s.configureAuditPolicy(cfg); err != nil {
		klog.Fatalf("Failed to configure kube-apiserver audit policy %v", err)
	}
	if err := s.configureTracingConfig(cfg); err != nil {
		klog.Fatalf("Failed to configure kube-apiserver tracing configuration %v", err)
	}
	if err := s.configureOAuth(cfg); err != nil {
		klog.Fatalf("Failed to configure kube-apiserver OAuth %v", err)
	}

	// Get the apiserver port so we can set it as an argument
	apiServerPort, err := cfg.Cluster.ApiServerPort()
	if err != nil {
		// FIXME: This function needs to deal with errors
		return
	}

	// configure the kube-apiserver instance
	s.serverOptions = options.NewServerRunOptions()

	s.serverOptions.AllowPrivileged = true
	s.serverOptions.EnableAggregatorRouting = true
	s.serverOptions.Features.EnableProfiling = false
	s.serverOptions.Traces = serveropts.NewTracingOptions()
	s.serverOptions.Traces.ConfigFile = cfg.DataDir + "/resources/kube-apiserver/tracing-config.yaml"
	s.serverOptions.ServiceAccountSigningKeyFile = cfg.DataDir + "/resources/kube-apiserver/secrets/service-account-key/service-account.key"
	s.serverOptions.ServiceClusterIPRanges = cfg.Cluster.ServiceCIDR
	s.serverOptions.ServiceNodePortRange = utilnet.PortRange{}
	// let the type deal with parsing the config file content
	err = s.serverOptions.ServiceNodePortRange.Set(cfg.Cluster.ServiceNodePortRange)
	if err != nil {
		return
	}

	s.serverOptions.ProxyClientCertFile = cfg.DataDir + "/certs/kube-apiserver/secrets/aggregator-client/tls.crt"
	s.serverOptions.ProxyClientKeyFile = cfg.DataDir + "/certs/kube-apiserver/secrets/aggregator-client/tls.key"

	s.serverOptions.Admission.GenericAdmission.EnablePlugins = []string{"NodeRestriction"}

	s.serverOptions.Audit.PolicyFile = cfg.DataDir + "/resources/kube-apiserver-audit-policies/default.yaml"

	s.serverOptions.Authentication.APIAudiences = []string{"https://kubernetes.svc"}
	s.serverOptions.Authentication.Anonymous.Allow = false
	s.serverOptions.Authentication.ClientCert.ClientCA = caCertFile

	s.serverOptions.Authentication.RequestHeader.AllowedNames = []string{
		"aggregator",
		"system:aggregator",
		"kube-apiserver-proxy",
		"system:kube-apiserver-proxy",
		"openshift-aggregator",
		"system:openshift-aggregator",
	}
	s.serverOptions.Authentication.RequestHeader.ClientCAFile = caCertFile
	s.serverOptions.Authentication.RequestHeader.ExtraHeaderPrefixes = []string{"X-Remote-Extra-"}
	s.serverOptions.Authentication.RequestHeader.GroupHeaders = []string{"X-Remote-Group"}
	s.serverOptions.Authentication.RequestHeader.UsernameHeaders = []string{"X-Remote-User"}

	s.serverOptions.Authentication.ServiceAccounts.Issuers = []string{"https://kubernetes.svc"}
	s.serverOptions.Authentication.ServiceAccounts.KeyFiles = []string{
		cfg.DataDir + "/resources/kube-apiserver/secrets/service-account-key/service-account.crt",
	}

	s.serverOptions.Authorization.Modes = []string{"Node", "RBAC"}

	s.serverOptions.Etcd.StorageConfig.Transport.ServerList = []string{"https://127.0.0.1:2379"}
	s.serverOptions.Etcd.StorageConfig.Transport.CertFile = cfg.DataDir + "/resources/kube-apiserver/secrets/etcd-client/tls.crt"
	s.serverOptions.Etcd.StorageConfig.Transport.TrustedCAFile = caCertFile
	s.serverOptions.Etcd.StorageConfig.Transport.KeyFile = cfg.DataDir + "/resources/kube-apiserver/secrets/etcd-client/tls.key"
	s.serverOptions.Etcd.StorageConfig.Transport.ServerList = []string{"https://127.0.0.1:2379"}
	s.serverOptions.Etcd.StorageConfig.Type = "etcd3"

	s.serverOptions.GenericServerRunOptions.CorsAllowedOriginList = []string{
		"/127.0.0.1(:[0-9]+)?$,/localhost(:[0-9]+)?$",
	}

	s.serverOptions.KubeletConfig.CertFile = cfg.DataDir + "/resources/kube-apiserver/secrets/kubelet-client/tls.crt"
	s.serverOptions.KubeletConfig.CAFile = caCertFile
	s.serverOptions.KubeletConfig.KeyFile = cfg.DataDir + "/resources/kube-apiserver/secrets/kubelet-client/tls.key"

	s.serverOptions.SecureServing.BindAddress = net.IP{0, 0, 0, 0}
	s.serverOptions.SecureServing.BindPort = apiServerPort
	s.serverOptions.SecureServing.ServerCert.CertKey.CertFile = cfg.DataDir + "/certs/kube-apiserver/secrets/service-network-serving-certkey/tls.crt"
	s.serverOptions.SecureServing.ServerCert.CertKey.KeyFile = cfg.DataDir + "/certs/kube-apiserver/secrets/service-network-serving-certkey/tls.key"

	s.kubeconfig = filepath.Join(cfg.DataDir, "resources", "kubeadmin", "kubeconfig")
}

func (s *KubeAPIServer) configureTracingConfig(cfg *config.MicroshiftConfig) error {
	data := []byte(`
apiVersion: apiserver.config.k8s.io/v1alpha1
kind: TracingConfiguration
# 99% sampling rate
samplingRatePerMillion: 999999`)

	path := filepath.Join(cfg.DataDir, "resources", "kube-apiserver", "tracing-config.yaml")
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
	return ioutil.WriteFile(path, data, 0644)
}

func (s *KubeAPIServer) configureAuditPolicy(cfg *config.MicroshiftConfig) error {
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

	path := filepath.Join(cfg.DataDir, "resources", "kube-apiserver-audit-policies", "default.yaml")
	os.MkdirAll(filepath.Dir(path), os.FileMode(0700))
	return ioutil.WriteFile(path, data, 0644)
}

func (s *KubeAPIServer) configureOAuth(cfg *config.MicroshiftConfig) error {
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

	path := filepath.Join(cfg.DataDir, "resources", "kube-apiserver", "oauthMetadata")
	os.MkdirAll(filepath.Dir(path), os.FileMode(0700))
	return ioutil.WriteFile(path, data, 0644)
}

func (s *KubeAPIServer) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	errorChannel := make(chan error, 1)

	// run readiness check
	go func() {
		restConfig, err := clientcmd.BuildConfigFromFlags("", s.kubeconfig)
		if err != nil {
			klog.Errorf("%s readiness check: %v", s.Name(), err)
			errorChannel <- err
		}

		versionedClient, err := kubernetes.NewForConfig(restConfig)
		if err != nil {
			klog.Errorf("%s readiness check: %v", s.Name(), err)
			errorChannel <- err
		}

		if genericcontrollermanager.WaitForAPIServer(versionedClient, kubeAPIStartupTimeout*time.Second) != nil {
			klog.Errorf("%s readiness check timed out: %v", s.Name(), err)
			errorChannel <- err
		}

		klog.Infof("%s is ready", s.Name())
		close(ready)
	}()

	// Work around that that the NewAPIServerCommand hardcodes the stop channel to SIGTERM signals,
	// so we cannot use the cobra RunE command directly.
	completedOptions, err := kubeapiserver.Complete(s.serverOptions)
	if err != nil {
		return fmt.Errorf("%s configuration error: %v", s.Name(), err)
	}
	if errs := completedOptions.Validate(); len(errs) != 0 {
		return fmt.Errorf("%s configuration error: %v", s.Name(), utilerrors.NewAggregate(errs))
	}

	go func() {
		errorChannel <- kubeapiserver.Run(completedOptions, ctx.Done())
	}()

	return <-errorChannel
}
