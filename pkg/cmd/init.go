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

	"github.com/openshift/library-go/pkg/crypto"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
)

type tlsConfigs struct {
	clusterTrustBundle []byte

	kubeControllerManagerClient *crypto.TLSCertificateConfig
	kubeSchedulerClient         *crypto.TLSCertificateConfig

	adminKubeconfigClient *crypto.TLSCertificateConfig
}

func initAll(cfg *config.MicroshiftConfig) error {
	// create CA and keys
	certConfig, err := initCerts(cfg)
	if err != nil {
		return err
	}
	// create kubeconfig for kube-scheduler, kubelet,controller-manager
	if err := initKubeconfig(cfg, certConfig); err != nil {
		return err
	}

	return nil
}

func loadCA(cfg *config.MicroshiftConfig) error {
	return util.LoadRootCA(filepath.Join(cfg.DataDir, "/certs/ca-bundle"), "ca-bundle.crt", "ca-bundle.key")
}

func initCerts(cfg *config.MicroshiftConfig) (*tlsConfigs, error) {
	_, svcNet, err := net.ParseCIDR(cfg.Cluster.ServiceCIDR)
	if err != nil {
		return nil, err
	}

	_, apiServerServiceIP, err := ctrl.ServiceIPRange(*svcNet)
	if err != nil {
		return nil, err
	}

	certConfigs := &tlsConfigs{}

	certsDir := cryptomaterial.CertsDirectory(cfg.DataDir)
	// store root CA for all
	//TODO generate ca bundles for each component
	certConfigs.clusterTrustBundle, _, err = util.StoreRootCA("https://kubernetes.svc", filepath.Join(certsDir, "/ca-bundle"),
		"ca-bundle.crt", "ca-bundle.key",
		[]string{"https://kubernetes.svc"})

	if err != nil {
		return nil, err
	}

	// FIXME: don't add the whole root CA to client CA bundle, get rid of a general trust by splitting the root CA
	if err := cryptomaterial.AddToTotalClientCABundle(certsDir, certConfigs.clusterTrustBundle); err != nil {
		return nil, fmt.Errorf("failed to add the root CA to the total client CA bundle: %w", err)
	}
	if err := cryptomaterial.AddToKubeletClientCABundle(certsDir, certConfigs.clusterTrustBundle); err != nil {
		return nil, fmt.Errorf("failed to add the root CA to the kubelet client CA bundle: %w", err)
	}

	// based on https://github.com/openshift/cluster-etcd-operator/blob/master/bindata/bootkube/bootstrap-manifests/etcd-member-pod.yaml#L19
	if err := util.GenCerts("etcd-server", filepath.Join(certsDir, "/etcd"),
		"etcd-serving.crt", "etcd-serving.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName}); err != nil {
		return nil, err
	}

	if err := util.GenCerts("etcd-peer", filepath.Join(certsDir, "/etcd"),
		"etcd-peer.crt", "etcd-peer.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName}); err != nil {
		return nil, err
	}

	// kube-control-plane-signer
	controlPlaneSignerCA, controlPlaneSignerCAPEM, err := generateClientCA(
		certsDir,
		cryptomaterial.KubeControlPlaneSignerCertDir(certsDir),
		"kube-control-plane-signer")
	if err != nil {
		return nil, err
	}

	if err := cryptomaterial.AddToKubeletClientCABundle(certsDir, controlPlaneSignerCAPEM); err != nil {
		return nil, fmt.Errorf("failed to add %s to kubelet client CA bundle: %w", "kube-control-plane-signer", err)
	}

	kcmClientDir := cryptomaterial.KubeControllerManagerClientCertDir(certsDir)
	certConfigs.kubeControllerManagerClient, _, err = controlPlaneSignerCA.EnsureClientCertificate(
		cryptomaterial.ClientCertPath(kcmClientDir),
		cryptomaterial.ClientKeyPath(kcmClientDir),
		&user.DefaultInfo{Name: "system:kube-controller-manager"},
		cryptomaterial.ClientCertValidityDays,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate kube-controller-manager client certificate: %w", err)
	}

	schedulerClientDir := cryptomaterial.KubeSchedulerClientCertDir(certsDir)
	certConfigs.kubeSchedulerClient, _, err = controlPlaneSignerCA.EnsureClientCertificate(
		cryptomaterial.ClientCertPath(schedulerClientDir),
		cryptomaterial.ClientKeyPath(schedulerClientDir),
		&user.DefaultInfo{Name: "system:kube-scheduler"},
		cryptomaterial.ClientCertValidityDays,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate kube-scheduler client certificate: %w", err)
	}

	// kube-apiserver
	if err := util.GenCerts("etcd-client", filepath.Join(cfg.DataDir, "/resources/kube-apiserver/secrets/etcd-client"),
		"tls.crt", "tls.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName}); err != nil {
		return nil, err
	}
	if err := util.GenCerts("kube-apiserver", filepath.Join(cfg.DataDir, "/certs/kube-apiserver/secrets/service-network-serving-certkey"),
		"tls.crt", "tls.key",
		[]string{"kube-apiserver", cfg.NodeIP, cfg.NodeName, "127.0.0.1", "kubernetes.default.svc", "kubernetes.default", "kubernetes",
			"localhost",
			apiServerServiceIP.String()}); err != nil {
		return nil, err
	}
	if err := util.GenKeys(filepath.Join(cfg.DataDir, "/resources/kube-apiserver/secrets/service-account-key"),
		"service-account.crt", "service-account.key"); err != nil {
		return nil, err
	}
	if err := util.GenCerts("system:masters", filepath.Join(cfg.DataDir, "/certs/kube-apiserver/secrets/aggregator-client"),
		"tls.crt", "tls.key",
		[]string{"system:admin", "system:masters"}); err != nil {
		return nil, err
	}

	// kube-apiserver-to-kubelet-signer
	kubeAPIServerToKubeletSignerCA, kubeAPIServerToKubeletSignerCAPEM, err := generateClientCA(
		certsDir,
		cryptomaterial.KubeAPIServerToKubeletSignerCertDir(certsDir),
		"kube-apiserver-to-kubelet-signer")
	if err != nil {
		return nil, err
	}

	if err := cryptomaterial.AddToKubeletClientCABundle(certsDir, kubeAPIServerToKubeletSignerCAPEM); err != nil {
		return nil, fmt.Errorf("failed to add %s to kubelet client CA bundle: %w", "kube-apiserver-to-kubelet-signer", err)
	}

	kubeAPIServerToKubeletClientDir := cryptomaterial.KubeAPIServerToKubeletClientCertDir(certsDir)
	_, _, err = kubeAPIServerToKubeletSignerCA.EnsureClientCertificate(
		cryptomaterial.ClientCertPath(kubeAPIServerToKubeletClientDir),
		cryptomaterial.ClientKeyPath(kubeAPIServerToKubeletClientDir),
		&user.DefaultInfo{Name: "system:kube-apiserver", Groups: []string{"kube-master"}},
		cryptomaterial.ClientCertValidityDays,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate kube-apiserver-to-kubelet client certificate: %w", err)
	}

	// admin-kubeconfig-signer
	adminKubeconfigSigner, adminKubeconfigSignerPEM, err := generateClientCA(
		certsDir,
		cryptomaterial.AdminKubeconfigSignerDir(certsDir),
		"admin-kubeconfig-signer",
	)
	if err != nil {
		return nil, err
	}

	if err := cryptomaterial.AddToKubeletClientCABundle(certsDir, adminKubeconfigSignerPEM); err != nil {
		return nil, fmt.Errorf("failed to add %s to kubelet client CA bundle: %w", "admin-kubeconfig-signer", err)
	}
	adminKubeconfigClientDir := cryptomaterial.AdminKubeconfigClientCertDir(certsDir)
	certConfigs.adminKubeconfigClient, _, err = adminKubeconfigSigner.EnsureClientCertificate(
		cryptomaterial.ClientCertPath(adminKubeconfigClientDir),
		cryptomaterial.ClientKeyPath(adminKubeconfigClientDir),
		&user.DefaultInfo{Name: "system:admin", Groups: []string{"system:masters"}},
		cryptomaterial.AdminKubeconfigClientCertValidityDays,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate 'system:admin' client certificate: %w", err)
	}

	if err := util.GenKeys(filepath.Join(cfg.DataDir, "/resources/kube-apiserver/sa-public-key"),
		"serving-ca.pub", "serving-ca.key"); err != nil {
		return nil, err
	}

	if err := util.GenCerts("kubelet", filepath.Join(cfg.DataDir, "/resources/kubelet/secrets/kubelet-client"),
		"tls.crt", "tls.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName}); err != nil {
		return nil, err
	}

	// ocp
	if err := util.GenCerts("openshift-controller-manager", filepath.Join(cfg.DataDir, "/resources/openshift-controller-manager/secrets"),
		"tls.crt", "tls.key",
		[]string{"openshift-controller-manager", cfg.NodeName, cfg.NodeIP, "127.0.0.1", "kubernetes.default.svc", "kubernetes.default",
			"kubernetes", "localhost"}); err != nil {
		return nil, err
	}
	if err := util.GenCerts("service-ca", filepath.Join(cfg.DataDir, "/resources/service-ca/secrets/service-ca"),
		"tls.crt", "tls.key",
		[]string{"localhost", cfg.NodeIP, "127.0.0.1", cfg.NodeName, apiServerServiceIP.String()}); err != nil {
		return nil, err
	}
	return certConfigs, nil
}

