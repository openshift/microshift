package components

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util/cryptomaterial"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const certificateGenerationAnnotation = "microshift.openshift.io/certificate-generation"

// Reconcile from active public certificates on every start, including after an
// interrupted rollout. Pending renewal must never change running pod templates.
func certificatePodAnnotations(ctx context.Context, kubeconfigPath, namespace string, extraCertificates ...string) (map[string]string, error) {
	caPath := cryptomaterial.ServiceAccountTokenCABundlePath(cryptomaterial.CertsDirectory(config.DataDir))
	caBundle, err := os.ReadFile(caPath)
	if err != nil {
		return nil, fmt.Errorf("reading component API CA bundle: %w", err)
	}
	generation, err := certificateGeneration(caBundle, extraCertificates...)
	if err != nil {
		return nil, err
	}
	client, err := getKubernetesClient(kubeconfigPath)
	if err != nil {
		return nil, err
	}
	// The root CA publisher is asynchronous. A new pod must not start with an
	// old projected CA that its API client could cache for the process lifetime.
	if err := waitForComponentCABundle(ctx, client, namespace, string(caBundle)); err != nil {
		return nil, err
	}
	return map[string]string{certificateGenerationAnnotation: generation}, nil
}

func certificateGeneration(caBundle []byte, extraCertificates ...string) (string, error) {
	hash := sha256.New()
	_, _ = hash.Write(caBundle)
	for _, path := range extraCertificates {
		certificate, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("reading component certificate %s: %w", path, err)
		}
		_, _ = hash.Write(certificate)
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func waitForComponentCABundle(ctx context.Context, client kubernetes.Interface, namespace, caBundle string) error {
	err := wait.PollUntilContextTimeout(ctx, time.Second, time.Minute, true, func(ctx context.Context) (bool, error) {
		cm, err := client.CoreV1().ConfigMaps(namespace).Get(ctx, "kube-root-ca.crt", metav1.GetOptions{})
		if apierrors.IsNotFound(err) || apierrors.IsServiceUnavailable(err) || apierrors.IsTimeout(err) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		return cm.Data["ca.crt"] == caBundle, nil
	})
	if err != nil {
		return fmt.Errorf("waiting for current API CA bundle in namespace %s before certificate rollout: %w", namespace, err)
	}
	return nil
}

func applyDeploymentsWithCertificates(ctx context.Context, apps []string, params assets.RenderParams, kubeconfigPath, namespace string, extraCertificates ...string) error {
	annotations, err := certificatePodAnnotations(ctx, kubeconfigPath, namespace, extraCertificates...)
	if err != nil {
		return err
	}
	return assets.ApplyDeployments(ctx, apps, renderTemplate, params, kubeconfigPath, annotations)
}

func applyDaemonSetsWithCertificates(ctx context.Context, apps []string, params assets.RenderParams, kubeconfigPath, namespace string, extraCertificates ...string) error {
	annotations, err := certificatePodAnnotations(ctx, kubeconfigPath, namespace, extraCertificates...)
	if err != nil {
		return err
	}
	return assets.ApplyDaemonSets(ctx, apps, renderTemplate, params, kubeconfigPath, annotations)
}

func ovnCertificateRolloutRequired(ctx context.Context, client kubernetes.Interface, generation string, multiNode bool) (bool, error) {
	names := []string{"ovnkube-master"}
	if multiNode {
		names = append(names, "ovnkube-node")
	}
	for _, name := range names {
		ds, err := client.AppsV1().DaemonSets("openshift-ovn-kubernetes").Get(ctx, name, metav1.GetOptions{})
		if apierrors.IsNotFound(err) {
			continue // Initial bootstrap has no old credentials to replace.
		}
		if err != nil {
			return false, err
		}
		if ds.Spec.Template.Annotations[certificateGenerationAnnotation] != generation {
			return true, nil
		}
		// A previous startup may have updated the template but exited before
		// replacing all pods. Check credentials, not general pod readiness.
		selector, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
		if err != nil {
			return false, err
		}
		pods, err := client.CoreV1().Pods(ds.Namespace).List(ctx, metav1.ListOptions{LabelSelector: selector.String()})
		if err != nil {
			return false, err
		}
		for _, pod := range pods.Items {
			if metav1.IsControlledBy(&pod, ds) && pod.Annotations[certificateGenerationAnnotation] != generation {
				return true, nil
			}
		}
	}
	return false, nil
}

func waitForOVNCertificateRollout(ctx context.Context, client kubernetes.Interface) error {
	klog.Info("Waiting for OVN pod rollout before starting API-dependent components")
	err := wait.PollUntilContextTimeout(ctx, time.Second, 2*time.Minute, true, func(ctx context.Context) (bool, error) {
		for _, name := range []string{"ovnkube-master", "ovnkube-node"} {
			ds, err := client.AppsV1().DaemonSets("openshift-ovn-kubernetes").Get(ctx, name, metav1.GetOptions{})
			if apierrors.IsNotFound(err) || apierrors.IsServiceUnavailable(err) || apierrors.IsTimeout(err) {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			if !daemonSetRolledOut(ds) {
				return false, nil
			}
		}
		return true, nil
	})
	if err != nil {
		return fmt.Errorf("waiting for OVN certificate rollout: %w", err)
	}
	return nil
}

func daemonSetRolledOut(ds *appsv1.DaemonSet) bool {
	return ds.Status.ObservedGeneration >= ds.Generation &&
		ds.Status.DesiredNumberScheduled > 0 &&
		ds.Status.UpdatedNumberScheduled == ds.Status.DesiredNumberScheduled &&
		ds.Status.NumberAvailable == ds.Status.DesiredNumberScheduled
}
