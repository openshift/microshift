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
	"crypto/x509"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"

	"k8s.io/apiserver/pkg/authentication/user"
	ctrl "k8s.io/kubernetes/pkg/controlplane"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
)

var microshiftDataDir = config.GetDataDir()

func initCerts(cfg *config.MicroshiftConfig) (*certchains.CertificateChains, error) {
	certChains, err := certSetup(cfg)
	if err != nil {
		return nil, err
	}

	// we cannot just remove the certs dir and regenerate all the certificates
	// because there are some long-lived certs and CAs that shouldn't be swapped
	// - for example system:admin client certs, KAS serving CAs
	const startupLifetimeThreshhold = 0.5 // FIXME: this would mean regen after 5 years of a 10 years-lived CAs
	regenCerts := [][]string{}
	err = certChains.WalkChains(nil, func(certPath []string, c x509.Certificate) error {
		if now := time.Now(); now.Before(c.NotBefore) || now.After(c.NotAfter) {
			regenCerts = append(regenCerts, certPath)
		}

		totalTime := c.NotAfter.Sub(c.NotBefore).Seconds()
		timeElapsed := time.Since(c.NotBefore).Seconds()

		if timeElapsed/totalTime > startupLifetimeThreshhold {
			regenCerts = append(regenCerts, certPath)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	for _, c := range regenCerts {
		if err := certChains.Regenerate(c...); err != nil {
			return nil, err
		}
	}

	return certChains, err
}

func certSetup(cfg *config.MicroshiftConfig) (*certchains.CertificateChains, error) {
	_, svcNet, err := net.ParseCIDR(cfg.Cluster.ServiceCIDR)
	if err != nil {
		return nil, err
	}

	_, apiServerServiceIP, err := ctrl.ServiceIPRange(*svcNet)
	if err != nil {
		return nil, err
	}

	certsDir := cryptomaterial.CertsDirectory(microshiftDataDir)

	certChains, err := certchains.NewCertificateChains(
		// ------------------------------
		// CLIENT CERTIFICATE SIGNERS
		// ------------------------------

		// kube-control-plane-signer
		certchains.NewCertificateSigner(
			"kube-control-plane-signer",
			cryptomaterial.KubeControlPlaneSignerCertDir(certsDir),
			cryptomaterial.KubeControlPlaneSignerCAValidityDays,
		).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "kube-controller-manager",
					ValidityDays: cryptomaterial.ClientCertValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-controller-manager"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "kube-scheduler",
					ValidityDays: cryptomaterial.ClientCertValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-scheduler"},
			}),

		// kube-apiserver-to-kubelet-signer
		certchains.NewCertificateSigner(
			"kube-apiserver-to-kubelet-signer",
			cryptomaterial.KubeAPIServerToKubeletSignerCertDir(certsDir),
			cryptomaterial.KubeAPIServerToKubeletCAValidityDays,
		).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "kube-apiserver-to-kubelet-client",
					ValidityDays: cryptomaterial.ClientCertValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-apiserver", Groups: []string{"kube-master"}},
			}),

		// admin-kubeconfig-signer
		certchains.NewCertificateSigner(
			"admin-kubeconfig-signer",
			cryptomaterial.AdminKubeconfigSignerDir(certsDir),
			cryptomaterial.AdminKubeconfigCAValidityDays,
		).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "admin-kubeconfig-client",
					ValidityDays: cryptomaterial.AdminKubeconfigClientCertValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:admin", Groups: []string{"system:masters"}},
			}),

		// kubelet + CSR signing chain
		certchains.NewCertificateSigner(
			"kubelet-signer",
			cryptomaterial.KubeletCSRSignerSignerCertDir(certsDir),
			cryptomaterial.KubeControllerManagerCSRSignerSignerCAValidityDays,
		).WithSubCAs(
			certchains.NewCertificateSigner(
				"kube-csr-signer",
				cryptomaterial.CSRSignerCertDir(certsDir),
				cryptomaterial.KubeControllerManagerCSRSignerCAValidityDays,
			).WithClientCertificates(
				&certchains.ClientCertificateSigningRequestInfo{
					CSRMeta: certchains.CSRMeta{
						Name:         "kubelet-client",
						ValidityDays: cryptomaterial.ClientCertValidityDays,
					},
					// userinfo per https://kubernetes.io/docs/reference/access-authn-authz/node/#overview
					UserInfo: &user.DefaultInfo{Name: "system:node:" + cfg.NodeName, Groups: []string{"system:nodes"}},
				},
			).WithServingCertificates(
				&certchains.ServingCertificateSigningRequestInfo{
					CSRMeta: certchains.CSRMeta{
						Name:         "kubelet-server",
						ValidityDays: cryptomaterial.ServingCertValidityDays,
					},
					Hostnames: []string{cfg.NodeName, cfg.NodeIP},
				},
			),
		),
		certchains.NewCertificateSigner(
			"aggregator-signer",
			cryptomaterial.AggregatorSignerDir(certsDir),
			cryptomaterial.AggregatorFrontProxySignerCAValidityDays,
		).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "aggregator-client",
					ValidityDays: cryptomaterial.ClientCertValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:openshift-aggregator"},
			},
		),

		//------------------------------
		// SERVING CERTIFICATE SIGNERS
		//------------------------------
		certchains.NewCertificateSigner(
			"service-ca",
			cryptomaterial.ServiceCADir(certsDir),
			cryptomaterial.ServiceCAValidityDays,
		).WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "route-controller-manager-serving",
					ValidityDays: cryptomaterial.ServiceCAServingCertValidityDays,
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
			cryptomaterial.IngressSignerCAValidityDays,
		).WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "router-default-serving",
					ValidityDays: cryptomaterial.IngressServingCertValidityDays,
				},
				Hostnames: []string{
					"router-default.apps." + cfg.Cluster.Domain,
				},
			},
		),

		// this signer replaces the loadbalancer signers of OCP, we don't need those
		// in Microshift
		certchains.NewCertificateSigner(
			"kube-apiserver-external-signer",
			cryptomaterial.KubeAPIServerExternalSigner(certsDir),
			cryptomaterial.KubeAPIServerServingSignerCAValidityDays,
		).WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "kube-external-serving",
					ValidityDays: cryptomaterial.KubeAPIServerServingCertValidityDays,
				},
				Hostnames: []string{
					cfg.NodeName,
				},
			},
		),

		certchains.NewCertificateSigner(
			"kube-apiserver-localhost-signer",
			cryptomaterial.KubeAPIServerLocalhostSigner(certsDir),
			cryptomaterial.KubeAPIServerServingSignerCAValidityDays,
		).WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "kube-apiserver-localhost-serving",
					ValidityDays: cryptomaterial.KubeAPIServerServingCertValidityDays,
				},
				Hostnames: []string{
					"127.0.0.1",
					"localhost",
				},
			},
		),

		certchains.NewCertificateSigner(
			"kube-apiserver-service-network-signer",
			cryptomaterial.KubeAPIServerServiceNetworkSigner(certsDir),
			cryptomaterial.KubeAPIServerServingSignerCAValidityDays,
		).WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "kube-apiserver-service-network-serving",
					ValidityDays: cryptomaterial.KubeAPIServerServingCertValidityDays,
				},
				Hostnames: []string{
					"kubernetes",
					"kubernetes.default",
					"kubernetes.default.svc",
					"kubernetes.default.svc.cluster.local",
					"openshift",
					"openshift.default",
					"openshift.default.svc",
					"openshift.default.svc.cluster.local",
					apiServerServiceIP.String(),
				},
			},
		),

		//------------------------------
		// 	ETCD CERTIFICATE SIGNER
		//------------------------------
		certchains.NewCertificateSigner(
			"etcd-signer",
			cryptomaterial.EtcdSignerDir(certsDir),
			cryptomaterial.EtcdSignerCAValidityDays,
		).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "apiserver-etcd-client",
					ValidityDays: 10 * 365,
				},
				UserInfo: &user.DefaultInfo{Name: "etcd", Groups: []string{"etcd"}},
			},
		).WithPeerCertificiates(
			&certchains.PeerCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "etcd-peer",
					ValidityDays: 3 * 365,
				},
				UserInfo:  &user.DefaultInfo{Name: "system:etcd-peer:etcd-client", Groups: []string{"system:etcd-peers"}},
				Hostnames: []string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName},
			},
			&certchains.PeerCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "etcd-serving",
					ValidityDays: 3 * 365,
				},
				UserInfo:  &user.DefaultInfo{Name: "system:etcd-server:etcd-client", Groups: []string{"system:etcd-servers"}},
				Hostnames: []string{"localhost", "127.0.0.1", cfg.NodeIP, cfg.NodeName},
			},
		),
	).WithCABundle(
		cryptomaterial.TotalClientCABundlePath(certsDir),
		[]string{"kube-control-plane-signer"},
		[]string{"kube-apiserver-to-kubelet-signer"},
		[]string{"admin-kubeconfig-signer"},
		[]string{"kubelet-signer"},
		[]string{"kubelet-signer", "kube-csr-signer"},
	).WithCABundle(
		cryptomaterial.KubeletClientCAPath(certsDir),
		[]string{"kube-control-plane-signer"},
		[]string{"kube-apiserver-to-kubelet-signer"},
		[]string{"admin-kubeconfig-signer"},
		[]string{"kubelet-signer"},
		[]string{"kubelet-signer", "kube-csr-signer"},
	).WithCABundle(
		cryptomaterial.ServiceAccountTokenCABundlePath(certsDir),
		[]string{"kube-apiserver-external-signer"},
		[]string{"kube-apiserver-localhost-signer"},
		[]string{"kube-apiserver-service-network-signer"},
	).Complete()

	if err != nil {
		return nil, err
	}

	if err := util.GenKeys(filepath.Join(microshiftDataDir, "/resources/kube-apiserver/secrets/service-account-key"),
		"service-account.crt", "service-account.key"); err != nil {
		return nil, err
	}

	cfg.Ingress.ServingCertificate, cfg.Ingress.ServingKey, err = certChains.GetCertKey("ingress-ca", "router-default-serving")
	if err != nil {
		return nil, err
	}

	return certChains, nil
}

