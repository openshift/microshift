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
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/openshift/microshift/pkg/config"

	utilerrors "k8s.io/apimachinery/pkg/util/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/component-base/cli/globalflag"
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
	caCertFile := filepath.Join(cfg.DataDir, "certs", "ca-bundle", "ca-bundle.crt")
	// certDir := filepath.Join(cfg.DataDir, "certs", s.Name())
	// dataDir := filepath.Join(cfg.DataDir, s.Name())

	if err := s.configureAuditPolicy(cfg); err != nil {
		klog.Fatalf("Failed to configure kube-apiserver audit policy %v", err)
	}
	if err := s.configureOAuth(cfg); err != nil {
		klog.Fatalf("Failed to configure kube-apiserver OAuth %v", err)
	}

	// configure the kube-apiserver instance
	// TODO: configure serverOptions directly rather than via cobra
	s.serverOptions = options.NewServerRunOptions()

	// Get the apiserver port so we can set it as an argument
	apiServerPort, err := cfg.Cluster.ApiServerPort()
	if err != nil {
		return
	}

	args := []string{
		//"--openshift-config=" + cfg.DataDir + "/resources/kube-apiserver/config/config.yaml", //TOOD
		//"--advertise-address=" + ip,
		"--allow-privileged=true",
		"--anonymous-auth=false",
		"--audit-policy-file=" + cfg.DataDir + "/resources/kube-apiserver-audit-policies/default.yaml",
		"--api-audiences=https://kubernetes.svc",
		"--authorization-mode=Node,RBAC",
		"--bind-address=0.0.0.0",
		"--secure-port=" + apiServerPort,
		"--client-ca-file=" + caCertFile,
		"--enable-admission-plugins=NodeRestriction",
		"--enable-aggregator-routing=true",
		"--etcd-cafile=" + caCertFile,
		"--etcd-certfile=" + cfg.DataDir + "/resources/kube-apiserver/secrets/etcd-client/tls.crt",
		"--etcd-keyfile=" + cfg.DataDir + "/resources/kube-apiserver/secrets/etcd-client/tls.key",
		"--etcd-servers=https://127.0.0.1:2379",
		"--kubelet-certificate-authority=" + caCertFile,
		"--kubelet-client-certificate=" + cfg.DataDir + "/resources/kube-apiserver/secrets/kubelet-client/tls.crt",
		"--kubelet-client-key=" + cfg.DataDir + "/resources/kube-apiserver/secrets/kubelet-client/tls.key",
		"--profiling=false",
		"--proxy-client-cert-file=" + cfg.DataDir + "/certs/kube-apiserver/secrets/aggregator-client/tls.crt",
		"--proxy-client-key-file=" + cfg.DataDir + "/certs/kube-apiserver/secrets/aggregator-client/tls.key",
		"--requestheader-allowed-names=aggregator,system:aggregator,openshift-apiserver,system:openshift-apiserver,kube-apiserver-proxy,system:kube-apiserver-proxy,openshift-aggregator,system:openshift-aggregator",
		"--requestheader-client-ca-file=" + caCertFile,
		"--requestheader-extra-headers-prefix=X-Remote-Extra-",
		"--requestheader-group-headers=X-Remote-Group",
		"--requestheader-username-headers=X-Remote-User",
		"--service-account-issuer=https://kubernetes.svc",
		"--service-account-key-file=" + cfg.DataDir + "/resources/kube-apiserver/secrets/service-account-key/service-account.crt",
		"--service-account-signing-key-file=" + cfg.DataDir + "/resources/kube-apiserver/secrets/service-account-key/service-account.key",
		"--service-cluster-ip-range=" + cfg.Cluster.ServiceCIDR,
		"--storage-backend=etcd3",
		"--tls-cert-file=" + cfg.DataDir + "/certs/kube-apiserver/secrets/service-network-serving-certkey/tls.crt",
		"--tls-private-key-file=" + cfg.DataDir + "/certs/kube-apiserver/secrets/service-network-serving-certkey/tls.key",
		"--cors-allowed-origins=/127.0.0.1(:[0-9]+)?$,/localhost(:[0-9]+)?$",
	}
	if cfg.AuditLogDir != "" {
		args = append(args,
			"--audit-log-path="+filepath.Join(cfg.AuditLogDir, "kube-apiserver-audit.log"))
		args = append(args, "--audit-log-maxage=7")
	}

	// fake the kube-apiserver cobra command to parse args into serverOptions
	cmd := &cobra.Command{
		Use:          "kube-apiserver",
		Long:         `kube-apiserver`,
		SilenceUsage: true,
		RunE:         func(cmd *cobra.Command, args []string) error { return nil },
	}
	namedFlagSets := s.serverOptions.Flags()
	globalflag.AddGlobalFlags(namedFlagSets.FlagSet("global"), cmd.Name())
	options.AddCustomGlobalFlags(namedFlagSets.FlagSet("generic"))
	for _, f := range namedFlagSets.FlagSets {
		cmd.Flags().AddFlagSet(f)
	}
	if err := cmd.ParseFlags(args); err != nil {
		klog.Fatalf("%s failed to parse flags", s.Name(), err)
	}

	s.kubeconfig = filepath.Join(cfg.DataDir, "resources", "kubeadmin", "kubeconfig")
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
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
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
	os.MkdirAll(filepath.Dir(path), os.FileMode(0755))
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
