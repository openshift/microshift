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
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"go.etcd.io/etcd/client/pkg/v3/transport"
	clientv3 "go.etcd.io/etcd/client/v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	kubeapiserver "k8s.io/kubernetes/cmd/kube-apiserver/app"
	hostassignmentv1 "k8s.io/kubernetes/openshift-kube-apiserver/admission/route/apis/hostassignment/v1"
	"sigs.k8s.io/yaml"

	configv1 "github.com/openshift/api/config/v1"
	kubecontrolplanev1 "github.com/openshift/api/kubecontrolplane/v1"
	"github.com/openshift/library-go/pkg/operator/resource/resourcemerge"

	embedded "github.com/openshift/microshift/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/config/apiserver"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

const (
	kubeAPIStartupTimeout = 60
	// rbacHookDeadlockTimeout is the time to wait for the RBAC bootstrap hook
	// before declaring a deadlock. This is shorter than kubeAPIStartupTimeout
	// to allow for faster recovery.
	rbacHookDeadlockTimeout = 15
	// rbacHookCheckInterval is how often to check the RBAC hook status
	rbacHookPollDelayStart = 5 * time.Second
	rbacHookCheckInterval  = 2
	// rbacHookMaxWaitDuration is the absolute maximum time to wait for the RBAC hook
	// regardless of etcd health state changes. This prevents flapping from extending
	// detection indefinitely.
	rbacHookMaxWaitDuration = 30 * time.Second
)

var (
	baseKubeAPIServerConfigs = [][]byte{
		// todo: a nice way to generate the baseline kas config for microshift
		embedded.MustAsset("controllers/kube-apiserver/defaultconfig.yaml"),
		embedded.MustAsset("controllers/kube-apiserver/config-overrides.yaml"),
	}
)
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
	configuration  *config.Config

	masterURL        string
	servingCAPath    string
	advertiseAddress string
}

func NewKubeAPIServer(cfg *config.Config) *KubeAPIServer {
	s := &KubeAPIServer{
		configuration: cfg,
	}
	return s
}

func (s *KubeAPIServer) Name() string           { return "kube-apiserver" }
func (s *KubeAPIServer) Dependencies() []string { return []string{"etcd", "network-configuration"} }