func initKubeconfigs(
	cfg *config.MicroshiftConfig,
	certChains *certchains.CertificateChains,
) error {
	inClusterTrustBundlePEM, err := os.ReadFile(cryptomaterial.ServiceAccountTokenCABundlePath(cryptomaterial.CertsDirectory(microshiftDataDir)))
	if err != nil {
		return fmt.Errorf("failed to load the in-cluster trust bundle: %v", err)
	}

	adminKubeconfigCertPEM, adminKubeconfigKeyPEM, err := certChains.GetCertKey("admin-kubeconfig-signer", "admin-kubeconfig-client")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		cfg.KubeConfigPath(config.KubeAdmin),
		cfg.Cluster.URL,
		inClusterTrustBundlePEM,
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
		cfg.KubeConfigPath(config.KubeControllerManager),
		cfg.Cluster.URL,
		inClusterTrustBundlePEM,
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
		cfg.KubeConfigPath(config.KubeScheduler),
		cfg.Cluster.URL,
		inClusterTrustBundlePEM,
		schedulerCertPEM, schedulerKeyPEM,
	); err != nil {
		return err
	}

	kubeletCertPEM, kubeletKeyPEM, err := certChains.GetCertKey("kubelet-signer", "kube-csr-signer", "kubelet-client")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		cfg.KubeConfigPath(config.Kubelet),
		cfg.Cluster.URL,
		inClusterTrustBundlePEM,
		kubeletCertPEM, kubeletKeyPEM,
	); err != nil {
		return err
	}
	return nil
}
