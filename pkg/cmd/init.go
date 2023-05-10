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
	"net/url"
	"os"
	"path/filepath"
	"time"

	"k8s.io/apiserver/pkg/authentication/serviceaccount"
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
	regenCerts, err := certsToRegenerate(certChains)
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

	externalCertNames := []string{
		cfg.NodeName,
		"api." + cfg.BaseDomain,
	}
	externalCertNames = append(externalCertNames, cfg.SubjectAltNames...)
	// When Kube apiserver advertise address matches the node IP we can not add
	// it to the certificates or else the internal pod access to apiserver is
	// broken. Because of client-go not using SNI and the way apiserver handles
	// which certificate to serve which destination IP, internal pods start
	// getting the external certificate, which is signed by a different CA and
	// does not match the hostname.
	if cfg.KASAdvertiseAddress != cfg.NodeIP {
		externalCertNames = append(externalCertNames, cfg.NodeIP)
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
			cryptomaterial.ShortLivedCertificateValidityDays,
		).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "kube-controller-manager",
					ValidityDays: cryptomaterial.ShortLivedCertificateValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-controller-manager"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "kube-scheduler",
					ValidityDays: cryptomaterial.ShortLivedCertificateValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-scheduler"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "cluster-policy-controller",
					ValidityDays: cryptomaterial.ShortLivedCertificateValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-controller-manager"},
			},
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "route-controller-manager",
					ValidityDays: cryptomaterial.ShortLivedCertificateValidityDays,
				},
				UserInfo: serviceaccount.UserInfo("openshift-route-controller-manager", "route-controller-manager-sa", ""),
			}),

		// kube-apiserver-to-kubelet-signer
		certchains.NewCertificateSigner(
			"kube-apiserver-to-kubelet-signer",
			cryptomaterial.KubeAPIServerToKubeletSignerCertDir(certsDir),
			cryptomaterial.ShortLivedCertificateValidityDays,
		).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "kube-apiserver-to-kubelet-client",
					ValidityDays: cryptomaterial.ShortLivedCertificateValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:kube-apiserver", Groups: []string{"kube-master"}},
			}),

		// admin-kubeconfig-signer
		certchains.NewCertificateSigner(
			"admin-kubeconfig-signer",
			cryptomaterial.AdminKubeconfigSignerDir(certsDir),
			cryptomaterial.LongLivedCertificateValidityDays,
		).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "admin-kubeconfig-client",
					ValidityDays: cryptomaterial.LongLivedCertificateValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "system:admin", Groups: []string{"system:masters"}},
			}),

		// kubelet + CSR signing chain
		certchains.NewCertificateSigner(
			"kubelet-signer",
			cryptomaterial.KubeletCSRSignerSignerCertDir(certsDir),
			cryptomaterial.ShortLivedCertificateValidityDays,
		).WithSubCAs(
			certchains.NewCertificateSigner(
				"kube-csr-signer",
				cryptomaterial.CSRSignerCertDir(certsDir),
				cryptomaterial.ShortLivedCertificateValidityDays,
			).WithClientCertificates(
				&certchains.ClientCertificateSigningRequestInfo{
					CSRMeta: certchains.CSRMeta{
						Name:         "kubelet-client",
						ValidityDays: cryptomaterial.ShortLivedCertificateValidityDays,
					},
					// userinfo per https://kubernetes.io/docs/reference/access-authn-authz/node/#overview
					UserInfo: &user.DefaultInfo{Name: "system:node:" + cfg.NodeName, Groups: []string{"system:nodes"}},
				},
			).WithServingCertificates(
				&certchains.ServingCertificateSigningRequestInfo{
					CSRMeta: certchains.CSRMeta{
						Name:         "kubelet-server",
						ValidityDays: cryptomaterial.ShortLivedCertificateValidityDays,
					},
					Hostnames: []string{cfg.NodeName, cfg.NodeIP},
				},
			),
		),
		certchains.NewCertificateSigner(
			"aggregator-signer",
			cryptomaterial.AggregatorSignerDir(certsDir),
			cryptomaterial.ShortLivedCertificateValidityDays,
		).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "aggregator-client",
					ValidityDays: cryptomaterial.ShortLivedCertificateValidityDays,
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
			cryptomaterial.LongLivedCertificateValidityDays,
		).WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "route-controller-manager-serving",
					ValidityDays: cryptomaterial.ShortLivedCertificateValidityDays,
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
			cryptomaterial.LongLivedCertificateValidityDays,
		).WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "router-default-serving",
					ValidityDays: cryptomaterial.ShortLivedCertificateValidityDays,
				},
				Hostnames: []string{
					"*.apps." + cfg.BaseDomain, // wildcard for any additional auto-generated domains
				},
			},
		),

		// this signer replaces the loadbalancer signers of OCP, we don't need those
		// in Microshift
		certchains.NewCertificateSigner(
			"kube-apiserver-external-signer",
			cryptomaterial.KubeAPIServerExternalSigner(certsDir),
			cryptomaterial.LongLivedCertificateValidityDays,
		).WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "kube-external-serving",
					ValidityDays: cryptomaterial.ShortLivedCertificateValidityDays,
				},
				Hostnames: externalCertNames,
			},
		),

		certchains.NewCertificateSigner(
			"kube-apiserver-localhost-signer",
			cryptomaterial.KubeAPIServerLocalhostSigner(certsDir),
			cryptomaterial.LongLivedCertificateValidityDays,
		).WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "kube-apiserver-localhost-serving",
					ValidityDays: cryptomaterial.ShortLivedCertificateValidityDays,
				},
				Hostnames: []string{
					"localhost",
				},
			},
		),

		certchains.NewCertificateSigner(
			"kube-apiserver-service-network-signer",
			cryptomaterial.KubeAPIServerServiceNetworkSigner(certsDir),
			cryptomaterial.LongLivedCertificateValidityDays,
		).WithServingCertificates(
			&certchains.ServingCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "kube-apiserver-service-network-serving",
					ValidityDays: cryptomaterial.ShortLivedCertificateValidityDays,
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
					"api." + cfg.BaseDomain,
					"api-int." + cfg.BaseDomain,
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
			cryptomaterial.LongLivedCertificateValidityDays,
		).WithClientCertificates(
			&certchains.ClientCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "apiserver-etcd-client",
					ValidityDays: cryptomaterial.LongLivedCertificateValidityDays,
				},
				UserInfo: &user.DefaultInfo{Name: "etcd", Groups: []string{"etcd"}},
			},
		).WithPeerCertificiates(
			&certchains.PeerCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "etcd-peer",
					ValidityDays: cryptomaterial.LongLivedCertificateValidityDays,
				},
				UserInfo:  &user.DefaultInfo{Name: "system:etcd-peer:etcd-client", Groups: []string{"system:etcd-peers"}},
				Hostnames: []string{"localhost", cfg.NodeName},
			},
			&certchains.PeerCertificateSigningRequestInfo{
				CSRMeta: certchains.CSRMeta{
					Name:         "etcd-serving",
					ValidityDays: cryptomaterial.LongLivedCertificateValidityDays,
				},
				UserInfo:  &user.DefaultInfo{Name: "system:etcd-server:etcd-client", Groups: []string{"system:etcd-servers"}},
				Hostnames: []string{"localhost", cfg.NodeName},
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
		[]string{"kube-apiserver-localhost-signer"},
		[]string{"kube-apiserver-service-network-signer"},
	).Complete()

	if err != nil {
		return nil, err
	}

	saKeyDir := filepath.Join(microshiftDataDir, "/resources/kube-apiserver/secrets/service-account-key")
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
	cfg *config.MicroshiftConfig,
	certChains *certchains.CertificateChains,
) error {
	externalTrustPEM, err := os.ReadFile(cryptomaterial.CACertPath(cryptomaterial.KubeAPIServerExternalSigner(cryptomaterial.CertsDirectory(microshiftDataDir))))
	if err != nil {
		return fmt.Errorf("failed to load the external trust signer: %v", err)
	}
	internalTrustPEM, err := os.ReadFile(cryptomaterial.CACertPath(cryptomaterial.KubeAPIServerLocalhostSigner(cryptomaterial.CertsDirectory(microshiftDataDir))))
	if err != nil {
		return fmt.Errorf("failed to load the internal trust signer: %v", err)
	}

	adminKubeconfigCertPEM, adminKubeconfigKeyPEM, err := certChains.GetCertKey("admin-kubeconfig-signer", "admin-kubeconfig-client")
	if err != nil {
		return err
	}

	u, err := url.Parse(cfg.Cluster.URL)
	if err != nil {
		return fmt.Errorf("failed to parse cluster URL: %v", err)
	}
	apiServerPort, err := cfg.Cluster.ApiServerPort()
	if err != nil {
		return fmt.Errorf("failed to get apiserver port: %v", err)
	}

	// Generate one kubeconfigs per name
	for _, name := range append(cfg.SubjectAltNames, cfg.NodeName) {
		u.Host = fmt.Sprintf("%s:%d", name, apiServerPort)
		if err := util.KubeConfigWithClientCerts(
			cfg.KubeConfigAdminPath(name),
			u.String(),
			externalTrustPEM,
			adminKubeconfigCertPEM,
			adminKubeconfigKeyPEM,
		); err != nil {
			return err
		}
	}

	if err := util.KubeConfigWithClientCerts(
		cfg.KubeConfigPath(config.KubeAdmin),
		cfg.Cluster.URL,
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
		cfg.KubeConfigPath(config.KubeControllerManager),
		cfg.Cluster.URL,
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
		cfg.KubeConfigPath(config.KubeScheduler),
		cfg.Cluster.URL,
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
		cfg.KubeConfigPath(config.Kubelet),
		cfg.Cluster.URL,
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
		cfg.KubeConfigPath(config.ClusterPolicyController),
		cfg.Cluster.URL,
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
		cfg.KubeConfigPath(config.RouteControllerManager),
		cfg.Cluster.URL,
		internalTrustPEM,
		routeControllerManagerCertPEM, routeControllerManagerKeyPEM,
	); err != nil {
		return err
	}
	return nil
}

// certsToRegenerate returns paths to certificates in the given certificate chains
// bundle that need to be regenerated
func certsToRegenerate(cs *certchains.CertificateChains) ([][]string, error) {
	regenCerts := [][]string{}
	err := cs.WalkChains(nil, func(certPath []string, c x509.Certificate) error {
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
			return nil
		}

		// long lived certs
		if timeLeft < 18*month {
			regenCerts = append(regenCerts, certPath)
		}

		return nil
	})

	return regenCerts, err
}
