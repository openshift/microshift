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
package cmd

import (
	"fmt"
	"io/fs"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"k8s.io/apiserver/pkg/authentication/serviceaccount"
	"k8s.io/apiserver/pkg/authentication/user"
	apiserveroptions "k8s.io/kubernetes/pkg/controlplane/apiserver/options"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
	"github.com/openshift/microshift/pkg/version"

	"k8s.io/klog/v2"
)

func initCerts(cfg *config.Config) (*certchains.CertificateChains, error) {
	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	backupDir, err := migrateLegacyCertLayout(certsDir, version.Get().GitVersion)
	if err != nil {
		return nil, fmt.Errorf("cert layout migration failed: %w", err)
	}

	certChains, err := certSetup(cfg)
	if err != nil {
		// certSetup may leave a partial certs/ directory if it fails mid-way.
		// certchains writes each signer atomically and idempotently: on the next
		// run it skips already-written signers and generates only the missing ones.
		// The caller (run.go) fatals here; greenboot restarts the service and the
		// retry completes cleanly without removing the partial directory.
		return nil, err
	}
	if backupDir != "" {
		if err := os.RemoveAll(backupDir); err != nil {
			klog.Warningf("Failed to remove cert migration backup dir %s: %v", backupDir, err)
		} else {
			klog.Infof("Removed cert migration backup dir %s", backupDir)
		}
	}

	// we cannot just remove the certs dir and regenerate all the certificates
	// because there are some long-lived certs and CAs that shouldn't be swapped
	// - for example system:admin client certs, KAS serving CAs
	regenCerts := certsToRegenerate(certChains)
	for _, c := range regenCerts {
		if err := certChains.Regenerate(c...); err != nil {
			return nil, err
		}
	}

	return certChains, nil
}

// legacyKubeControlPlaneSignerDir returns the old CA path used only for migration detection.
func legacyKubeControlPlaneSignerDir(certsDir string) string {
	return filepath.Join(certsDir, "kube-control-plane-signer")
}

// migrateLegacyCertLayout detects the pre-consolidation cert layout and atomically
// backs it up so certSetup can regenerate a fresh 6-CA hierarchy.
// service-ca is copied back into the live directory after the rename so that
// existing in-cluster service-serving Secrets remain trusted until they are
// reissued by the service-ca controller.
// MicroShift is always a single-node deployment with a single embedded etcd member,
// so peer-CA rotation during migration requires no cross-member coordination.
func migrateLegacyCertLayout(certsDir, gitVersion string) (string, error) {
	if _, err := os.Stat(legacyKubeControlPlaneSignerDir(certsDir)); err != nil {
		if os.IsNotExist(err) {
			return "", nil // fresh install or already migrated
		}
		return "", fmt.Errorf("failed to check legacy cert layout: %w", err)
	}
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	// gitVersion can contain '/' (e.g. from EMBEDDED_GIT_TAG); replace to avoid
	// creating an unintended subdirectory in the backup path.
	safeVer := strings.ReplaceAll(gitVersion, "/", "_")
	backupDir := certsDir + ".backup." + safeVer + "." + ts
	klog.Infof("Migrating cert layout: backing up %s → %s", certsDir, backupDir)
	if err := os.Rename(certsDir, backupDir); err != nil {
		return "", err
	}
	// rollback removes the partial certsDir and renames backupDir back so the
	// next startup sees the original layout intact and retries the migration.
	rollback := func(cause error) error {
		if rmErr := os.RemoveAll(certsDir); rmErr != nil {
			return fmt.Errorf("%w (rollback cleanup failed: %v)", cause, rmErr)
		}
		if rbErr := os.Rename(backupDir, certsDir); rbErr != nil {
			return fmt.Errorf("%w (rollback rename failed: %v)", cause, rbErr)
		}
		return cause
	}
	// Recreate the live directory and restore service-ca so that existing
	// service-serving Secrets remain trusted through the CA transition.
	if err := os.MkdirAll(certsDir, 0700); err != nil {
		return "", rollback(fmt.Errorf("failed to recreate certs directory after migration: %w", err))
	}
	if err := os.CopyFS(
		filepath.Join(certsDir, "service-ca"),
		os.DirFS(filepath.Join(backupDir, "service-ca")),
	); err != nil {
		return "", rollback(fmt.Errorf("failed to restore service-ca after migration: %w", err))
	}
	// os.CopyFS does not preserve file permissions; restrict all CA private keys.
	if err := filepath.WalkDir(filepath.Join(certsDir, "service-ca"), func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(path, ".key") {
			return os.Chmod(path, 0600)
		}
		return nil
	}); err != nil {
		return "", rollback(fmt.Errorf("failed to set permissions on restored service-ca keys: %w", err))
	}
	return backupDir, nil
}

