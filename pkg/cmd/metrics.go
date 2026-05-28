package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

	kubeconfigPath := cfg.KubeConfigPath(config.KubeAdmin)

	restCfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return fmt.Errorf("building kubeconfig: %w", err)
	}
	clientset, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("creating clientset: %w", err)
	}
	const ns = "openshift-monitoring"
	err = wait.PollUntilContextTimeout(ctx, 2*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		_, err := clientset.CoreV1().Namespaces().Get(ctx, ns, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		klog.V(2).Infof("Waiting for namespace %s to be created by kustomize", ns)
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("waiting for namespace %s: %w", ns, err)
	}

	certsDir := cryptomaterial.CertsDirectory(config.DataDir)

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
