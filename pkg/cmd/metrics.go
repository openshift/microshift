package cmd

import (
	"context"
	"os"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	"k8s.io/klog/v2"
)

const metricsServerManifestPath = "/usr/lib/microshift/manifests.d/080-microshift-metrics-server"

func provisionMetricsServerCerts(ctx context.Context, cfg *config.Config) error {
	exists, err := util.PathExists(metricsServerManifestPath)
	if err != nil {
		return err
	}
	if !exists {
		klog.V(2).Infof("Metrics-server manifests not found at %s, skipping cert provisioning", metricsServerManifestPath)
		return nil
	}

	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	kubeconfigPath := cfg.KubeConfigPath(config.KubeAdmin)

	certDir := cryptomaterial.MetricsServerKubeletClientCertDir(certsDir)
	certPEM, err := os.ReadFile(cryptomaterial.ClientCertPath(certDir))
	if err != nil {
		return err
	}
	keyPEM, err := os.ReadFile(cryptomaterial.ClientKeyPath(certDir))
	if err != nil {
		return err
	}

	secretData := map[string][]byte{
		"tls.crt": certPEM,
		"tls.key": keyPEM,
	}
	if err := assets.ApplySecretWithData(ctx, "components/metrics-server/kubelet-client-secret.yaml", secretData, kubeconfigPath); err != nil {
		return err
	}

	caPEM, err := os.ReadFile(cryptomaterial.KubeletClientCAPath(certsDir))
	if err != nil {
		return err
	}

	cmData := map[string]string{
		"ca-bundle.crt": string(caPEM),
	}
	if err := assets.ApplyConfigMapWithData(ctx, "components/metrics-server/kubelet-ca-configmap.yaml", cmData, kubeconfigPath); err != nil {
		return err
	}

	klog.Infof("Provisioned metrics-server kubelet client cert and CA bundle")
	return nil
}
