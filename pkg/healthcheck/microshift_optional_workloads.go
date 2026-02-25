package healthcheck

import (
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
}

// fillOptionalMicroShiftWorkloads assembles list of optional MicroShift workloads
// existing on the filesystem as manifests (in comparison to Multus which
// manifests are part of MicroShift binary).
func fillOptionalMicroShiftWorkloads(workloadsToCheck map[string]NamespaceWorkloads) error {
	for path, ow := range optionalWorkloadPaths {
		if exists, err := util.PathExists(path); err != nil {
			return err
		} else if exists {
			klog.Infof("Optional component path exists: %s - expecting %v in namespace %q", path, ow.Workloads.String(), ow.Namespace)
			workloadsToCheck[ow.Namespace] = ow.Workloads
		}
	}
	return nil
}
