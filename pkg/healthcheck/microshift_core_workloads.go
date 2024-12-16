package healthcheck

import (
	"github.com/openshift/microshift/pkg/config"
	"k8s.io/klog/v2"
)

// getCoreMicroShiftWorkloads assembles a structure with the core MicroShift
// workloads that the healthcheck should verify.
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
