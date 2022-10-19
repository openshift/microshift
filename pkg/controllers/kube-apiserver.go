/*
Copyright © 2021 MicroShift Contributors

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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	kubeapiserver "k8s.io/kubernetes/cmd/kube-apiserver/app"

	configv1 "github.com/openshift/api/config/v1"
	kubecontrolplanev1 "github.com/openshift/api/kubecontrolplane/v1"
	"github.com/openshift/library-go/pkg/crypto"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"

	embedded "github.com/openshift/microshift/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

const (
	kubeAPIStartupTimeout = 60
)

var baseKubeAPIServerConfigs = [][]byte{
	// todo: a nice way to generate the baseline kas config for microshift
	embedded.MustAsset("components/kube-apiserver/defaultconfig.yaml"),
	embedded.MustAsset("components/kube-apiserver/config-overrides.yaml"),
}

var fixedTLSProfile *configv1.TLSProfileSpec

func init() {
	var ok bool
	fixedTLSProfile, ok = configv1.TLSProfiles[configv1.TLSProfileIntermediateType]
	if !ok {
		panic("lookup of intermediate tls profile failed")
	}
}

type KubeAPIServer struct {
	kasConfigBytes []byte
	verbosity      int
	configureErr   error // todo: report configuration errors immediately

	masterURL     string
	servingCAPath string
}

func NewKubeAPIServer(cfg *config.MicroshiftConfig) *KubeAPIServer {
	s := &KubeAPIServer{}
	if err := s.configure(cfg); err != nil {
		s.configureErr = err
	}
	return s
}

func (s *KubeAPIServer) Name() string           { return "kube-apiserver" }
func (s *KubeAPIServer) Dependencies() []string { return []string{"etcd"} }

func (s *KubeAPIServer) configure(cfg *config.MicroshiftConfig) error {
	s.verbosity = cfg.LogVLevel

	certsDir := cryptomaterial.CertsDirectory(microshiftDataDir)
	kubeCSRSignerDir := cryptomaterial.CSRSignerCertDir(certsDir)
	kubeletClientDir := cryptomaterial.KubeAPIServerToKubeletClientCertDir(certsDir)
	clientCABundlePath := cryptomaterial.TotalClientCABundlePath(certsDir)
	aggregatorCAPath := cryptomaterial.CACertPath(cryptomaterial.AggregatorSignerDir(certsDir))
	aggregatorClientCertDir := cryptomaterial.AggregatorClientCertDir(certsDir)
	etcdClientCertDir := cryptomaterial.EtcdAPIServerClientCertDir(certsDir)
	serviceNetworkServingCertDir := cryptomaterial.KubeAPIServerServiceNetworkServingCertDir(certsDir)
	servingCert := cryptomaterial.ServingCertPath(serviceNetworkServingCertDir)
	servingKey := cryptomaterial.ServingKeyPath(serviceNetworkServingCertDir)

	if err := s.configureAuditPolicy(cfg); err != nil {
		return fmt.Errorf("failed to configure kube-apiserver audit policy: %w", err)
	}

	// Get the apiserver port so we can set it as an argument
	apiServerPort, err := cfg.Cluster.ApiServerPort()
	if err != nil {
		return err
	}

	s.masterURL = cfg.Cluster.URL
	s.servingCAPath = cryptomaterial.ServiceAccountTokenCABundlePath(certsDir)

	overrides := &kubecontrolplanev1.KubeAPIServerConfig{
		APIServerArguments: map[string]kubecontrolplanev1.Arguments{
			"advertise-address": {cfg.NodeIP},
			"audit-policy-file": {microshiftDataDir + "/resources/kube-apiserver-audit-policies/default.yaml"},
			"client-ca-file":    {clientCABundlePath},
			"etcd-cafile":       {cryptomaterial.CACertPath(cryptomaterial.EtcdSignerDir(certsDir))},
			"etcd-certfile":     {cryptomaterial.ClientCertPath(etcdClientCertDir)},
			"etcd-keyfile":      {cryptomaterial.ClientKeyPath(etcdClientCertDir)},
			"etcd-servers": {
				"https://127.0.0.1:2379",
			},
			"kubelet-certificate-authority": {cryptomaterial.CABundlePath(kubeCSRSignerDir)},
			"kubelet-client-certificate":    {cryptomaterial.ClientCertPath(kubeletClientDir)},
			"kubelet-client-key":            {cryptomaterial.ClientKeyPath(kubeletClientDir)},

			"proxy-client-cert-file":           {cryptomaterial.ClientCertPath(aggregatorClientCertDir)},
			"proxy-client-key-file":            {cryptomaterial.ClientKeyPath(aggregatorClientCertDir)},
			"requestheader-client-ca-file":     {aggregatorCAPath},
			"service-account-signing-key-file": {microshiftDataDir + "/resources/kube-apiserver/secrets/service-account-key/service-account.key"},
			"service-node-port-range":          {cfg.Cluster.ServiceNodePortRange},
			"tls-cert-file":                    {servingCert},
			"tls-private-key-file":             {servingKey},
			"disable-admission-plugins": {
				"authorization.openshift.io/RestrictSubjectBindings",
				"authorization.openshift.io/ValidateRoleBindingRestriction",
				"autoscaling.openshift.io/ManagementCPUsOverride",
				"config.openshift.io/DenyDeleteClusterConfiguration",
				"config.openshift.io/ValidateAPIServer",
				"config.openshift.io/ValidateAuthentication",
				"config.openshift.io/ValidateConsole",
				"config.openshift.io/ValidateFeatureGate",
				"config.openshift.io/ValidateImage",
				"config.openshift.io/ValidateOAuth",
				"config.openshift.io/ValidateProject",
				"config.openshift.io/ValidateScheduler",
				"image.openshift.io/ImagePolicy",
				"quota.openshift.io/ClusterResourceQuota",
				"quota.openshift.io/ValidateClusterResourceQuota",
			},
			"enable-admission-plugins":              {},
			"send-retry-after-while-not-ready-once": {"true"},
		},
		GenericAPIServerConfig: configv1.GenericAPIServerConfig{
			// from cluster-kube-apiserver-operator
			CORSAllowedOrigins: []string{
				`//127\.0\.0\.1(:|$)`,
				`//localhost(:|$)`,
			},
			ServingInfo: configv1.HTTPServingInfo{
				ServingInfo: configv1.ServingInfo{
					BindAddress:   net.JoinHostPort("0.0.0.0", strconv.Itoa(apiServerPort)),
					MinTLSVersion: string(fixedTLSProfile.MinTLSVersion),
					CipherSuites:  crypto.OpenSSLToIANACipherSuites(fixedTLSProfile.Ciphers),
					NamedCertificates: []configv1.NamedCertificate{
						{
							CertInfo: configv1.CertInfo{
								CertFile: cryptomaterial.ServingCertPath(cryptomaterial.KubeAPIServerExternalServingCertDir(certsDir)),
								KeyFile:  cryptomaterial.ServingKeyPath(cryptomaterial.KubeAPIServerExternalServingCertDir(certsDir)),
							},
						},
						{
							CertInfo: configv1.CertInfo{
								CertFile: cryptomaterial.ServingCertPath(cryptomaterial.KubeAPIServerLocalhostServingCertDir(certsDir)),
								KeyFile:  cryptomaterial.ServingKeyPath(cryptomaterial.KubeAPIServerLocalhostServingCertDir(certsDir)),
							},
						},
						{
							CertInfo: configv1.CertInfo{
								CertFile: servingCert,
								KeyFile:  servingKey,
							},
						},
					},
				},
			},
		},
		ServiceAccountPublicKeyFiles: []string{
			microshiftDataDir + "/resources/kube-apiserver/secrets/service-account-key/service-account.crt",
		},
		ServicesSubnet:        cfg.Cluster.ServiceCIDR,
		ServicesNodePortRange: cfg.Cluster.ServiceNodePortRange,
	}

	overridesBytes, err := json.Marshal(overrides)
	if err != nil {
		return err
	}

	s.kasConfigBytes, err = resourcemerge.MergePrunedProcessConfig(
		&kubecontrolplanev1.KubeAPIServerConfig{},
		map[string]resourcemerge.MergeFunc{
			".apiServerArguments.enable-admission-plugins": func(dst interface{}, src interface{}, currentPath string) (interface{}, error) {
				var result []interface{}

				for _, existing := range dst.([]interface{}) {
					drop := false
					for _, disabled := range overrides.APIServerArguments["disable-admission-plugins"] {
						if existing == disabled {
							drop = true
						}
					}
					if !drop {
						result = append(result, existing)
					}
				}

				result = append(result, src.([]interface{})...)

				return result, nil
			},
		},
		append(baseKubeAPIServerConfigs, overridesBytes)...,
	)
	if err != nil {
		return err
	}

	return nil
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

	path := filepath.Join(microshiftDataDir, "resources", "kube-apiserver-audit-policies", "default.yaml")
	os.MkdirAll(filepath.Dir(path), os.FileMode(0700))
	return os.WriteFile(path, data, 0644)
}

func (s *KubeAPIServer) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	if s.configureErr != nil {
		return fmt.Errorf("configuration failed: %w", s.configureErr)
	}

	defer close(stopped)
	errorChannel := make(chan error, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// run readiness check
	go func() {
		err := wait.PollImmediateWithContext(ctx, time.Second, kubeAPIStartupTimeout*time.Second, func(ctx context.Context) (bool, error) {
			restConfig, err := clientcmd.BuildConfigFromFlags(s.masterURL, "")
			if err != nil {
				return false, err
			}
			if err := rest.SetKubernetesDefaults(restConfig); err != nil {
				return false, err
			}
			restConfig.NegotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme())
			restConfig.CAFile = s.servingCAPath

			restClient, err := rest.UnversionedRESTClientFor(restConfig)
			if err != nil {
				return false, err
			}

			var status int
			if err := restClient.Get().AbsPath("/readyz").Do(ctx).StatusCode(&status).Error(); err != nil {
				klog.Infof("%q not yet ready: %v", s.Name(), err)
				return false, nil
			}
			if status < 200 || status >= 400 {
				klog.Infof("%q not yet ready: received http status %d", s.Name(), status)
				return false, nil
			}
			return true, nil
		})
		if err != nil {
			errorChannel <- fmt.Errorf("readiness check failed: %w", err)
			cancel()
			return
		}
		klog.Infof("%q is ready", s.Name())
		close(ready)
	}()

	fd, err := os.CreateTemp("", "kube-apiserver-config-*.yaml")
	if err != nil {
		return err
	}
	defer func() {
		err := os.Remove(fd.Name())
		if err != nil {
			klog.Warningf("failed to delete temporary kube-apiserver config file: %v", err)
		}
	}()

	err = func() error {
		defer fd.Close()
		_, err = io.Copy(fd, bytes.NewBuffer(s.kasConfigBytes))
		return err
	}()
	if err != nil {
		return err
	}

	// audit logs go here
	os.MkdirAll("/var/log/kube-apiserver", 0700)

	// Carrying a patch for NewAPIServerCommand to use cmd.Context().Done() as the stop channel
	// instead of the channel returned by SetupSignalHandler, which expects to be called at most
	// once in a process.
	cmd := kubeapiserver.NewAPIServerCommand()
	cmd.SetArgs([]string{
		"--openshift-config", fd.Name(),
		"-v", strconv.Itoa(s.verbosity),
	})
	go func() {
		errorChannel <- cmd.ExecuteContext(ctx)
	}()
	return <-errorChannel
}
