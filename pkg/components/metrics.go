package components

import (
	"context"
	"fmt"
	"os"
	"time"

	"k8s.io/client-go/kubernetes"

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

func waitForNamespace(ctx context.Context, clientset kubernetes.Interface, namespace string) error {
	return wait.PollUntilContextTimeout(ctx, 2*time.Second, 5*time.Minute, true, func(ctx context.Context) (bool, error) {
		_, err := clientset.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		if !apierrors.IsNotFound(err) {
			klog.Errorf("getting namespace %s: %v", namespace, err)
			return false, nil
		}
		klog.V(2).Infof("Waiting for namespace %s to be created by kustomize", namespace)
		return false, nil
	})
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

	if err := waitForNamespace(ctx, clientset, metricsNamespace); err != nil {
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

	if err := ensureMetricsServerServingCert(ctx, clientset, certsDir); err != nil {
		return fmt.Errorf("ensuring metrics-server serving cert: %w", err)
	}

	klog.Infof("Provisioned metrics-server kubelet client cert and CA bundle")
	return nil
}

// ensureMetricsServerServingCert makes sure the metrics-server serving cert
// secret ("metrics-server-tls") exists. In normal operation the service-ca
// operator generates it from the serving-cert annotation on the metrics-server
// Service. However, if that secret is lost (e.g. after a data cleanup / restart
// storm) the operator does not always regenerate it, leaving metrics-server
// stuck ContainerCreating and the metrics.k8s.io APIService unavailable
// (USHIFT-7500). As a safety net, mint the serving cert from MicroShift's
// service-CA when the secret is missing.
//
// The secret is only created when absent so we do not fight the operator in the
// normal case: both sign with the same service-CA, so whichever cert is present
// is valid against the service-CA bundle the operator injects into the
// APIService.
func ensureMetricsServerServingCert(ctx context.Context, clientset kubernetes.Interface, certsDir string) error {
	const servingSecretName = "metrics-server-tls"

	_, err := clientset.CoreV1().Secrets(metricsNamespace).Get(ctx, servingSecretName, metav1.GetOptions{})
	if err == nil {
		klog.V(2).Infof("Secret %s/%s already exists, leaving it to the service-ca operator", metricsNamespace, servingSecretName)
		return nil
	}
	if !apierrors.IsNotFound(err) {
		return fmt.Errorf("getting secret %s/%s: %w", metricsNamespace, servingSecretName, err)
	}

	servingDir := cryptomaterial.MetricsServerServingCertDir(certsDir)
	certPEM, err := os.ReadFile(cryptomaterial.ServingCertPath(servingDir))
	if err != nil {
		return err
	}
	keyPEM, err := os.ReadFile(cryptomaterial.ServingKeyPath(servingDir))
	if err != nil {
		return err
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      servingSecretName,
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
		_, createErr := clientset.CoreV1().Secrets(metricsNamespace).Create(ctx, secret, metav1.CreateOptions{})
		if createErr != nil && !apierrors.IsAlreadyExists(createErr) {
			klog.Errorf("creating metrics-server serving cert secret: %v", createErr)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("creating metrics-server serving cert secret: %w", err)
	}

	klog.Infof("Provisioned metrics-server serving cert secret %s/%s (service-ca operator secret was missing)", metricsNamespace, servingSecretName)
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

	if err := waitForNamespace(ctx, clientset, metricsNamespace); err != nil {
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
