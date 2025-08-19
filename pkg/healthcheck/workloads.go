package healthcheck

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	appsclientv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

type NamespaceWorkloads struct {
	Deployments  []string `json:"deployments"`
	DaemonSets   []string `json:"daemonsets"`
	StatefulSets []string `json:"statefulsets"`
}

func (nw NamespaceWorkloads) String() string {
	var parts []string

	if len(nw.Deployments) > 0 {
		parts = append(parts, fmt.Sprintf("Deployments: [%s]", strings.Join(nw.Deployments, ", ")))
	}
	if len(nw.DaemonSets) > 0 {
		parts = append(parts, fmt.Sprintf("DaemonSets: [%s]", strings.Join(nw.DaemonSets, ", ")))
	}
	if len(nw.StatefulSets) > 0 {
		parts = append(parts, fmt.Sprintf("StatefulSets: [%s]", strings.Join(nw.StatefulSets, ", ")))
	}

	if len(parts) == 0 {
		return ""
	}

	return strings.Join(parts, ", ")
}

func getKubeconfigPath() string {
	if os.Geteuid() == 0 {
		return filepath.Join(config.DataDir, "resources", string(config.KubeAdmin), "kubeconfig")
	}

	getKubeconfigFromEnv := func() string {
		kubeconfigPath, ok := os.LookupEnv("KUBECONFIG")
		if !ok {
			return ""
		}
		if kubeconfigPath == "" {
			klog.Warning("KUBECONFIG env var is defined but empty")
			return ""
		}
		ok, err := util.PathExists(kubeconfigPath)
		if err != nil {
			klog.Errorf("Failed to verify access to file (%s) defined by KUBECONFIG env var: %v", kubeconfigPath, err)
			return ""
		}
		if !ok {
			klog.Errorf("File (%s) defined by KUBECONFIG env var does not exist", kubeconfigPath)
			return ""
		}

		return kubeconfigPath
	}

	getKubeconfigFromDefaultPath := func() string {
		defaultUserKubeconfig := fmt.Sprintf("%s/.kube/config", os.Getenv("HOME"))
		ok, err := util.PathExists(defaultUserKubeconfig)
		if err != nil {
			klog.Errorf("Failed to verify access to ~/.kube/config: %v", err)
			return ""
		}
		if !ok {
			klog.Errorf("~/.kube/config does not exist")
			return ""
		}
		return defaultUserKubeconfig
	}

	if kubeconfigPath := getKubeconfigFromEnv(); kubeconfigPath != "" {
		klog.Warningf("WARNING: Running healthcheck as non-root user, using KUBECONFIG environment variable: %s", kubeconfigPath)
		return kubeconfigPath
	}

	if kubeconfigPath := getKubeconfigFromDefaultPath(); kubeconfigPath != "" {
		klog.Warningf("WARNING: Running healthcheck as non-root user, using ~/.kube/config")
		return kubeconfigPath
	}

	klog.Errorf("ERROR: Could not find suitable kubeconfig")
	return ""
}

func waitForWorkloads(ctx context.Context, timeout time.Duration, workloads map[string]NamespaceWorkloads) error {
	kubeconfigPath := getKubeconfigPath()
	if kubeconfigPath == "" {
		return fmt.Errorf("could not find existing kubeconfig file")
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig from %s: %v", kubeconfigPath, err)
	}
	client, err := appsclientv1.NewForConfig(rest.AddUserAgent(restConfig, "healthcheck"))
	if err != nil {
		return fmt.Errorf("unable to create Kubernetes client: %v", err)
	}

	coreClient, err := coreclientv1.NewForConfig(rest.AddUserAgent(restConfig, "healthcheck"))
	if err != nil {
		return fmt.Errorf("unable to create Kubernetes core client: %v", err)
	}

	interval := max(timeout/30, 1*time.Second)
	klog.Infof("API Server will be queried every %v", interval)

	aeg := &util.AllErrGroup{}
	for ns, wls := range workloads {
		for _, deploy := range wls.Deployments {
			aeg.Go(func() error { return waitForDeployment(ctx, client, timeout, interval, ns, deploy) })
		}
		for _, ds := range wls.DaemonSets {
			aeg.Go(func() error { return waitForDaemonSet(ctx, client, timeout, interval, ns, ds) })
		}
		for _, sts := range wls.StatefulSets {
			aeg.Go(func() error { return waitForStatefulSet(ctx, client, timeout, interval, ns, sts) })
		}
	}
	errs := aeg.Wait()
	if errs != nil {
		printPostFailureDebugInfo(ctx, coreClient)
		return errs
	}
	return nil
}

