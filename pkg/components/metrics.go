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
	metricsServerNamespace    = "openshift-monitoring"
)

var metricsEventRecorder events.Recorder = events.NewLoggingEventRecorder("microshift-metrics-server", clock.RealClock{})

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
		_, err := clientset.CoreV1().Namespaces().Get(ctx, metricsServerNamespace, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		if !apierrors.IsNotFound(err) {
			klog.Errorf("getting namespace %s: %v", metricsServerNamespace, err)
			return false, nil
		}
		klog.V(2).Infof("Waiting for namespace %s to be created by kustomize", metricsServerNamespace)
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("waiting for namespace %s: %w", metricsServerNamespace, err)
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
			Namespace: metricsServerNamespace,
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
		_, _, err := resourceapply.ApplySecret(ctx, clientset.CoreV1(), metricsEventRecorder, secret)
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
			Namespace: metricsServerNamespace,
			Annotations: map[string]string{
				"openshift.io/owning-component": "metrics-server",
			},
		},
		Data: map[string]string{
			"ca-bundle.crt": string(caPEM),
		},
	}

	err = wait.PollUntilContextTimeout(ctx, 2*time.Second, 1*time.Minute, true, func(ctx context.Context) (bool, error) {
		_, _, err := resourceapply.ApplyConfigMap(ctx, clientset.CoreV1(), metricsEventRecorder, cm)
		if err != nil {
			klog.Errorf("applying kubelet serving CA configmap: %v", err)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("applying kubelet serving CA configmap: %v", err)
	}

	klog.Infof("Provisioned metrics-server kubelet client cert and CA bundle")
	return nil
}
