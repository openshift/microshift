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
	"path/filepath"

	"k8s.io/apiserver/pkg/authentication/user"
	ctrl "k8s.io/kubernetes/pkg/controlplane"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

func initAll(cfg *config.MicroshiftConfig) error {
	// create CA and keys
	clusterTrustBundlePEM, certChains, err := initCerts(cfg)
	if err != nil {
		return err
	}
	// create kubeconfig for kube-scheduler, kubelet,controller-manager
	if err := initKubeconfig(cfg, clusterTrustBundlePEM, certChains); err != nil {
		return err
	}

	return nil
}

func loadCA(cfg *config.MicroshiftConfig) error {
	return util.LoadRootCA(filepath.Join(cfg.DataDir, "/certs/ca-bundle"), "ca-bundle.crt", "ca-bundle.key")
}

func initCerts(cfg *config.MicroshiftConfig) ([]byte, *cryptomaterial.CertificateChains, error) {
	_, svcNet, err := net.ParseCIDR(cfg.Cluster.ServiceCIDR)
	if err != nil {
		return nil, nil, err
	}

	_, apiServerServiceIP, err := ctrl.ServiceIPRange(*svcNet)
	if err != nil {
		return nil, nil, err
	}

	certsDir := cryptomaterial.CertsDirectory(cfg.DataDir)
	// store root CA for all
	//TODO generate ca bundles for each component
	clusterTrustBundlePEM, _, err := util.StoreRootCA("https://kubernetes.svc", filepath.Join(certsDir, "/ca-bundle"),
		"ca-bundle.crt", "ca-bundle.key",
		[]string{"https://kubernetes.svc"})

	if err != nil {
		return nil, nil, err
	}

	// FIXME: don't add the whole root CA to client CA bundle, get rid of a general trust by splitting the root CA
	if err := cryptomaterial.AddToTotalClientCABundle(certsDir, clusterTrustBundlePEM); err != nil {
		return nil, nil, fmt.Errorf("failed to add the root CA to the total client CA bundle: %w", err)
	}

	// based on https://github.com/openshift/cluster-etcd-operator/blob/master/bindata/bootkube/bootstrap-manifests/etcd-member-pod.yaml#L19
	if err := util.GenCerts("etcd-server", filepath.Join(certsDir, "/etcd"),
		"etcd-serving.crt", "etcd-serving.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName}); err != nil {
		return nil, nil, err
	}

	if err := util.GenCerts("etcd-peer", filepath.Join(certsDir, "/etcd"),
		"etcd-peer.crt", "etcd-peer.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName}); err != nil {
		return nil, nil, err
	}

	certChains, err := cryptomaterial.NewCertificateChains(
		// kube-control-plane-signer
		cryptomaterial.NewCertificateSigner(
			"kube-control-plane-signer",
			cryptomaterial.KubeControlPlaneSignerCertDir(certsDir),
			cryptomaterial.ClientCAValidityDays,
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
			cryptomaterial.ClientCAValidityDays,
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
			cryptomaterial.ClientCAValidityDays,
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
			cryptomaterial.ClientCAValidityDays,
		).WithClientCertificates(
			&cryptomaterial.ClientCertificateSigningRequestInfo{
				CertificateSigningRequestInfo: cryptomaterial.CertificateSigningRequestInfo{
					Name:         "kubelet-client",
					ValidityDays: cryptomaterial.ClientCertValidityDays,
				},
				// userinfo per https://kubernetes.io/docs/reference/access-authn-authz/node/#overview
				UserInfo: &user.DefaultInfo{Name: "system:node:" + cfg.NodeName, Groups: []string{"system:nodes"}},
			},
		).WithSubCAs(
			cryptomaterial.NewCertificateSigner("kube-csr-signer", cryptomaterial.CSRSignerCertDir(certsDir), cryptomaterial.ClientCAValidityDays),
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
	).Complete()

	if err != nil {
		return nil, nil, err
	}

	csrSignerCAPEM, err := certChains.GetSigner("kubelet-signer", "kube-csr-signer").GetSignerCertPEM()
	if err != nil {
		return nil, nil, err
	}

	if err := cryptomaterial.AddToKubeletClientCABundle(certsDir, csrSignerCAPEM); err != nil {
		return nil, nil, err
	}

	if err := cryptomaterial.AddToTotalClientCABundle(certsDir, csrSignerCAPEM); err != nil {
		return nil, nil, err
	}

	// kube-apiserver
	if err := util.GenCerts("etcd-client", filepath.Join(cfg.DataDir, "/resources/kube-apiserver/secrets/etcd-client"),
		"tls.crt", "tls.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName}); err != nil {
		return nil, nil, err
	}
	if err := util.GenCerts("kube-apiserver", filepath.Join(cfg.DataDir, "/certs/kube-apiserver/secrets/service-network-serving-certkey"),
		"tls.crt", "tls.key",
		[]string{"kube-apiserver", cfg.NodeIP, cfg.NodeName, "127.0.0.1", "kubernetes.default.svc", "kubernetes.default", "kubernetes",
			"localhost",
			apiServerServiceIP.String()}); err != nil {
		return nil, nil, err
	}
	if err := util.GenKeys(filepath.Join(cfg.DataDir, "/resources/kube-apiserver/secrets/service-account-key"),
		"service-account.crt", "service-account.key"); err != nil {
		return nil, nil, err
	}
	if err := util.GenCerts("system:masters", filepath.Join(cfg.DataDir, "/certs/kube-apiserver/secrets/aggregator-client"),
		"tls.crt", "tls.key",
		[]string{"system:admin", "system:masters"}); err != nil {
		return nil, nil, err
	}

	if err := util.GenCerts("kubelet", filepath.Join(cfg.DataDir, "/resources/kubelet/secrets/kubelet-server"),
		"tls.crt", "tls.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName}); err != nil {
		return nil, nil, err
	}

	if err := util.GenKeys(filepath.Join(cfg.DataDir, "/resources/kube-apiserver/sa-public-key"),
		"serving-ca.pub", "serving-ca.key"); err != nil {
		return nil, nil, err
	}

	// ocp
	if err := util.GenCerts("openshift-controller-manager", filepath.Join(cfg.DataDir, "/resources/openshift-controller-manager/secrets"),
		"tls.crt", "tls.key",
		[]string{"openshift-controller-manager", cfg.NodeName, cfg.NodeIP, "127.0.0.1", "kubernetes.default.svc", "kubernetes.default",
			"kubernetes", "localhost"}); err != nil {
		return nil, nil, err
	}
	if err := util.GenCerts("service-ca", filepath.Join(cfg.DataDir, "/resources/service-ca/secrets/service-ca"),
		"tls.crt", "tls.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName, apiServerServiceIP.String()}); err != nil {
		return nil, nil, err
	}
	return clusterTrustBundlePEM, certChains, nil
}

func initKubeconfig(
	cfg *config.MicroshiftConfig,
	clusterTrustBundlePEM []byte,
	certChains *cryptomaterial.CertificateChains,
) error {

	adminKubeconfigCertPEM, adminKubeconfigKeyPEM, err := certChains.GetCertKey("admin-kubeconfig-signer", "admin-kubeconfig-client")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		filepath.Join(cfg.DataDir, "resources", "kubeadmin", "kubeconfig"),
		cfg.Cluster.URL,
		clusterTrustBundlePEM,
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
		filepath.Join(cfg.DataDir, "resources", "kube-controller-manager", "kubeconfig"),
		cfg.Cluster.URL,
		clusterTrustBundlePEM,
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
		filepath.Join(cfg.DataDir, "resources", "kube-scheduler", "kubeconfig"),
		cfg.Cluster.URL,
		clusterTrustBundlePEM,
		schedulerCertPEM, schedulerKeyPEM,
	); err != nil {
		return err
	}

	kubeletCertPEM, kubeletKeyPEM, err := certChains.GetCertKey("kubelet-signer", "kubelet-client")
	if err != nil {
		return err
	}
	if err := util.KubeConfigWithClientCerts(
		filepath.Join(cfg.DataDir, "resources", "kubelet", "kubeconfig"),
		cfg.Cluster.URL,
		clusterTrustBundlePEM,
		kubeletCertPEM, kubeletKeyPEM,
	); err != nil {
		return err
	}
	return nil
}
