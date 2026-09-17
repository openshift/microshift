package components

import (
	"context"
	"fmt"
	"os"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/openshift/library-go/pkg/operator/events"
	"github.com/openshift/library-go/pkg/operator/resource/resourceapply"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilnet "k8s.io/apimachinery/pkg/util/net"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/klog/v2"
	"k8s.io/utils/clock"
)

const (
	metricsServerManifestPath    = "/usr/lib/microshift/manifests.d/080-microshift-metrics-server"
	metricsNamespace             = "openshift-monitoring"
	metricsServerServiceName     = "metrics-server"
	metricsServerTLSResourceName = "metrics-server-tls"
	serviceCANamespace           = "openshift-service-ca"
	serviceCADeploymentName      = "service-ca"

	// metricsServerServingCertRecoveryAnnotation changes whenever the serving
	// certificate needs recovery. The service-ca controller watches Service
	// updates, so this requeues certificate generation while the expected Secret
	// is absent or incomplete.
	metricsServerServingCertRecoveryAnnotation = "microshift.openshift.io/service-ca-reconcile-at"

	// The pinned service-ca controller stops trying after ten failures until these
	// annotations are cleared. Keep these names in sync with
	// service-ca-operator/pkg/controller/api.
	servingCertGenerationErrorAnnotation         = "service.beta.openshift.io/serving-cert-generation-error"
	servingCertGenerationErrorNumAnnotation      = "service.beta.openshift.io/serving-cert-generation-error-num"
	alphaServingCertGenerationErrorAnnotation    = "service.alpha.openshift.io/serving-cert-generation-error"
	alphaServingCertGenerationErrorNumAnnotation = "service.alpha.openshift.io/serving-cert-generation-error-num"
)

type metricsServerServingCertWaitOptions struct {
	timeout                time.Duration
	pollInterval           time.Duration
	controllerRetryBackoff wait.Backoff
	serviceRetryBackoff    wait.Backoff
}

var defaultMetricsServerServingCertWaitOptions = metricsServerServingCertWaitOptions{
	timeout:      5 * time.Minute,
	pollInterval: 2 * time.Second,
	// These discovery waits only read resources. Once service-ca is ready and
	// the Service exists, reconciliation retries are bounded by this backoff.
	controllerRetryBackoff: wait.Backoff{Duration: time.Second, Factor: 2, Steps: 8, Cap: 30 * time.Second},
	serviceRetryBackoff:    wait.Backoff{Duration: time.Second, Factor: 2, Steps: 8, Cap: 30 * time.Second},
}

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

func metricsServerServingCertReady(secret *corev1.Secret) bool {
	return secret != nil && len(secret.Data[corev1.TLSCertKey]) > 0 && len(secret.Data[corev1.TLSPrivateKeyKey]) > 0
}

