package healthcheck

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/openshift/microshift/pkg/config"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	appsclientv1 "k8s.io/client-go/kubernetes/typed/apps/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
	deploymentutil "k8s.io/kubectl/pkg/util/deployment"
)

type NamespaceWorkloads struct {
	Deployments  []string `json:"deployments"`
	DaemonSets   []string `json:"daemonsets"`
	StatefulSets []string `json:"statefulsets"`
}

func waitForWorkloads(ctx context.Context, timeout time.Duration, workloads map[string]NamespaceWorkloads) error {
	restConfig, err := clientcmd.BuildConfigFromFlags("", filepath.Join(config.DataDir, "resources", string(config.KubeAdmin), "kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create restConfig: %v", err)
	}
	client, err := appsclientv1.NewForConfig(rest.AddUserAgent(restConfig, "healthcheck"))
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	aeg := &AllErrGroup{}
	for ns, wls := range workloads {
		for _, deploy := range wls.Deployments {
			aeg.Go(func() error { return waitForDeployment(ctx, client, timeout, ns, deploy) })
		}
		for _, ds := range wls.DaemonSets {
			aeg.Go(func() error { return waitForDaemonSet(ctx, client, timeout, ns, ds) })
		}
		for _, sts := range wls.StatefulSets {
			aeg.Go(func() error { return waitForStatefulSet(ctx, client, timeout, ns, sts) })
		}
	}
	errs := aeg.Wait()
	if errs != nil {
		logPodsAndEvents()
		return errs
	}
	return nil
}

func waitForDaemonSet(ctx context.Context, client *appsclientv1.AppsV1Client, timeout time.Duration, namespace, name string) error {
	klog.Infof("Waiting %v for daemonset/%s in %s", timeout, name, namespace)
	err := wait.PollUntilContextTimeout(ctx, 10*time.Second, timeout, true, func(ctx context.Context) (done bool, err error) {
		ds, err := client.DaemonSets(namespace).Get(ctx, name, v1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				// Resources created by an operator might not exist yet.
				// We allow for full timeout duration to be created and become ready.
				return false, nil
			}
			klog.Errorf("Error getting daemonset/%s in %q: %v", name, namespace, err)
			// Ignore errors, give chance until timeout
			return false, nil
		}
		klog.V(3).Infof("Status of daemonset/%s in %s: %+v", name, namespace, ds.Status)

		// Borrowed and adjusted from k8s.io/kubectl/pkg/polymorphichelpers/rollout_status.go
		if ds.Generation > ds.Status.ObservedGeneration {
			return false, nil
		}
		if ds.Status.UpdatedNumberScheduled < ds.Status.DesiredNumberScheduled {
			return false, nil
		}
		if ds.Status.NumberAvailable < ds.Status.DesiredNumberScheduled {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		klog.Errorf("Failed waiting for daemonset/%s in %s: %v", name, namespace, err)
		return err
	}
	klog.Infof("Daemonset/%s in %s is ready", name, namespace)
	return nil
}

func waitForDeployment(ctx context.Context, client *appsclientv1.AppsV1Client, timeout time.Duration, namespace, name string) error {
	klog.Infof("Waiting %v for deployment/%s in %s", timeout, name, namespace)
	err := wait.PollUntilContextTimeout(ctx, 10*time.Second, timeout, true, func(ctx context.Context) (done bool, err error) {
		deployment, err := client.Deployments(namespace).Get(ctx, name, v1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				// Resources created by an operator might not exist yet.
				// We allow for full timeout duration to be created and become ready.
				return false, nil
			}
			klog.Errorf("Error getting deployment/%s in %q: %v", name, namespace, err)
			// Ignore errors, give chance until timeout
			return false, nil
		}
		klog.V(3).Infof("Status of deployment/%s in %s: %+v", name, namespace, deployment.Status)

		// Borrowed and adjusted from k8s.io/kubectl/pkg/polymorphichelpers/rollout_status.go
		if deployment.Generation > deployment.Status.ObservedGeneration {
			return false, nil
		}
		cond := deploymentutil.GetDeploymentCondition(deployment.Status, appsv1.DeploymentProgressing)
		if cond != nil && cond.Reason == deploymentutil.TimedOutReason {
			return false, fmt.Errorf("deployment %q exceeded its progress deadline", deployment.Name)
		}
		if deployment.Spec.Replicas != nil && deployment.Status.UpdatedReplicas < *deployment.Spec.Replicas {
			return false, nil
		}
		if deployment.Status.Replicas > deployment.Status.UpdatedReplicas {
			return false, nil
		}
		if deployment.Status.AvailableReplicas < deployment.Status.UpdatedReplicas {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		klog.Errorf("Failed waiting for deployment/%s in %s: %v", name, namespace, err)
		return err
	}
	klog.Infof("Deployment/%s in %s is ready", name, namespace)
	return nil
}

func waitForStatefulSet(ctx context.Context, client *appsclientv1.AppsV1Client, timeout time.Duration, namespace, name string) error {
	klog.Infof("Waiting %v for statefulset/%s in %s", timeout, name, namespace)
	err := wait.PollUntilContextTimeout(ctx, 10*time.Second, timeout, true, func(ctx context.Context) (done bool, err error) {
		sts, err := client.StatefulSets(namespace).Get(ctx, name, v1.GetOptions{})
		if err != nil {
			if apierrors.IsNotFound(err) {
				// Resources created by an operator might not exist yet.
				// We allow for full timeout duration to be created and become ready.
				return false, nil
			}
			klog.Errorf("Error getting statefulset/%s in %s: %v", name, namespace, err)
			// Ignore errors, give chance until timeout
			return false, nil
		}
		klog.V(3).Infof("Status of statefulset/%s in %s: %+v", name, namespace, sts.Status)

		// Borrowed and adjusted from k8s.io/kubectl/pkg/polymorphichelpers/rollout_status.go
		if sts.Status.ObservedGeneration == 0 || sts.Generation > sts.Status.ObservedGeneration {
			return false, nil
		}
		if sts.Spec.Replicas != nil && sts.Status.ReadyReplicas < *sts.Spec.Replicas {
			return false, nil
		}
		if sts.Spec.UpdateStrategy.Type == appsv1.RollingUpdateStatefulSetStrategyType && sts.Spec.UpdateStrategy.RollingUpdate != nil {
			if sts.Spec.Replicas != nil && sts.Spec.UpdateStrategy.RollingUpdate.Partition != nil {
				if sts.Status.UpdatedReplicas < (*sts.Spec.Replicas - *sts.Spec.UpdateStrategy.RollingUpdate.Partition) {
					return false, nil
				}
			}
			return true, nil
		}
		if sts.Status.UpdateRevision != sts.Status.CurrentRevision {
			return false, nil
		}
		return true, nil
	})
	if err != nil {
		klog.Errorf("Failed waiting for statefulset/%s in %s: %v", name, namespace, err)
		return err
	}
	klog.Infof("StatefulSet/%s in %s is ready", name, namespace)
	return nil
}