func (s *KubeAPIServer) configure(ctx context.Context, cfg *config.Config) error {
	s.verbosity = cfg.GetVerbosity()

	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
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

	s.masterURL = cfg.ApiServer.URL
	s.servingCAPath = cryptomaterial.ServiceAccountTokenCABundlePath(certsDir)
	s.advertiseAddress = cfg.ApiServer.AdvertiseAddresses[0]

	namedCerts := []configv1.NamedCertificate{
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
	}
	if len(cfg.ApiServer.NamedCertificates) > 0 {
		for _, namedCertsCfg := range cfg.ApiServer.NamedCertificates {
			//Validate the cert is non-destructive
			certAllowed, err := util.IsCertAllowed(cfg.ApiServer.AdvertiseAddresses[0], cfg.Network.ClusterNetwork, cfg.Network.ServiceNetwork, namedCertsCfg.CertPath, namedCertsCfg.Names)
			if err != nil {
				klog.Warningf("Failed to read NamedCertificate from %s - ignoring: %v", namedCertsCfg.CertPath, err)
				continue
			}

			if !certAllowed {
				klog.Warningf("Certificate %v is not allowed - ignoring", namedCertsCfg)
				continue
			}

			cert := []configv1.NamedCertificate{
				{
					Names: namedCertsCfg.Names,
					CertInfo: configv1.CertInfo{
						CertFile: namedCertsCfg.CertPath,
						KeyFile:  namedCertsCfg.KeyPath,
					},
				},
			}
			// prepend the named certs to the beginning of the slice (so it will take precedence for same SNI)
			namedCerts = append(cert, namedCerts...)
		}
	}

	etcdServers, err := discoverEtcdServers(ctx, s.configuration.BootstrapKubeConfigPath())
	if err != nil {
		return fmt.Errorf("failed to discover etcd servers: %w", err)
	}

	featureGateArgs, err := cfg.ApiServer.FeatureGates.ToApiserverArgs()
	if err != nil {
		return fmt.Errorf("failed to convert feature gates to CLI flags: %w", err)
	}
	// Inject required feature gates into the feature-gates argument.
	enabledRequiredFGs := make(sets.Set[string], 0)
	for _, fg := range config.RequiredFeatureGates {
		enabledRequiredFGs.Insert(fg + "=true")
	}
	featureGateArgs = sets.New(featureGateArgs...).Union(enabledRequiredFGs).UnsortedList()

	overrides := &kubecontrolplanev1.KubeAPIServerConfig{
		APIServerArguments: map[string]kubecontrolplanev1.Arguments{
			"advertise-address":             {s.advertiseAddress},
			"audit-policy-file":             {filepath.Join(config.DataDir, "/resources/kube-apiserver-audit-policies/default.yaml")},
			"audit-log-maxage":              {strconv.Itoa(cfg.ApiServer.AuditLog.MaxFileAge)},
			"audit-log-maxbackup":           {strconv.Itoa(cfg.ApiServer.AuditLog.MaxFiles)},
			"audit-log-maxsize":             {strconv.Itoa(cfg.ApiServer.AuditLog.MaxFileSize)},
			"client-ca-file":                {clientCABundlePath},
			"etcd-cafile":                   {cryptomaterial.CACertPath(cryptomaterial.EtcdSignerDir(certsDir))},
			"etcd-certfile":                 {cryptomaterial.ClientCertPath(etcdClientCertDir)},
			"etcd-keyfile":                  {cryptomaterial.ClientKeyPath(etcdClientCertDir)},
			"etcd-servers":                  etcdServers,
			"kubelet-certificate-authority": {cryptomaterial.CABundlePath(kubeCSRSignerDir)},
			"kubelet-client-certificate":    {cryptomaterial.ClientCertPath(kubeletClientDir)},
			"kubelet-client-key":            {cryptomaterial.ClientKeyPath(kubeletClientDir)},
			// MicroShift nodes expose these two types of addresses. In order to support having more than one
			// node with the current approach (which is running a stand alone kubelet and share certificates
			// with the master node) we need to use names only because of the way certificates are generated.
			// The names are fed through the subjectAltNames configuration parameter, and using master node
			// IP in the certificates when having more than one node is not possible due to apiserver SNI
			// limitations. For this, we prefer using names and IPs as a fallback, supporting both single
			// and multi node.
			"kubelet-preferred-address-types": {"Hostname", "InternalIP"},
			"service-cluster-ip-range":        {strings.Join(cfg.Network.ServiceNetwork, ",")},

			"proxy-client-cert-file":           {cryptomaterial.ClientCertPath(aggregatorClientCertDir)},
			"proxy-client-key-file":            {cryptomaterial.ClientKeyPath(aggregatorClientCertDir)},
			"requestheader-client-ca-file":     {aggregatorCAPath},
			"service-account-signing-key-file": {filepath.Join(config.DataDir, "/resources/kube-apiserver/secrets/service-account-key/service-account.key")},
			"service-node-port-range":          {cfg.Network.ServiceNodePortRange},
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
			"shutdown-delay-duration":               {"5s"},
			"feature-gates":                         featureGateArgs,
		},
		GenericAPIServerConfig: configv1.GenericAPIServerConfig{
			AdmissionConfig: configv1.AdmissionConfig{
				PluginConfig: map[string]configv1.AdmissionPluginConfig{
					"route.openshift.io/RouteHostAssignment": {
						Configuration: runtime.RawExtension{
							Object: &hostassignmentv1.HostAssignmentAdmissionConfig{
								TypeMeta: metav1.TypeMeta{
									APIVersion: "route.openshift.io/v1",
									Kind:       "HostAssignmentAdmissionConfig",
								},
								Domain: "apps." + cfg.DNS.BaseDomain,
							},
						},
					},
				},
			},
			// from cluster-kube-apiserver-operator
			CORSAllowedOrigins: []string{
				`//127\.0\.0\.1(:|$)`,
				`//localhost(:|$)`,
			},
			ServingInfo: configv1.HTTPServingInfo{
				ServingInfo: configv1.ServingInfo{
					BindAddress:       net.JoinHostPort("0.0.0.0", strconv.Itoa(cfg.ApiServer.Port)),
					MinTLSVersion:     cfg.ApiServer.TLS.MinVersion,
					CipherSuites:      cfg.ApiServer.TLS.CipherSuites,
					NamedCertificates: namedCerts,
				},
			},
		},
		ServiceAccountPublicKeyFiles: []string{
			filepath.Join(config.DataDir, "/resources/kube-apiserver/secrets/service-account-key/service-account.pub"),
		},
		ServicesNodePortRange: cfg.Network.ServiceNodePortRange,
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

func (s *KubeAPIServer) configureAuditPolicy(cfg *config.Config) error {
	p, err := apiserver.GetPolicy(cfg.ApiServer.AuditLog.Profile)
	if err != nil {
		return err
	}
	path := filepath.Join(config.DataDir, "resources", "kube-apiserver-audit-policies", "default.yaml")
	if err := os.MkdirAll(filepath.Dir(path), os.FileMode(0700)); err != nil {
		return err
	}
	data, err := yaml.Marshal(p)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0400)
}

func (s *KubeAPIServer) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	if err := s.configure(ctx, s.configuration); err != nil {
		return fmt.Errorf("configuration failed: %w", err)
	}

	defer close(stopped)
	errorChannel := make(chan error, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	restConfig, err := clientcmd.BuildConfigFromFlags(s.masterURL, "")
	if err != nil {
		return err
	}
	if err := rest.SetKubernetesDefaults(restConfig); err != nil {
		return err
	}
	restConfig.NegotiatedSerializer = serializer.NewCodecFactory(runtime.NewScheme())
	restConfig.CAFile = s.servingCAPath

	restClient, err := rest.UnversionedRESTClientFor(restConfig)
	if err != nil {
		return err
	}

	// Channel to signal RBAC hook deadlock detection
	rbacDeadlockDetected := make(chan struct{})

	// Run RBAC hook deadlock detector
	go s.detectRBACHookDeadlock(ctx, restClient, rbacDeadlockDetected)

	// Run standard readiness check
	go func() {
		err := wait.PollUntilContextTimeout(ctx, time.Second, kubeAPIStartupTimeout*time.Second, true, func(ctx context.Context) (bool, error) {
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
		defer func() { _ = fd.Close() }()
		_, err = io.Copy(fd, bytes.NewBuffer(s.kasConfigBytes))
		return err
	}()
	if err != nil {
		return err
	}

	// audit logs go here
	if err := os.MkdirAll("/var/log/kube-apiserver", 0700); err != nil {
		return err
	}

	// Carrying a patch for NewAPIServerCommand to use cmd.Context().Done() as the stop channel
	// instead of the channel returned by SetupSignalHandler, which expects to be called at most
	// once in a process.
	cmd := kubeapiserver.NewAPIServerCommand()
	cmd.SetArgs([]string{
		"--openshift-config", fd.Name(),
		"-v", strconv.Itoa(s.verbosity),
	})

	panicChannel := make(chan any, 1)
	go func() {
		defer func() {
			if r := recover(); r != nil {
				panicChannel <- r
			}
		}()
		errorChannel <- cmd.ExecuteContext(ctx)
	}()

	select {
	case err := <-errorChannel:
		return err
	case perr := <-panicChannel:
		panic(perr)
	case <-rbacDeadlockDetected:
		klog.Error("RBAC bootstrap hook deadlock detected - restarting microshift-etcd.scope to recover")
		if err := restartMicroshiftEtcdScope(); err != nil {
			klog.Errorf("Failed to restart microshift-etcd.scope: %v", err)
		}
		return fmt.Errorf("RBAC bootstrap hook deadlock detected after %d seconds", rbacHookDeadlockTimeout)
	}
}

// detectRBACHookDeadlock monitors the RBAC bootstrap hook status and detects deadlock conditions.
// A deadlock is detected when:
// 1. The RBAC hook is not completing (stuck in "not finished" state)
// 2. etcd is healthy and responsive
// This indicates the circular dependency where the hook waits for API server
// while API server waits for the hook.
//
// Closed upstream Kubernetes issues:
// https://github.com/kubernetes/kubernetes/issues/86715
// https://github.com/kubernetes/kubernetes/issues/97119
func (s *KubeAPIServer) detectRBACHookDeadlock(ctx context.Context, restClient rest.Interface, deadlockDetected chan<- struct{}) {
	// Wait a few seconds before starting detection to allow normal startup
	select {
	case <-ctx.Done():
		return
	case <-time.After(rbacHookPollDelayStart):
	}

	checkCount := 0
	maxChecks := int((rbacHookDeadlockTimeout - rbacHookPollDelayStart) / rbacHookCheckInterval) // Account for initial delay
	// Track wall-clock deadline to prevent flapping from extending detection indefinitely
	startTime := time.Now()

	for {
		// Check absolute deadline first - this cannot be reset by etcd state changes
		if time.Since(startTime) >= rbacHookMaxWaitDuration {
			klog.Errorf("RBAC bootstrap hook exceeded maximum wait duration of %v", rbacHookMaxWaitDuration)
			// Only trigger deadlock recovery if we've confirmed the predicate enough times
			if checkCount >= maxChecks {
				break // Fall through to close(deadlockDetected)
			}
			// Timeout but not confirmed deadlock - exit without triggering recovery
			return
		}

		// Check if we've confirmed deadlock enough times
		if checkCount >= maxChecks {
			break // Fall through to close(deadlockDetected)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(rbacHookCheckInterval * time.Second):
		}

		// Check RBAC hook status
		probeCtx, cancel := context.WithTimeout(ctx, time.Second)
		var status int
		err := restClient.Get().
			AbsPath("/readyz/poststarthook/rbac/bootstrap-roles").
			Do(probeCtx).
			StatusCode(&status).
			Error()
		cancel()

		// If hook is ready, no deadlock
		if err == nil && status == 200 {
			klog.V(4).Info("RBAC bootstrap hook completed successfully")
			return
		}

		// If RBAC probe errored, skip this iteration (don't count toward deadlock)
		if err != nil {
			klog.V(4).Infof("RBAC probe error (not counting toward deadlock): %v", err)
			continue
		}

		// Hook not ready (status != 200) - check if etcd is healthy
		etcdHealthy, etcdErr := isEtcdHealthy(ctx)
		if etcdErr != nil {
			klog.V(4).Infof("Could not check etcd health (not counting toward deadlock): %v", etcdErr)
			continue
		}

		if etcdHealthy {
			// Only increment when BOTH conditions are met:
			// RBAC probe returned not-ready AND etcd is healthy
			checkCount++
			klog.Warningf("RBAC bootstrap hook not ready (check %d/%d, elapsed %v), but etcd is healthy - potential deadlock",
				checkCount, maxChecks, time.Since(startTime).Round(time.Second))
		} else {
			// etcd not healthy - not a deadlock, just waiting for etcd
			klog.V(4).Infof("RBAC hook waiting, etcd not yet healthy (check %d/%d)", checkCount, maxChecks)
			// Reset counter since this isn't a deadlock condition
			// Note: wall-clock deadline (startTime) is NOT reset - flapping cannot extend indefinitely
			checkCount = 0
		}
	}

	// Only reached when checkCount >= maxChecks (deadlock confirmed)
	klog.Errorf("RBAC bootstrap hook deadlock confirmed after %v: etcd healthy but hook not completing",
		time.Since(startTime).Round(time.Second))
	close(deadlockDetected)
}

// isEtcdHealthy checks if etcd is responsive by attempting to connect and get status.
func isEtcdHealthy(ctx context.Context) (bool, error) {
	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	etcdAPIServerClientCertDir := cryptomaterial.EtcdAPIServerClientCertDir(certsDir)

	tlsInfo := transport.TLSInfo{
		CertFile:      cryptomaterial.ClientCertPath(etcdAPIServerClientCertDir),
		KeyFile:       cryptomaterial.ClientKeyPath(etcdAPIServerClientCertDir),
		TrustedCAFile: cryptomaterial.CACertPath(cryptomaterial.EtcdSignerDir(certsDir)),
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return false, fmt.Errorf("failed to create TLS config: %w", err)
	}

	// Use a short timeout for health check
	checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"https://localhost:2379"},
		DialTimeout: 1 * time.Second,
		TLS:         tlsConfig,
		Context:     checkCtx,
	})
	if err != nil {
		return false, fmt.Errorf("failed to create etcd client: %w", err)
	}
	defer func() { _ = client.Close() }()

	_, err = client.Status(checkCtx, "localhost:2379")
	if err != nil {
		return false, nil // etcd not healthy, but not an error condition
	}

	return true, nil
}