func serviceCAControllerReady(deployment *appsv1.Deployment) bool {
	if deployment == nil {
		return false
	}
	for _, condition := range deployment.Status.Conditions {
		if condition.Type == appsv1.DeploymentAvailable && condition.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func isTransientKubernetesAPIError(err error) bool {
	return apierrors.IsInternalError(err) ||
		apierrors.IsServerTimeout(err) ||
		apierrors.IsServiceUnavailable(err) ||
		apierrors.IsTimeout(err) ||
		apierrors.IsTooManyRequests(err) ||
		apierrors.IsUnexpectedServerError(err) ||
		utilnet.IsTimeout(err) ||
		utilnet.IsConnectionRefused(err) ||
		utilnet.IsConnectionReset(err) ||
		utilnet.IsProbableEOF(err) ||
		utilnet.IsHTTP2ConnectionLost(err)
}

func waitForServiceCAController(ctx context.Context, clientset kubernetes.Interface, backoff wait.Backoff) error {
	return wait.ExponentialBackoffWithContext(ctx, backoff, func(ctx context.Context) (bool, error) {
		deployment, err := clientset.AppsV1().Deployments(serviceCANamespace).Get(ctx, serviceCADeploymentName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if isTransientKubernetesAPIError(err) {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("getting service-ca controller: %w", err)
		}
		return serviceCAControllerReady(deployment), nil
	})
}

func waitForMetricsServerService(ctx context.Context, clientset kubernetes.Interface, backoff wait.Backoff) error {
	return wait.ExponentialBackoffWithContext(ctx, backoff, func(ctx context.Context) (bool, error) {
		_, err := clientset.CoreV1().Services(metricsNamespace).Get(ctx, metricsServerServiceName, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		if isTransientKubernetesAPIError(err) {
			return false, nil
		}
		if err != nil {
			return false, fmt.Errorf("getting metrics-server Service: %w", err)
		}
		return true, nil
	})
}

func triggerMetricsServerServingCertReconciliation(ctx context.Context, clientset kubernetes.Interface, backoff wait.Backoff) error {
	return wait.ExponentialBackoffWithContext(ctx, backoff, func(ctx context.Context) (bool, error) {
		service, err := clientset.CoreV1().Services(metricsNamespace).Get(ctx, metricsServerServiceName, metav1.GetOptions{})
		if err != nil {
			if isTransientKubernetesAPIError(err) {
				return false, nil
			}
			return false, err
		}

		annotations := service.GetAnnotations()
		if annotations == nil {
			annotations = map[string]string{}
		}
		annotations[metricsServerServingCertRecoveryAnnotation] = time.Now().UTC().Format(time.RFC3339Nano)
		delete(annotations, servingCertGenerationErrorAnnotation)
		delete(annotations, servingCertGenerationErrorNumAnnotation)
		delete(annotations, alphaServingCertGenerationErrorAnnotation)
		delete(annotations, alphaServingCertGenerationErrorNumAnnotation)
		service.SetAnnotations(annotations)

		_, err = clientset.CoreV1().Services(metricsNamespace).Update(ctx, service, metav1.UpdateOptions{})
		if err == nil {
			return true, nil
		}
		if apierrors.IsConflict(err) || isTransientKubernetesAPIError(err) {
			return false, nil
		}
		return false, err
	})
}

// waitForMetricsServerServingCert waits for service-ca to restore the serving
// certificate used by metrics-server. It waits for the controller and Service
// with bounded exponential backoff, then makes one Service update that both
// clears service-ca's retry ceiling and requeues certificate generation.
func waitForMetricsServerServingCert(ctx context.Context, clientset kubernetes.Interface) error {
	return waitForMetricsServerServingCertWithOptions(ctx, clientset, defaultMetricsServerServingCertWaitOptions)
}

func waitForMetricsServerServingCertWithOptions(ctx context.Context, clientset kubernetes.Interface, options metricsServerServingCertWaitOptions) error {
	secret, err := clientset.CoreV1().Secrets(metricsNamespace).Get(ctx, metricsServerTLSResourceName, metav1.GetOptions{})
	if err == nil && metricsServerServingCertReady(secret) {
		return nil
	}
	if err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("getting metrics-server serving cert secret: %w", err)
	}

	waitCtx, cancel := context.WithTimeout(ctx, options.timeout)
	defer cancel()

	if err := waitForServiceCAController(waitCtx, clientset, options.controllerRetryBackoff); err != nil {
		return fmt.Errorf("waiting for service-ca controller: %w", err)
	}
	if err := waitForMetricsServerService(waitCtx, clientset, options.serviceRetryBackoff); err != nil {
		return fmt.Errorf("waiting for metrics-server Service: %w", err)
	}
	if err := triggerMetricsServerServingCertReconciliation(waitCtx, clientset, options.serviceRetryBackoff); err != nil {
		return fmt.Errorf("triggering metrics-server serving cert reconciliation: %w", err)
	}

	return wait.PollUntilContextTimeout(waitCtx, options.pollInterval, options.timeout, true, func(ctx context.Context) (bool, error) {
		secret, err := clientset.CoreV1().Secrets(metricsNamespace).Get(ctx, metricsServerTLSResourceName, metav1.GetOptions{})
		if err == nil && metricsServerServingCertReady(secret) {
			return true, nil
		}
		if err != nil && !apierrors.IsNotFound(err) {
			return false, fmt.Errorf("getting metrics-server serving cert secret: %w", err)
		}

		klog.V(2).Info("Waiting for service-ca to restore the metrics-server serving certificate")
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
	if err := waitForMetricsServerServingCert(ctx, clientset); err != nil {
		return fmt.Errorf("waiting for metrics-server serving cert: %w", err)
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
