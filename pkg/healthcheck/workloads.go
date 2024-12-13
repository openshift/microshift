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

func getCoreMicroShiftWorkloads() (map[string]NamespaceWorkloads, error) {
	cfg, err := config.ActiveConfig()
	if err != nil {
		return nil, err
	}

	workloads := map[string]NamespaceWorkloads{
		"openshift-ovn-kubernetes": {
			DaemonSets: []string{"ovnkube-master", "ovnkube-node"},
		},
		"openshift-service-ca": {
			Deployments: []string{"service-ca"},
		},
		"openshift-ingress": {
			Deployments: []string{"router-default"},
		},
		"openshift-dns": {
			DaemonSets: []string{
				"dns-default",
				"node-resolver",
			},
		},
	}
	fillOptionalWorkloadsIfApplicable(cfg, workloads)

	return workloads, nil
}

func waitForCoreWorkloads(ctx context.Context, timeout time.Duration) error {
	workloads, err := getCoreMicroShiftWorkloads()
	if err != nil {
		return err
	}

	return waitForWorkloads(ctx, timeout, workloads)
}

func fillOptionalWorkloadsIfApplicable(cfg *config.Config, workloads map[string]NamespaceWorkloads) {
	klog.V(2).Infof("Configured storage driver value: %q", string(cfg.Storage.Driver))
	if cfg.Storage.IsEnabled() {
		klog.Infof("LVMS is enabled")
		workloads["openshift-storage"] = NamespaceWorkloads{
			DaemonSets:  []string{"vg-manager"},
			Deployments: []string{"lvms-operator"},
		}
	}
	if comps := getExpectedCSIComponents(cfg); len(comps) != 0 {
		klog.Infof("At least one CSI Component is enabled")
		workloads["kube-system"] = NamespaceWorkloads{
			Deployments: comps,
		}
	}
}

func getExpectedCSIComponents(cfg *config.Config) []string {
	klog.V(2).Infof("Configured optional CSI components: %v", cfg.Storage.OptionalCSIComponents)

	if len(cfg.Storage.OptionalCSIComponents) == 0 {
		return []string{"csi-snapshot-controller", "csi-snapshot-webhook"}
	}

	// Validation fails when there's more than one component provided and one of them is "None".
	// In other words: if "None" is used, it can be the only element.
	if len(cfg.Storage.OptionalCSIComponents) == 1 && cfg.Storage.OptionalCSIComponents[0] == config.CsiComponentNone {
		return nil
	}

	deployments := []string{}
	for _, comp := range cfg.Storage.OptionalCSIComponents {
		if comp == config.CsiComponentSnapshot {
			deployments = append(deployments, "csi-snapshot-controller")
		}
		if comp == config.CsiComponentSnapshotWebhook {
			deployments = append(deployments, "csi-snapshot-webhook")
		}
	}
	return deployments
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
		return errs
	}
	return nil
}

func waitForDaemonSet(ctx context.Context, client *appsclientv1.AppsV1Client, timeout time.Duration, namespace, name string) error {
	klog.Infof("Waiting for daemonset/%s in %s", name, namespace)
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
	klog.Infof("Waiting for deployment/%s in %s", name, namespace)
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
	klog.Infof("Waiting for statefulset/%s in %s", name, namespace)
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
