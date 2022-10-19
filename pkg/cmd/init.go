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

func initCerts(cfg *config.MicroshiftConfig) (*cryptomaterial.CertificateChains, error) {
	_, svcNet, err := net.ParseCIDR(cfg.Cluster.ServiceCIDR)
	if err != nil {
		return nil, err
	}

	_, apiServerServiceIP, err := ctrl.ServiceIPRange(*svcNet)
	if err != nil {
		return nil, err
	}

	certsDir := cryptomaterial.CertsDirectory(microshiftDataDir)

	certChains, err := cryptomaterial.NewCertificateChains(
		// ------------------------------
		// CLIENT CERTIFICATE SIGNERS
		// ------------------------------

		// kube-control-plane-signer
		cryptomaterial.NewCertificateSigner(
			"kube-control-plane-signer",
			cryptomaterial.KubeControlPlaneSignerCertDir(certsDir),
			cryptomaterial.KubeControlPlaneSignerCAValidityDays,
		).WithClientCertificates(
			&cryptomaterial.ClientCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         "kube-controller-manager",
					ValidityDays: cryptomaterial.ClientCertValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-controller-manager"},
			},
			&cryptomaterial.ClientCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         "kube-scheduler",
					ValidityDays: cryptomaterial.ClientCertValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-scheduler"},
			}),

		// kube-apiserver-to-kubelet-signer
		cryptomaterial.NewCertificateSigner(
			"kube-apiserver-to-kubelet-signer",
			cryptomaterial.KubeAPIServerToKubeletSignerCertDir(certsDir),
			cryptomaterial.KubeAPIServerToKubeletCAValidityDays,
		).WithClientCertificates(
			&cryptomaterial.ClientCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         "kube-apiserver-to-kubelet-client",
					ValidityDays: cryptomaterial.ClientCertValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-apiserver", Groups: []string{"kube-master"}},
			}),

		// admin-kubeconfig-signer
		cryptomaterial.NewCertificateSigner(
			"admin-kubeconfig-signer",
			cryptomaterial.AdminKubeconfigSignerDir(certsDir),
			cryptomaterial.AdminKubeconfigCAValidityDays,
		).WithClientCertificates(
			&cryptomaterial.ClientCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         "admin-kubeconfig-client",
					ValidityDays: cryptomaterial.AdminKubeconfigClientCertValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:admin", Groups: []string{"system:masters"}},
			}),

		// kubelet + CSR signing chain
		cryptomaterial.NewCertificateSigner(
			"kubelet-signer",
			cryptomaterial.KubeletCSRSignerSignerCertDir(certsDir),
			cryptomaterial.KubeControllerManagerCSRSignerSignerCAValidityDays,
		).WithSubCAs(
			cryptomaterial.NewCertificateSigner(
				"kube-csr-signer",
				cryptomaterial.CSRSignerCertDir(certsDir),
				cryptomaterial.KubeControllerManagerCSRSignerCAValidityDays,
			).WithClientCertificates(
				&cryptomaterial.ClientCertificateSigningRequestInfo{
					CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
						Name:         "kubelet-client",
						ValidityDays: cryptomaterial.ClientCertValidityDays,
					},
					// userinfo per https://kubernetes.io/docs/reference/access-authn-authz/node/#overview
					UserInfo: &user.DefaultInfo{Name: "system:node:" + cfg.NodeName, Groups: []string{"system:nodes"}},
				},
			).WithServingCertificates(
				&cryptomaterial.ServingCertificateSigningRequestInfo{
					CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
						Name:         "kubelet-server",
						ValidityDays: cryptomaterial.ServingCertValidityDays,
					},
					Hostnames: []string{cfg.NodeName, cfg.NodeIP},
				},
			),
		),
		cryptomaterial.NewCertificateSigner(
			"aggregator-signer",
			cryptomaterial.AggregatorSignerDir(certsDir),
			cryptomaterial.AggregatorFrontProxySignerCAValidityDays,
		).WithClientCertificates(
			&cryptomaterial.ClientCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         "aggregator-client",
					ValidityDays: cryptomaterial.ClientCertValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:openshift-aggregator"},
			},
		),

		//------------------------------
		// SERVING CERTIFICATE SIGNERS
		//------------------------------
		cryptomaterial.NewCertificateSigner(
			"service-ca",
			cryptomaterial.ServiceCADir(certsDir),
			cryptomaterial.ServiceCAValidityDays,
		).WithServingCertificates(
			&cryptomaterial.ServingCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         "route-controller-manager-serving",
					ValidityDays: cryptomaterial.ServiceCAServingCertValidityDays,
				},
				Hostnames: []string{
					"route-controller-manager.openshift-route-controller-manager.svc",
					"route-controller-manager.openshift-route-controller-manager.svc.cluster.local",
				},
			},
		),

		cryptomaterial.NewCertificateSigner(
			"ingress-ca",
			cryptomaterial.IngressCADir(certsDir),
			cryptomaterial.IngressSignerCAValidityDays,
		).WithServingCertificates(
			&cryptomaterial.ServingCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
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
		cryptomaterial.NewCertificateSigner(
			"kube-apiserver-external-signer",
			cryptomaterial.KubeAPIServerExternalSigner(certsDir),
			cryptomaterial.KubeAPIServerServingSignerCAValidityDays,
		).WithServingCertificates(
			&cryptomaterial.ServingCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         "kube-external-serving",
					ValidityDays: cryptomaterial.KubeAPIServerServingCertValidityDays,
				},
				Hostnames: []string{
					cfg.NodeName,
				},
			},
		),

		cryptomaterial.NewCertificateSigner(
			"kube-apiserver-localhost-signer",
			cryptomaterial.KubeAPIServerLocalhostSigner(certsDir),
			cryptomaterial.KubeAPIServerServingSignerCAValidityDays,
		).WithServingCertificates(
			&cryptomaterial.ServingCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         "kube-apiserver-localhost-serving",
					ValidityDays: cryptomaterial.KubeAPIServerServingCertValidityDays,
				},
				Hostnames: []string{
					"127.0.0.1",
					"localhost",
				},
			},
		),

		cryptomaterial.NewCertificateSigner(
			"kube-apiserver-service-network-signer",
			cryptomaterial.KubeAPIServerServiceNetworkSigner(certsDir),
			cryptomaterial.KubeAPIServerServingSignerCAValidityDays,
		).WithServingCertificates(
			&cryptomaterial.ServingCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
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
		cryptomaterial.NewCertificateSigner(
			"etcd-signer",
			cryptomaterial.EtcdSignerDir(certsDir),
			cryptomaterial.EtcdSignerCAValidityDays,
		).WithClientCertificates(
			&cryptomaterial.ClientCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         "apiserver-etcd-client",
					ValidityDays: 10 * 365,
				},
				UserInfo: &user.DefaultInfo{Name: "etcd", Groups: []string{"etcd"}},
			},
		).WithPeerCertificiates(
			&cryptomaterial.PeerCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         "etcd-peer",
					ValidityDays: 3 * 365,
				},
				UserInfo:  &user.DefaultInfo{Name: "system:etcd-peer:etcd-client", Groups: []string{"system:etcd-peers"}},
				Hostnames: []string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName},
			},
			&cryptomaterial.PeerCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
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
	certChains *cryptomaterial.CertificateChains,
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