func initKubeconfig(
	cfg *config.MicroshiftConfig,
	certConfigs *tlsConfigs,
) error {
	clusterTrustBundlePEM := certConfigs.clusterTrustBundle

	adminKubeconfigCertPEM, adminKubeconfigKeyPEM, err := certConfigs.adminKubeconfigClient.GetPEMBytes()
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

	kcmCertPEM, kcmKeyPEM, err := certConfigs.kubeControllerManagerClient.GetPEMBytes()
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

	schedulerCertPEM, schedulerKeyPEM, err := certConfigs.kubeSchedulerClient.GetPEMBytes()
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

	// per https://kubernetes.io/docs/reference/access-authn-authz/node/#overview
	if err := util.Kubeconfig(filepath.Join(cfg.DataDir, "/resources/kubelet/kubeconfig"),
		clusterTrustBundlePEM,
		"system:node:"+cfg.NodeName, []string{"system:nodes"}, cfg.Cluster.URL); err != nil {
		return err
	}
	return nil
}

func generateClientCA(certsDir, signerDir, signerName string) (*crypto.CA, []byte, error) {
	signerCA, _, err := crypto.EnsureCA(
		cryptomaterial.CACertPath(signerDir),
		cryptomaterial.CAKeyPath(signerDir),
		cryptomaterial.CASerialsPath(signerDir),
		signerName,
		cryptomaterial.ClientCAValidityDays,
	)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate %s CA certificate: %w", signerName, err)
	}

	signerCAPEM, _, err := signerCA.Config.GetPEMBytes()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve %s CA PEM: %w", signerName, err)
	}

	if err := cryptomaterial.AddToTotalClientCABundle(certsDir, signerCAPEM); err != nil {
		return nil, nil, fmt.Errorf("failed to add %s to trusted client CA bundle: %w", signerName, err)
	}

	return signerCA, signerCAPEM, nil
}