func certSetup(cfg *config.Config) (*certchains.CertificateChains, error) {
	builder, err := certificateChainsSetup(cfg, config.DataDir)
	if err != nil {
		return nil, err
	}

	certChains, err := builder.Complete()
	if err != nil {
		return nil, err
	}

	saKeyDir := filepath.Join(config.DataDir, "/resources/kube-apiserver/secrets/service-account-key")
	if err := util.EnsureKeyPair(
		filepath.Join(saKeyDir, "service-account.pub"),
		filepath.Join(saKeyDir, "service-account.key"),
	); err != nil {
		return nil, err
	}

	cfg.Ingress.ServingCertificate, cfg.Ingress.ServingKey, err = certChains.GetCertKey("ingress-ca", "router-default-serving")
	if err != nil {
		return nil, err
	}

	return certChains, nil
}

//nolint:ireturn // Callers choose read-only inventory loading or certificate creation.
func certificateChainsSetup(cfg *config.Config, dataDir string) (certchains.CertificateChainsBuilder, error) {
	// Anchor certificate expiration to the next day. This forces
	// homogenous expiry dates for all certificates with the same validity.
	startTime := time.Now()
	nextMidnight := time.Date(
		startTime.Year(),
		startTime.Month(),
		startTime.Day()+1,
		0, 0, 0, 0,
		startTime.Location(),
	)
	alignValidity := func(baseValidity time.Duration) time.Duration {
		targetExpiration := nextMidnight.Add(baseValidity)
		return time.Until(targetExpiration)
	}

	_, svcNet, err := net.ParseCIDR(cfg.Network.ServiceNetwork[0])
	if err != nil {
		return nil, err
	}

	_, apiServerServiceIP, err := apiserveroptions.ServiceIPRange(*svcNet)
	if err != nil {
		return nil, err
	}

	certsDir := cryptomaterial.CertsDirectory(dataDir)

	builder := certchains.NewCertificateChains(
		//------------------------------
		// CLIENT CA
		//------------------------------
		certchains.NewCertificateSigner(
			"client-ca",
			cryptomaterial.ClientCADir(certsDir),
			alignValidity(cryptomaterial.LongLivedCertificateValidity),
		).WithService("authentication").WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "kube-controller-manager",
					Service:        "kube-controller-manager",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-controller-manager"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "kube-scheduler",
					Service:        "kube-scheduler",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-scheduler"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "cluster-policy-controller",
					Service:        "cluster-policy-controller",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-controller-manager"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "route-controller-manager",
					Service:        "route-controller-manager",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: serviceaccount.UserInfo("openshift-route-controller-manager", "route-controller-manager-sa", ""),
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "kube-apiserver-to-kubelet-client",
					Service:        "kube-apiserver",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-apiserver", Groups: []string{"kube-master"}},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "metrics-server-kubelet-client",
					Service:        "metrics-server",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "system:metrics-server"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "admin-kubeconfig-client",
					Service:        "authentication",
					Validity:       alignValidity(cryptomaterial.LongLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyExtended,
				},
				UserInfo: &user.DefaultInfo{Name: "system:admin", Groups: []string{"system:masters"}},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "openshift-observability-client",
					Service:        "observability",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "openshift-observability-client"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "kubelet-client",
					Service:        "kubelet",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				// userinfo per https://kubernetes.io/docs/reference/access-authn-authz/node/#overview
				UserInfo: &user.DefaultInfo{Name: "system:node:" + cfg.CanonicalNodeName(), Groups: []string{"system:nodes"}},
			},
		),

		//------------------------------
		// SERVING CA
		//------------------------------
		certchains.NewCertificateSigner(
			"serving-ca",
			cryptomaterial.ServingCADir(certsDir),
			alignValidity(cryptomaterial.LongLivedCertificateValidity),
		).WithService("kube-apiserver").WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:     "kube-apiserver-serving",
					Service:  "kube-apiserver",
					Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity),
				},
				Hostnames: append([]string{
					"kubernetes", "kubernetes.default", "kubernetes.default.svc",
					"kubernetes.default.svc.cluster.local",
					"openshift", "openshift.default", "openshift.default.svc",
					"openshift.default.svc.cluster.local",
					"api." + cfg.DNS.BaseDomain,
					"api-int." + cfg.DNS.BaseDomain,
					cfg.ApiServer.AdvertiseAddress,
					apiServerServiceIP.String(),
					"localhost", "127.0.0.1", "::1",
					cfg.Node.HostnameOverride,
					cfg.Node.NodeIP,
				}, cfg.ApiServer.SubjectAltNames...),
			},
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:     "kubelet-server",
					Service:  "kubelet",
					Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity),
				},
				Hostnames: []string{cfg.Node.HostnameOverride, cfg.Node.NodeIP},
			},
		),

		//------------------------------
		// PEER CA
		//------------------------------
		certchains.NewCertificateSigner(
			"peer-ca",
			cryptomaterial.PeerCADir(certsDir),
			alignValidity(cryptomaterial.LongLivedCertificateValidity),
		).WithService("etcd").WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "apiserver-etcd-client",
					Service:        "kube-apiserver",
					Validity:       alignValidity(cryptomaterial.LongLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyExtended,
				},
				UserInfo: &user.DefaultInfo{Name: "etcd", Groups: []string{"etcd"}},
			},
		).WithPeerCertificiates(
			&certchains.PeerCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "etcd-peer",
					Service:        "etcd",
					Validity:       alignValidity(cryptomaterial.LongLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyExtended,
				},
				UserInfo:  &user.DefaultInfo{Name: "system:etcd-peer:etcd-client", Groups: []string{"system:etcd-peers"}},
				Hostnames: []string{"localhost", cfg.Node.HostnameOverride, cfg.Node.NodeIP},
			},
			&certchains.PeerCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "etcd-serving",
					Service:        "etcd",
					Validity:       alignValidity(cryptomaterial.LongLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyExtended,
				},
				UserInfo:  &user.DefaultInfo{Name: "system:etcd-server:etcd-client", Groups: []string{"system:etcd-servers"}},
				Hostnames: []string{"localhost", cfg.Node.HostnameOverride, cfg.Node.NodeIP},
			},
		),

		//------------------------------
		// UNCHANGED CAs
		//------------------------------
		certchains.NewCertificateSigner(
			"aggregator-signer",
			cryptomaterial.AggregatorSignerDir(certsDir),
			alignValidity(cryptomaterial.ShortLivedCertificateValidity),
		).WithService("kube-apiserver").WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:           "aggregator-client",
					Service:        "kube-apiserver",
					Validity:       alignValidity(cryptomaterial.ShortLivedCertificateValidity),
					RotationPolicy: certchains.RotationPolicyStandard,
				},
				UserInfo: &user.DefaultInfo{Name: "system:openshift-aggregator"},
			},
		),

		certchains.NewCertificateSigner(
			"service-ca",
			cryptomaterial.ServiceCADir(certsDir),
			alignValidity(cryptomaterial.LongLivedCertificateValidity),
		).WithService("service-ca").WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:     "route-controller-manager-serving",
					Service:  "route-controller-manager",
					Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity),
				},
				Hostnames: []string{
					"route-controller-manager.openshift-route-controller-manager.svc",
					"route-controller-manager.openshift-route-controller-manager.svc.cluster.local",
				},
			},
		),

		certchains.NewCertificateSigner(
			"ingress-ca",
			cryptomaterial.IngressCADir(certsDir),
			alignValidity(cryptomaterial.LongLivedCertificateValidity),
		).WithService("ingress").WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:     "router-default-serving",
					Service:  "ingress",
					Validity: alignValidity(cryptomaterial.ShortLivedCertificateValidity),
				},
				Hostnames: []string{
					"*.apps." + cfg.DNS.BaseDomain,
				},
			},
		),
	).WithCABundle(
		cryptomaterial.TotalClientCABundlePath(certsDir),
		[]string{"client-ca"},
	).WithCABundle(
		cryptomaterial.KubeletClientCAPath(certsDir),
		[]string{"client-ca"},
	).WithCABundle(
		cryptomaterial.KubeletServingCAPath(certsDir),
		[]string{"serving-ca"},
	).WithCABundle(
		cryptomaterial.ServiceAccountTokenCABundlePath(certsDir),
		[]string{"serving-ca"},
	)

	return builder, nil
}

