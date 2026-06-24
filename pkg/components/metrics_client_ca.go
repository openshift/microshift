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

var metricsClientCARecorder events.Recorder = events.NewLoggingEventRecorder("metrics-client-ca", clock.RealClock{})

var metricsClientCAConsumerPaths = []string{
	"/usr/lib/microshift/manifests.d/081-microshift-kube-state-metrics",
	"/usr/lib/microshift/manifests.d/082-microshift-node-exporter",
}

const metricsClientCANamespace = "openshift-monitoring"

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
		_, err := clientset.CoreV1().Namespaces().Get(ctx, metricsClientCANamespace, metav1.GetOptions{})
		if err == nil {
			return true, nil
		}
		if !apierrors.IsNotFound(err) {
			klog.Errorf("getting namespace %s: %v", metricsClientCANamespace, err)
			return false, nil
		}
		klog.V(2).Infof("Waiting for namespace %s to be created by kustomize", metricsClientCANamespace)
		return false, nil
	})
	if err != nil {
		return fmt.Errorf("waiting for namespace %s: %w", metricsClientCANamespace, err)
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
			Namespace: metricsClientCANamespace,
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

	klog.Infof("Provisioned metrics-client-ca configmap in %s", metricsClientCANamespace)
	return nil
}
