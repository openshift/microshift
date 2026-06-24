package components

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

const (
	metricsServerManifestPath = "/usr/lib/microshift/manifests.d/080-microshift-metrics-server"
	metricsNamespace          = "openshift-monitoring"
)

var metricsServerEventRecorder events.Recorder = events.NewLoggingEventRecorder("microshift-metrics-server", clock.RealClock{})

var metricsClientCARecorder events.Recorder = events.NewLoggingEventRecorder("metrics-client-ca", clock.RealClock{})

var metricsClientCAConsumerPaths = []string{
	"/usr/lib/microshift/manifests.d/081-microshift-kube-state-metrics",
	"/usr/lib/microshift/manifests.d/082-microshift-node-exporter",
}

// ProvisionMetricsServerCerts provisions the TLS client certificate and kubelet
// serving CA that metrics-server needs to authenticate to kubelet and verify its
// serving certificate when scraping /metrics/resource. These are provisioned at
// runtime rather than baked into manifests because the certificates are generated
// by MicroShift's certificate lifecycle and must be refreshed from the live PKI.
func ProvisionMetricsServerCerts(ctx context.Context, cfg *config.Config) error {
	exists, err := util.PathExists(metricsServerManifestPath)
	if err != nil {
		return err
	}
	if !exists {
		klog.V(2).Infof("Metrics-server manifests not found at %s, skipping cert provisioning", metricsServerManifestPath)
		return nil
	}

	kubeconfigPath := cfg.KubeConfigPath(config.KubeAdmin)

	clientset, err := getKubernetesClient(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("creating clientset: %w", err)
	}

	err = wait.PollUntilContextTimeout(ctx, 2*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		_, err := clientset.CoreV1().Namespaces().Get(ctx, metricsNamespace, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		if !apierrors.IsNotFound(err) {
			klog.Errorf("getting namespace %s: %v", metricsNamespace, err)
			return false, nil
		}
		klog.V(2).Infof("Waiting for namespace %s to be created by kustomize", metricsNamespace)
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("waiting for namespace %s: %w", metricsNamespace, err)
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

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metrics-server-client-certs",
			Namespace: metricsNamespace,
			Annotations: map[string]string{
				"openshift.io/owning-component": "metrics-server",
			},
		},
		Type: corev1.SecretTypeTLS,
		Data: map[string][]byte{
			"tls.crt": certPEM,
			"tls.key": keyPEM,
		},
	}

	err = wait.PollUntilContextTimeout(ctx, 2*time.Second, 1*time.Minute, true, func(ctx context.Context) (bool, error) {
		_, _, err := resourceapply.ApplySecret(ctx, clientset.CoreV1(), metricsServerEventRecorder, secret)
		if err != nil {
			klog.Errorf("applying metrics-server client cert secret: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("applying metrics-server client cert secret: %w", err)
	}

	caPEM, err := os.ReadFile(cryptomaterial.KubeletServingCAPath(certsDir))
	if err != nil {
		return err
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "kubelet-serving-ca-bundle",
			Namespace: metricsNamespace,
			Annotations: map[string]string{
				"openshift.io/owning-component": "metrics-server",
			},
		},
		Data: map[string]string{
			"ca-bundle.crt": string(caPEM),
		},
	}

	err = wait.PollUntilContextTimeout(ctx, 2*time.Second, 1*time.Minute, true, func(ctx context.Context) (bool, error) {
		_, _, err := resourceapply.ApplyConfigMap(ctx, clientset.CoreV1(), metricsServerEventRecorder, cm)
		if err != nil {
			klog.Errorf("applying kubelet serving CA configmap: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("applying kubelet serving CA configmap: %w", err)
	}

	klog.Infof("Provisioned metrics-server kubelet client cert and CA bundle")
	return nil
}

// ProvisionMetricsClientCA provisions the admin-kubeconfig-signer CA that
// kube-rbac-proxy sidecars in kube-state-metrics and node-exporter use to
// verify client certificates on incoming scrape requests. The CA cannot be
// included in static manifests because it is generated at MicroShift startup
// and may be rotated; this function ensures the ConfigMap reflects the current CA.
func ProvisionMetricsClientCA(ctx context.Context, cfg *config.Config) error {
	needed := false
	for _, p := range metricsClientCAConsumerPaths {
		exists, err := util.PathExists(p)
		if err != nil {
			return err
		}
		if exists {
			needed = true
			break
		}
	}
	if !needed {
		klog.V(2).Infof("No monitoring components found, skipping metrics-client-ca provisioning")
		return nil
	}

	kubeconfigPath := cfg.KubeConfigPath(config.KubeAdmin)
	clientset, err := getKubernetesClient(kubeconfigPath)
	if err != nil {
		return fmt.Errorf("creating clientset: %w", err)
	}

	err = wait.PollUntilContextTimeout(ctx, 2*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		_, err := clientset.CoreV1().Namespaces().Get(ctx, metricsNamespace, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		if !apierrors.IsNotFound(err) {
			klog.Errorf("getting namespace %s: %v", metricsNamespace, err)
			return false, nil
		}
		klog.V(2).Infof("Waiting for namespace %s to be created by kustomize", metricsNamespace)
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("waiting for namespace %s: %w", metricsNamespace, err)
	}

	certsDir := cryptomaterial.CertsDirectory(config.DataDir)
	caCertPath := cryptomaterial.CACertPath(cryptomaterial.AdminKubeconfigSignerDir(certsDir))
	caPEM, err := os.ReadFile(caCertPath)
	if err != nil {
		return fmt.Errorf("reading admin-kubeconfig-signer CA: %w", err)
	}

	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "metrics-client-ca",
			Namespace: metricsNamespace,
			Annotations: map[string]string{
				"openshift.io/owning-component": "Monitoring",
			},
			Labels: map[string]string{
				"app.kubernetes.io/managed-by": "cluster-monitoring-operator",
				"app.kubernetes.io/part-of":    "openshift-monitoring",
			},
		},
		Data: map[string]string{
			"client-ca.crt": string(caPEM),
		},
	}

	err = wait.PollUntilContextTimeout(ctx, 2*time.Second, 1*time.Minute, true, func(ctx context.Context) (bool, error) {
		_, _, err := resourceapply.ApplyConfigMap(ctx, clientset.CoreV1(), metricsClientCARecorder, cm)
		if err != nil {
			klog.Errorf("applying metrics-client-ca configmap: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("applying metrics-client-ca configmap: %w", err)
	}

	klog.Infof("Provisioned metrics-client-ca configmap in %s", metricsNamespace)
	return nil
}
