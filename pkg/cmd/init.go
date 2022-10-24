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
	"os"
	"path/filepath"

	"k8s.io/apiserver/pkg/authentication/user"
	ctrl "k8s.io/kubernetes/pkg/controlplane"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"github.com/openshift/microshift/pkg/util/cryptomaterial/certchains"
)

var microshiftDataDir = config.GetDataDir()

func initAll(cfg *config.MicroshiftConfig) error {
	// create CA and keys
	certChains, err := initCerts(cfg)
	if err != nil {
		return err
	}
	// create kubeconfig for kube-scheduler, kubelet,controller-manager
	if err := initKubeconfig(cfg, certChains); err != nil {
		return err
	}

	return nil
}

func initCerts(cfg *config.MicroshiftConfig) (*certchains.CertificateChains, error) {
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
				CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
					Name:         "kube-controller-manager",
					ValidityDays: cryptomaterial.ClientCertValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-controller-manager"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
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
				CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
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
				CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
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
					CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
						Name:         "kubelet-client",
						ValidityDays: cryptomaterial.ClientCertValidityDays,
					},
					// userinfo per https://kubernetes.io/docs/reference/access-authn-authz/node/#overview
					UserInfo: &user.DefaultInfo{Name: "system:node:" + cfg.NodeName, Groups: []string{"system:nodes"}},
				},
			).WithServingCertificates(
				&certchains.ServingCertificateSigningRequestInfo{
					CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
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
				CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
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
				CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
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
				CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
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
				CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
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
				CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
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
				CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
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
				CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
					Name:         "apiserver-etcd-client",
					ValidityDays: 10 * 365,
				},
				UserInfo: &user.DefaultInfo{Name: "etcd", Groups: []string{"etcd"}},
			},
		).WithPeerCertificiates(
			&certchains.PeerCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
					Name:         "etcd-peer",
					ValidityDays: 3 * 365,
				},
				UserInfo:  &user.DefaultInfo{Name: "system:etcd-peer:etcd-client", Groups: []string{"system:etcd-peers"}},
				Hostnames: []string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName},
			},
			&certchains.PeerCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: certchains.CertificateSigningRequestInfo{
					Name:         "etcd-serving",
					ValidityDays: 3 * 365,
				},
				UserInfo:  &user.DefaultInfo{Name: "system:etcd-server:etcd-client", Groups: []string{"system:etcd-servers"}},
				Hostnames: []string{"localhost", "127.0.0.1", cfg.NodeIP, cfg.NodeName},
			},
		),
	).WithCABundle(
		cryptomaterial.TotalClientCABundlePath(certsDir),
		"kube-control-plane-signer",
		"kube-apiserver-to-kubelet-signer",
		"admin-kubeconfig-signer",
		"kubelet-signer",
		// kube-csr-signer is being added below
	).WithCABundle(
		cryptomaterial.KubeletClientCAPath(certsDir),
		"kube-control-plane-signer",
		"kube-apiserver-to-kubelet-signer",
		"admin-kubeconfig-signer",
		"kubelet-signer",
		// kube-csr-signer is being added below
	).WithCABundle(
		cryptomaterial.ServiceAccountTokenCABundlePath(certsDir),
		"kube-apiserver-external-signer",
		"kube-apiserver-localhost-signer",
		"kube-apiserver-service-network-signer",
	).Complete()

	if err != nil {
		return nil, err
	}

	csrSignerCAPEM, err := certChains.GetSigner("kubelet-signer", "kube-csr-signer").GetSignerCertPEM()
	if err != nil {
		return nil, err
	}

	if err := cryptomaterial.AddToKubeletClientCABundle(certsDir, csrSignerCAPEM); err != nil {
		return nil, err
	}

	if err := cryptomaterial.AddToTotalClientCABundle(certsDir, csrSignerCAPEM); err != nil {
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

func initKubeconfig(
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
