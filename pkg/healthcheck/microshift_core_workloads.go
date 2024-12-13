package healthcheck

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/openshift/microshift/pkg/config"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"

	coreclientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

// getMicroShiftWorkloads assembles a structure with the MicroShift
// workloads that the healthcheck should verify: both core
// (expected every time) and optional (deployed by MicroShift shipped RPMs).
func getMicroShiftWorkloads(ctx context.Context) (map[string]NamespaceWorkloads, error) {
	workloads := map[string]NamespaceWorkloads{
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

	cfg, err := config.ActiveConfig()
	if err != nil {
		return nil, err
	}

	if cfg.Network.IsEnabled() {
		workloads["openshift-ovn-kubernetes"] = NamespaceWorkloads{
			DaemonSets: []string{"ovnkube-master", "ovnkube-node"},
		}
	}

	storageComponents(cfg, workloads)
	if err := optionalComponents(ctx, workloads); err != nil {
		return nil, err
	}

	return workloads, nil
}

func storageComponents(cfg *config.Config, workloads map[string]NamespaceWorkloads) {
	klog.V(2).Infof("Configured storage driver value: %q", string(cfg.Storage.Driver))
	if cfg.Storage.IsEnabled() {
		klog.Infof("LVMS is enabled")
		workloads["openshift-storage"] = NamespaceWorkloads{
			DaemonSets:  []string{"vg-manager"},
			Deployments: []string{"lvms-operator"},
		}
	}
	if comps := getExpectedCSIComponents(cfg); len(comps) != 0 {
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
			klog.Infof("CSI Snapshot Controller is enabled")
			deployments = append(deployments, "csi-snapshot-controller")
		}
		if comp == config.CsiComponentSnapshotWebhook {
			klog.Infof("CSI Snapshot Webhook is enabled")
			deployments = append(deployments, "csi-snapshot-webhook")
		}
	}
	return deployments
}

// optionalComponents checks for existence of namespaces that are deployed
// using MicroShift's optional RPMs. If the namespace exists, the workloads
// are added to the map of expected readiness.
func optionalComponents(ctx context.Context, workloads map[string]NamespaceWorkloads) error {
	optionalComponents := map[string]NamespaceWorkloads{
		"openshift-multus": {
			DaemonSets: []string{"multus", "dhcp-daemon"},
		},
		"openshift-operator-lifecycle-manager": {
			Deployments: []string{"olm-operator", "catalog-operator"},
		},
		"openshift-gateway-api": {
			Deployments: []string{"servicemesh-operator3", "istiod-openshift-gateway-api"},
		},
		"kube-flannel": {
			DaemonSets: []string{"kube-flannel-ds"},
		},
		"kube-proxy": {
			DaemonSets: []string{"kube-proxy"},
		},
	}

	restConfig, err := clientcmd.BuildConfigFromFlags("", filepath.Join(config.DataDir, "resources", string(config.KubeAdmin), "kubeconfig"))
	if err != nil {
		return fmt.Errorf("failed to create restConfig: %v", err)
	}
	client, err := coreclientv1.NewForConfig(restConfig)
	if err != nil {
		return fmt.Errorf("failed to create client: %v", err)
	}

	for ns, wls := range optionalComponents {
		if _, err = client.Namespaces().Get(ctx, ns, v1.GetOptions{}); err != nil {
			if !apierrors.IsNotFound(err) {
				klog.Errorf("Failure getting %q namespace: %v", ns, err)
				return err
			}
		} else {
			workloads[ns] = wls
		}
	}

	return nil
}