func initKubeconfigs(
	cfg *config.Config,
	certChains *certchains.CertificateChains,
) error {
	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	servingCAPEM, err := os.ReadFile(cryptomaterial.CACertPath(cryptomaterial.ServingCADir(certsDir)))
	if err != nil {
		return fmt.Errorf("failed to load serving CA: %v", err)
	}

	adminKubeconfigCertPEM, adminKubeconfigKeyPEM, err := certChains.GetCertKey("client-ca", "admin-kubeconfig-client")
	if err != nil {
		return err
	}

	u, err := url.Parse(cfg.ApiServer.URL)
	if err != nil {
		return fmt.Errorf("failed to parse cluster URL: %v", err)
	}

	// Generate one kubeconfigs per name
	for _, name := range append(cfg.ApiServer.SubjectAltNames, cfg.Node.HostnameOverride) {
		u.Host = net.JoinHostPort(name, strconv.Itoa(cfg.ApiServer.Port))
		if err := util.KubeConfigWithClientCerts(
			cfg.KubeConfigAdminPath(name),
			u.String(),
			servingCAPEM,
			adminKubeconfigCertPEM,
			adminKubeconfigKeyPEM,
		); err != nil {
			return err
		}
	}

	if err := cleanupStaleKubeconfigs(cfg, cfg.KubeConfigRootAdminPath()); err != nil {
		klog.Warningf("Unable to remove stale kubeconfigs: %v", err)
	}

	// Generate kubeconfigs for named certificates
	for _, customCert := range cfg.ApiServer.NamedCertificates {
		klog.Infof("Parsing certificate file: %s", customCert.CertPath)

		certsSNIs, err := util.GetSNIsFromCert(customCert.CertPath, customCert.Names)
		if err != nil {
			klog.Warningf("Unparsable certificates are ignored, error: %v", err)
			continue
		}

		nonWildcardSNI := 0
		// check for wildcard only certificate
		for _, sni := range certsSNIs {
			if !util.IsWildcardDNS(sni) {
				nonWildcardSNI++
			}
		}

		if nonWildcardSNI == 0 {
			klog.Infof("Only wildcard SNI found - placing NodeIP (%s) in kubeconfig API address (server)", cfg.Node.NodeIP)
			certsSNIs = []string{cfg.Node.NodeIP}
		}

		// iterate over the SNIs and generate kubeconfig files
		for _, dns := range certsSNIs {
			// check if SNI is allowed
			if !util.VerifyAllowedSNI(cfg.ApiServer.AdvertiseAddress, cfg.Network.ClusterNetwork, cfg.Network.ServiceNetwork, dns) {
				klog.Infof("Skipping kubeconfig generation, Certificate SNI is not allowed for %s", dns)
				continue
			}

			if util.IsWildcardDNS(dns) {
				klog.Infof("Skipping kubeconfig generation for wildcard DNS %s", dns)
				continue
			}
			klog.Infof("Generating kubeconfig for DNS %s", dns)
			ul, err := url.Parse("https://" + net.JoinHostPort(dns, strconv.Itoa(cfg.ApiServer.Port)))
			if err != nil {
				klog.Errorf("Error generating kubeconfig for %s: %v", ul, err)
				continue
			}

			if err := util.KubeConfigWithClientCerts(
				cfg.KubeConfigAdminPath(dns),
				ul.String(),
				[]byte{},
				adminKubeconfigCertPEM,
				adminKubeconfigKeyPEM,
			); err != nil {
				return err
			}
		}
	}

	if err := util.KubeConfigWithClientCerts(
		cfg.KubeConfigPath(config.KubeAdmin),
		cfg.ApiServer.URL,
		servingCAPEM,
		adminKubeconfigCertPEM,
		adminKubeconfigKeyPEM,
	); err != nil {
		return err
	}

	kcmCertPEM, kcmKeyPEM, err := certChains.GetCertKey("client-ca", "kube-controller-manager")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		cfg.KubeConfigPath(config.KubeControllerManager),
		cfg.ApiServer.URL,
		servingCAPEM,
		kcmCertPEM,
		kcmKeyPEM,
	); err != nil {
		return err
	}

	schedulerCertPEM, schedulerKeyPEM, err := certChains.GetCertKey("client-ca", "kube-scheduler")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		cfg.KubeConfigPath(config.KubeScheduler),
		cfg.ApiServer.URL,
		servingCAPEM,
		schedulerCertPEM, schedulerKeyPEM,
	); err != nil {
		return err
	}

	kubeletCertPEM, kubeletKeyPEM, err := certChains.GetCertKey("client-ca", "kubelet-client")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		cfg.KubeConfigPath(config.Kubelet),
		cfg.ApiServer.URL,
		servingCAPEM,
		kubeletCertPEM, kubeletKeyPEM,
	); err != nil {
		return err
	}
	clusterPolicyControllerCertPEM, clusterPolicyControllerKeyPEM, err := certChains.GetCertKey("client-ca", "cluster-policy-controller")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		cfg.KubeConfigPath(config.ClusterPolicyController),
		cfg.ApiServer.URL,
		servingCAPEM,
		clusterPolicyControllerCertPEM, clusterPolicyControllerKeyPEM,
	); err != nil {
		return err
	}

	routeControllerManagerCertPEM, routeControllerManagerKeyPEM, err := certChains.GetCertKey("client-ca", "route-controller-manager")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		cfg.KubeConfigPath(config.RouteControllerManager),
		cfg.ApiServer.URL,
		servingCAPEM,
		routeControllerManagerCertPEM, routeControllerManagerKeyPEM,
	); err != nil {
		return err
	}
	observabilityClientCertPEM, observabilityClientKeyPEM, err := certChains.GetCertKey("client-ca", "openshift-observability-client")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		cfg.KubeConfigPath(config.ObservabilityClient),
		cfg.ApiServer.URL,
		servingCAPEM,
		observabilityClientCertPEM, observabilityClientKeyPEM,
	); err != nil {
		return err
	}
	return nil
}