func waitForDaemonSet(ctx context.Context, client *appsclientv1.AppsV1Client, timeout, interval time.Duration, namespace, name string) error {
	klog.Infof("Waiting %v for daemonset/%s in %s", timeout, name, namespace)
	var lastHumanReadableErr error
	err := wait.PollUntilContextTimeout(ctx, interval, timeout, true, func(ctx context.Context) (done bool, err error) {
		getctx, cancel := context.WithTimeout(ctx, interval/2)
		defer cancel()

		ds, err := client.DaemonSets(namespace).Get(getctx, name, v1.GetOptions{})
		if err != nil {
			// Always return 'false, nil' to keep retrying until timeout.

			if commonErr := commonGetErrors(err); commonErr != nil {
				lastHumanReadableErr = commonErr
				return false, nil
			}
			if isDeadlineExceededError(err) {
				return false, nil
			}

			klog.Errorf("Unexpected error while getting daemonset %q in %q (ignoring): %v", name, namespace, err)
			return false, nil
		}
		klog.V(3).Infof("Status of DaemonSet %q in %q: %+v", name, namespace, ds.Status)

		// Borrowed and adjusted from k8s.io/kubectl/pkg/polymorphichelpers/rollout_status.go
		if ds.Generation > ds.Status.ObservedGeneration {
			lastHumanReadableErr = fmt.Errorf("daemonset is still being processed by the controller (generation %d > observed %d)", ds.Generation, ds.Status.ObservedGeneration)
			return false, nil
		}
		if ds.Status.UpdatedNumberScheduled < ds.Status.DesiredNumberScheduled {
			lastHumanReadableErr = fmt.Errorf("only %d of %d nodes have the updated daemonset pods", ds.Status.UpdatedNumberScheduled, ds.Status.DesiredNumberScheduled)
			return false, nil
		}
		if ds.Status.NumberAvailable < ds.Status.DesiredNumberScheduled {
			lastHumanReadableErr = fmt.Errorf("only %d of %d daemonset pods are ready across all nodes", ds.Status.NumberAvailable, ds.Status.DesiredNumberScheduled)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		if isDeadlineExceededError(err) {
			klog.Errorf("DaemonSet %q in %q namespace didn't become ready in %v: %v", name, namespace, timeout, lastHumanReadableErr)
			return fmt.Errorf("daemonset '%s' in namespace '%s' failed to become ready within %v. Last status: %v", name, namespace, timeout, lastHumanReadableErr)
		}
		klog.Errorf("Failed waiting for DaemonSet %q in namespace %q: %v", name, namespace, err)
		return err
	}
	klog.Infof("DaemonSet %q in namespace %q is ready", name, namespace)
	return nil
}

func waitForDeployment(ctx context.Context, client *appsclientv1.AppsV1Client, timeout, interval time.Duration, namespace, name string) error {
	klog.Infof("Waiting %v for deployment/%s in %s", timeout, name, namespace)
	var lastHumanReadableErr error
	err := wait.PollUntilContextTimeout(ctx, interval, timeout, true, func(ctx context.Context) (done bool, err error) {
		getctx, cancel := context.WithTimeout(ctx, interval/2)
		defer cancel()

		deployment, err := client.Deployments(namespace).Get(getctx, name, v1.GetOptions{})
		if err != nil {
			// Always return 'false, nil' to keep retrying until timeout.

			if commonErr := commonGetErrors(err); commonErr != nil {
				lastHumanReadableErr = commonErr
				return false, nil
			}
			if isDeadlineExceededError(err) {
				return false, nil
			}

			klog.Errorf("Unexpected error while getting deployment %q in %q (ignoring): %v", name, namespace, err)
			return false, nil
		}
		klog.V(3).Infof("Status of Deployment %q in %q: %+v", name, namespace, deployment.Status)

		// Borrowed and adjusted from k8s.io/kubectl/pkg/polymorphichelpers/rollout_status.go
		if deployment.Generation > deployment.Status.ObservedGeneration {
			lastHumanReadableErr = fmt.Errorf("deployment is still being processed by the controller (generation %d > observed %d)", deployment.Generation, deployment.Status.ObservedGeneration)
			return false, nil
		}
		// 'rollout status' command would check the 'Progressing' condition and if the reason is 'ProgressDeadlineExceeded',
		// it would return an error. We skip it here because:
		// - a false positive error can happen if the node was offline for more than the Deployment's progress deadline
		//   and the healthcheck runs before the controller has started progressing the Deployment again.
		// - we want to give full timeout duration for the Deployment to become ready, no early exits.

		if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
			lastHumanReadableErr = fmt.Errorf("only %d of %d pods have been updated with the latest configuration", deployment.Status.UpdatedReplicas, *deployment.Spec.Replicas)
			return false, nil
		}
		if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
			lastHumanReadableErr = fmt.Errorf("%d pods are still running the old configuration while %d are updated", deployment.Status.Replicas-deployment.Status.UpdatedReplicas, deployment.Status.UpdatedReplicas)
			return false, nil
		}
		if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
			lastHumanReadableErr = fmt.Errorf("only %d of %d updated pods are ready", deployment.Status.AvailableReplicas, deployment.Status.UpdatedReplicas)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		if isDeadlineExceededError(err) {
			klog.Errorf("Deployment/%s in %s didn't become ready in %v: %v", name, namespace, timeout, lastHumanReadableErr)
			return fmt.Errorf("deployment '%s' in namespace '%s' failed to become ready within %v. Last status: %v", name, namespace, timeout, lastHumanReadableErr)
		}
		klog.Errorf("Failed waiting for Deployment %q in namespace %q: %v", name, namespace, err)
		return err
	}
	klog.Infof("Deployment %q in namespace %q is ready", name, namespace)
	return nil
}

func waitForStatefulSet(ctx context.Context, client *appsclientv1.AppsV1Client, timeout, interval time.Duration, namespace, name string) error {
	klog.Infof("Waiting %v for statefulset/%s in %s", timeout, name, namespace)
	var lastHumanReadableErr error
	err := wait.PollUntilContextTimeout(ctx, interval, timeout, true, func(ctx context.Context) (done bool, err error) {
		getctx, cancel := context.WithTimeout(ctx, interval/2)
		defer cancel()

		sts, err := client.StatefulSets(namespace).Get(getctx, name, v1.GetOptions{})
		if err != nil {
			// Always return 'false, nil' to keep retrying until timeout.

			if commonErr := commonGetErrors(err); commonErr != nil {
				lastHumanReadableErr = commonErr
				return false, nil
			}
			if isDeadlineExceededError(err) {
				return false, nil
			}

			klog.Errorf("Unexpected error while getting statefulset %q in %q (ignoring): %v", name, namespace, err)
			return false, nil
		}
		klog.V(3).Infof("Status of StatefulSet %q in %q: %+v", name, namespace, sts.Status)

		// Borrowed and adjusted from k8s.io/kubectl/pkg/polymorphichelpers/rollout_status.go
		if sts.Status.ObservedGeneration == 0 || sts.Generation > sts.Status.ObservedGeneration {
			lastHumanReadableErr = fmt.Errorf("statefulset is still being processed by the controller (generation %d > observed %d)", sts.Generation, sts.Status.ObservedGeneration)
			return false, nil
		}
		if sts.Spec.Replicas != nil && sts.Status.ReadyReplicas < *sts.Spec.Replicas {
			lastHumanReadableErr = fmt.Errorf("only %d of %d replicas are ready", sts.Status.ReadyReplicas, *sts.Spec.Replicas)
			return false, nil
		}
		if sts.Spec.UpdateStrategy.Type == appsv1.RollingUpdateStatefulSetStrategyType && sts.Spec.UpdateStrategy.RollingUpdate != nil {
			if sts.Spec.Replicas != nil && sts.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
				if sts.Status.UpdatedReplicas < (*sts.Spec.Replicas - *sts.Spec.UpdateStrategy.RollingUpdate.Partition) {
					lastHumanReadableErr = fmt.Errorf("only %d of %d replicas have been updated (partition: %d)", sts.Status.UpdatedReplicas, *sts.Spec.Replicas, *sts.Spec.UpdateStrategy.RollingUpdate.Partition)
					return false, nil
				}
			}
			return true, nil
		}
		if sts.Status.UpdateRevision != sts.Status.CurrentRevision {
			lastHumanReadableErr = fmt.Errorf("update revision (%s) differs from current revision (%s)", sts.Status.UpdateRevision, sts.Status.CurrentRevision)
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		if isDeadlineExceededError(err) {
			klog.Errorf("Statefulset/%s in %s didn't become ready in %v: %v", name, namespace, timeout, lastHumanReadableErr)
			return fmt.Errorf("statefulset '%s' in namespace '%s' failed to become ready within %v. Last status: %v", name, namespace, timeout, lastHumanReadableErr)
		}
		klog.Errorf("Failed waiting for StatefulSet %q in namespace %q: %v", name, namespace, err)
		return err
	}
	klog.Infof("StatefulSet %q in namespace %q is ready", name, namespace)
	return nil
}

func isDeadlineExceededError(err error) bool {
	if strings.Contains(err.Error(), "would exceed context deadline") {
		return true
	}

	// 'client rate limiter Wait returned an error: context deadline exceeded' -> drop the wrapping errors
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	return false
}

func commonGetErrors(err error) error {
	if apierrors.IsNotFound(err) {
		// Resources created by an operator might not exist yet.
		// We allow for full timeout duration to be created and become ready.
		return fmt.Errorf("resource does not exist yet")
	}

	if errors.Is(err, syscall.ECONNREFUSED) {
		return fmt.Errorf("cannot connect to API server")
	}

	return nil
}
