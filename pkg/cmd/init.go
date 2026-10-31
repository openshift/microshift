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
	"net"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/openshift/microshift/pkg/admin/certificates"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"

	"k8s.io/klog/v2"
)

func initCerts(cfg *config.Config) (*certchains.CertificateChains, error) {
	certChains, err := certSetup(cfg)
	if err != nil {
		return nil, err
	}

	// we cannot just remove the certs dir and regenerate all the certificates
	// because there are some long-lived certs and CAs that shouldn't be swapped
	// - for example system:admin client certs, KAS serving CAs
	regenCerts, err := certsToRegenerate(certChains.Inventory(), time.Now())
	if err != nil {
		return nil, err
	}
	for _, c := range regenCerts {
		if err := certChains.Regenerate(c...); err != nil {
			return nil, err
		}
	}

	return certChains, nil
}

func certSetup(cfg *config.Config) (*certchains.CertificateChains, error) {
	builder, err := certificates.NewChains(cfg, config.DataDir)
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

func initKubeconfigs(
	cfg *config.Config,
	certChains *certchains.CertificateChains,
	dataDir string,
) error {
	kubeconfigPath := func(id config.KubeConfigID) string {
		return filepath.Join(dataDir, "resources", string(id), "kubeconfig")
	}
	adminRoot := filepath.Join(dataDir, "resources", string(config.KubeAdmin))
	externalTrustPEM, err := os.ReadFile(cryptomaterial.CACertPath(cryptomaterial.KubeAPIServerExternalSigner(cryptomaterial.CertsDirectory(dataDir))))
	if err != nil {
		return fmt.Errorf("failed to load the external trust signer: %v", err)
	}
	internalTrustPEM, err := os.ReadFile(cryptomaterial.CACertPath(cryptomaterial.KubeAPIServerLocalhostSigner(cryptomaterial.CertsDirectory(dataDir))))
	if err != nil {
		return fmt.Errorf("failed to load the internal trust signer: %v", err)
	}

	adminKubeconfigCertPEM, adminKubeconfigKeyPEM, err := certChains.GetCertKey("admin-kubeconfig-signer", "admin-kubeconfig-client")
	if err != nil {
		return err
	}

	u, err := url.Parse(cfg.ApiServer.URL)
	if err != nil {
		return fmt.Errorf("failed to parse cluster URL: %v", err)
	}

	// Generate one kubeconfigs per name
	for _, name := range append(cfg.ApiServer.SubjectAltNames, cfg.Node.HostnameOverride) {
		path, err := adminKubeconfigPath(adminRoot, name)
		if err != nil {
			return err
		}
		u.Host = net.JoinHostPort(name, strconv.Itoa(cfg.ApiServer.Port))
		if err := util.KubeConfigWithClientCerts(
			path,
			u.String(),
			externalTrustPEM,
			adminKubeconfigCertPEM,
			adminKubeconfigKeyPEM,
		); err != nil {
			return err
		}
	}

	if err := cleanupStaleKubeconfigs(cfg, adminRoot); err != nil {
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

			path, err := adminKubeconfigPath(adminRoot, dns)
			if err != nil {
				return err
			}
			if err := util.KubeConfigWithClientCerts(
				path,
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
		kubeconfigPath(config.KubeAdmin),
		cfg.ApiServer.URL,
		internalTrustPEM,
		adminKubeconfigCertPEM,
		adminKubeconfigKeyPEM,
	); err != nil {
		return err
	}

	kcmCertPEM, kcmKeyPEM, err := certChains.GetCertKey("kube-control-plane-signer", "kube-controller-manager")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		kubeconfigPath(config.KubeControllerManager),
		cfg.ApiServer.URL,
		internalTrustPEM,
		kcmCertPEM,
		kcmKeyPEM,
	); err != nil {
		return err
	}

	schedulerCertPEM, schedulerKeyPEM, err := certChains.GetCertKey("kube-control-plane-signer", "kube-scheduler")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		kubeconfigPath(config.KubeScheduler),
		cfg.ApiServer.URL,
		internalTrustPEM,
		schedulerCertPEM, schedulerKeyPEM,
	); err != nil {
		return err
	}

	kubeletCertPEM, kubeletKeyPEM, err := certChains.GetCertKey("kubelet-signer", "kube-csr-signer", "kubelet-client")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		kubeconfigPath(config.Kubelet),
		cfg.ApiServer.URL,
		internalTrustPEM,
		kubeletCertPEM, kubeletKeyPEM,
	); err != nil {
		return err
	}
	clusterPolicyControllerCertPEM, clusterPolicyControllerKeyPEM, err := certChains.GetCertKey("kube-control-plane-signer", "cluster-policy-controller")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		kubeconfigPath(config.ClusterPolicyController),
		cfg.ApiServer.URL,
		internalTrustPEM,
		clusterPolicyControllerCertPEM, clusterPolicyControllerKeyPEM,
	); err != nil {
		return err
	}

	routeControllerManagerCertPEM, routeControllerManagerKeyPEM, err := certChains.GetCertKey("kube-control-plane-signer", "route-controller-manager")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		kubeconfigPath(config.RouteControllerManager),
		cfg.ApiServer.URL,
		internalTrustPEM,
		routeControllerManagerCertPEM, routeControllerManagerKeyPEM,
	); err != nil {
		return err
	}
	observabilityClientCertPEM, observabilityClientKeyPEM, err := certChains.GetCertKey("admin-kubeconfig-signer", "openshift-observability-client")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		kubeconfigPath(config.ObservabilityClient),
		cfg.ApiServer.URL,
		internalTrustPEM,
		observabilityClientCertPEM, observabilityClientKeyPEM,
	); err != nil {
		return err
	}
	return nil
}

// Names become directory components. Keep all writes within the configured
// resource tree, particularly when that tree is a renewal staging directory.
func adminKubeconfigPath(root, name string) (string, error) {
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name {
		return "", fmt.Errorf("invalid name for generated administrator kubeconfig")
	}
	return filepath.Join(root, name, "kubeconfig"), nil
}

// certsToRegenerate returns paths to certificates in the given certificate chains
// bundle that need to be regenerated
func certsToRegenerate(inventory certchains.CertificateInventory, now time.Time) ([][]string, error) {
	regenCerts := [][]string{}
	for _, entry := range inventory {
		status, err := entry.StatusAt(now)
		if err != nil {
			return nil, err
		}
		if status != certchains.CertificateStatusHealthy {
			regenCerts = append(regenCerts, entry.Path)
		}
	}
	return regenCerts, nil
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
