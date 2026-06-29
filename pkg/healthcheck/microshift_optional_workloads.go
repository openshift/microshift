package healthcheck

import (
	"slices"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/util"
	"k8s.io/klog/v2"
)

type optionalWorkloads struct {
	Namespace string
	Workloads NamespaceWorkloads
}

// optionalWorkloadPaths defines the mapping of manifest filepath to the namespace and workloads.
var optionalWorkloadPaths = map[string]optionalWorkloads{
	"/usr/lib/microshift/manifests.d/001-microshift-olm": {
		Namespace: "openshift-operator-lifecycle-manager",
		Workloads: NamespaceWorkloads{Deployments: []string{"olm-operator", "catalog-operator"}},
	},

	"/usr/lib/microshift/manifests.d/000-microshift-gateway-api": {
		Namespace: "openshift-gateway-api",
		Workloads: NamespaceWorkloads{
			Deployments: []string{"servicemesh-operator3", "istiod-openshift-gateway-api"},
		},
	},

	"/usr/lib/microshift/manifests.d/060-microshift-cert-manager": {
		Namespace: "cert-manager",
		Workloads: NamespaceWorkloads{Deployments: []string{"cert-manager", "cert-manager-webhook", "cert-manager-cainjector"}},
	},

	"/usr/lib/microshift/manifests.d/010-microshift-ai-model-serving-kserve": {
		Namespace: "redhat-ods-applications",
		Workloads: NamespaceWorkloads{Deployments: []string{"kserve-controller-manager"}},
	},
	"/usr/lib/microshift/manifests.d/070-microshift-sriov": {
		Namespace: "sriov-network-operator",
		Workloads: NamespaceWorkloads{Deployments: []string{"sriov-network-operator"}},
	},

	"/usr/lib/microshift/manifests.d/080-microshift-metrics-server": {
		Namespace: "openshift-monitoring",
		Workloads: NamespaceWorkloads{Deployments: []string{"metrics-server"}},
	},
	"/usr/lib/microshift/manifests.d/081-microshift-kube-state-metrics": {
		Namespace: "openshift-monitoring",
		Workloads: NamespaceWorkloads{Deployments: []string{"kube-state-metrics"}},
	},
	"/usr/lib/microshift/manifests.d/082-microshift-node-exporter": {
		Namespace: "openshift-monitoring",
		Workloads: NamespaceWorkloads{DaemonSets: []string{"node-exporter"}},
	},
}

// mergeWorkloads combines two NamespaceWorkloads into one.
func mergeWorkloads(existing, incoming NamespaceWorkloads) NamespaceWorkloads {
	return NamespaceWorkloads{
		Deployments:  slices.Concat(existing.Deployments, incoming.Deployments),
		DaemonSets:   slices.Concat(existing.DaemonSets, incoming.DaemonSets),
		StatefulSets: slices.Concat(existing.StatefulSets, incoming.StatefulSets),
	}
}

// fillOptionalMicroShiftWorkloads assembles list of optional MicroShift workloads
// that are both present on the filesystem and included in the configured
// kustomizePaths. This ensures the healthcheck only waits for optional
// components that MicroShift was configured to deploy.
func fillOptionalMicroShiftWorkloads(workloadsToCheck map[string]NamespaceWorkloads) error {
	cfg, err := config.ActiveConfig()
	if err != nil {
		return err
	}

	configuredPaths, err := cfg.Manifests.GetKustomizationPaths()
	if err != nil {
		return err
	}

	configuredSet := make(map[string]bool, len(configuredPaths))
	for _, p := range configuredPaths {
		configuredSet[p] = true
	}

	for path, ow := range optionalWorkloadPaths {
		if exists, err := util.PathExists(path); err != nil {
			return err
		} else if !exists {
			continue
		}

		if !configuredSet[path] {
			klog.Infof("Optional component path exists but is not in configured kustomizePaths: %s - skipping", path)
			continue
		}

		klog.Infof("Optional component path exists and is configured: %s - expecting %v in namespace %q", path, ow.Workloads.String(), ow.Namespace)
		workloadsToCheck[ow.Namespace] = mergeWorkloads(workloadsToCheck[ow.Namespace], ow.Workloads)
	}
	return nil
}