// restartMicroshiftEtcdScope restarts the microshift-etcd.scope to recover from deadlock.
// This forces a clean restart of etcd which can help break the circular dependency.
func restartMicroshiftEtcdScope() error {
	klog.Info("Stopping microshift-etcd.scope for recovery")

	// Set a timeout in case systemd or DBus stalls and the fail-fast recovery path hangs and Run never returns
	cmdCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stopCmd := exec.CommandContext(cmdCtx, "systemctl", "stop", "microshift-etcd.scope")
	if out, err := stopCmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stop microshift-etcd.scope: %w, output: %s", err, string(out))
	}

	// Wait briefly for cleanup
	time.Sleep(1 * time.Second)

	klog.Info("microshift-etcd.scope stopped - MicroShift will restart")
	return nil
}

func discoverEtcdServers(ctx context.Context, kubeconfigPath string) ([]string, error) {
	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	etcdPeerCertDir := cryptomaterial.EtcdPeerCertDir(certsDir)

	tlsInfo := transport.TLSInfo{
		CertFile:      cryptomaterial.PeerCertPath(etcdPeerCertDir),
		KeyFile:       cryptomaterial.PeerKeyPath(etcdPeerCertDir),
		TrustedCAFile: cryptomaterial.CACertPath(cryptomaterial.EtcdSignerDir(certsDir)),
	}
	tlsConfig, err := tlsInfo.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client TLS config: %v", err)
	}

	client, err := clientv3.New(clientv3.Config{
		DialTimeout: 5 * time.Second,
		Endpoints:   []string{"https://localhost:2379"},
		TLS:         tlsConfig,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}
	defer func() { _ = client.Close() }()

	st, err := client.Status(ctx, "localhost:2379")
	if err != nil {
		return nil, fmt.Errorf("failed to get etcd status: %w", err)
	}

	// If I am not a learner it means I am a voting member, so connecting to my own etcd instance
	// is fine because everything is synced.
	if !st.IsLearner {
		return []string{"https://localhost:2379"}, nil
	}

	// If I am a learner I need to connect to a member, retrieve the list of voting
	// members and connect to all of them.
	kubeconfig, err := clientcmd.LoadFromFile(kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load bootstrap kubeconfig: %w", err)
	}

	if kubeconfig == nil || kubeconfig.Clusters == nil || len(kubeconfig.Clusters) == 0 {
		return nil, fmt.Errorf("invalid bootstrap kubeconfig: no clusters found")
	}

	if len(kubeconfig.Clusters) > 1 {
		return nil, fmt.Errorf("invalid bootstrap kubeconfig: multiple clusters found")
	}

	var etcdHost string
	for _, cluster := range kubeconfig.Clusters {
		etcdHost = cluster.Server
		break
	}

	if etcdHost == "" {
		return nil, fmt.Errorf("failed to extract etcd hostname from bootstrap kubeconfig")
	}

	etcdHost = strings.TrimPrefix(etcdHost, "https://")
	etcdHost, _, _ = net.SplitHostPort(etcdHost)
	etcdHost = fmt.Sprintf("https://%s", net.JoinHostPort(etcdHost, "2379"))
	client, err = clientv3.New(clientv3.Config{
		DialTimeout: 5 * time.Second,
		Endpoints:   []string{etcdHost},
		TLS:         tlsConfig,
		Context:     ctx,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create etcd client: %w", err)
	}

	resp, err := client.MemberList(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve etcd member list: %w", err)
	}

	var members []string
	for _, member := range resp.Members {
		if !member.IsLearner {
			members = append(members, member.ClientURLs...)
		}
	}
	return members, nil
}