// certsToRegenerate returns paths to certificates in the given certificate chains
// bundle that need to be regenerated
func certsToRegenerate(cs *certchains.CertificateChains) [][]string {
	regenCerts := [][]string{}
	for _, entry := range cs.Inventory() {
		certPath := entry.Path
		c := entry.Certificate
		if now := time.Now(); now.Before(c.NotBefore) || now.After(c.NotAfter) {
			regenCerts = append(regenCerts, certPath)
		}

		timeLeft := time.Until(c.NotAfter)

		const month = 30 * time.Hour * 24

		if cryptomaterial.IsCertShortLived(&c) {
			// the cert has less than 7 months to live, just rotate
			until := 7 * month
			if timeLeft < until {
				regenCerts = append(regenCerts, certPath)
			}
			continue
		}

		// long lived certs
		if timeLeft < 18*month {
			regenCerts = append(regenCerts, certPath)
		}
	}

	return regenCerts
}

func cleanupStaleKubeconfigs(cfg *config.Config, path string) error {
	currentKubeconfigs := make(map[string]struct{})
	for _, name := range append(cfg.ApiServer.SubjectAltNames, cfg.Node.HostnameOverride) {
		currentKubeconfigs[name] = struct{}{}
	}
	dirs, err := os.ReadDir(path)
	if err != nil {
		return err
	}
	for _, dir := range dirs {
		if !dir.IsDir() {
			continue
		}

		originalName := dir.Name()
		cleanName := filepath.Base(originalName)

		if cleanName != originalName || cleanName == ".." {
			klog.Warningf("Skipping directory with potentially malicious name: %s", originalName)
			continue
		}

		if _, ok := currentKubeconfigs[cleanName]; !ok {
			kubeConfigPath := filepath.Join(path, cleanName)
			if err := os.RemoveAll(kubeConfigPath); err != nil {
				klog.Warningf("Unable to remove %s: %v", kubeConfigPath, err)
			} else {
				klog.Infof("Removed stale kubeconfig %s", kubeConfigPath)
			}
		}
	}
	return nil
}
