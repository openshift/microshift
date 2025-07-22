package nsexemptions

import "k8s.io/apimachinery/pkg/util/sets"

// systemNSSyncExemptions is the list of namespaces deployed by an OpenShift install
// payload, as retrieved by listing the namespaces after a successful installation
// IMPORTANT: The Namespace openshift-operators must be an exception to this rule
// since it is used by OCP/OLM users to install their Operator bundle solutions.
var systemNSSyncExemptions = sets.NewString(
	// kube-specific system namespaces
	"default",
	"kube-node-lease",
	"kube-public",
	"kube-system",

	// openshift payload namespaces
	"openshift",
	"openshift-apiserver",
	"openshift-apiserver-operator",
	"openshift-authentication",
	"openshift-authentication-operator",
	"openshift-cloud-controller-manager",
	"openshift-cloud-controller-manager-operator",
	"openshift-cloud-credential-operator",
	"openshift-cloud-network-config-controller",
	"openshift-cluster-csi-drivers",
	"openshift-cluster-machine-approver",
	"openshift-cluster-node-tuning-operator",
	"openshift-cluster-samples-operator",
	"openshift-cluster-storage-operator",
	"openshift-cluster-version",
	"openshift-config",
	"openshift-config-managed",
	"openshift-config-operator",
	"openshift-console",
	"openshift-console-operator",
	"openshift-console-user-settings",
	"openshift-controller-manager",
	"openshift-controller-manager-operator",
	"openshift-dns",
	"openshift-dns-operator",
	"openshift-etcd",
	"openshift-etcd-operator",
	"openshift-host-network",
	"openshift-image-registry",
	"openshift-infra",
	"openshift-ingress",
	"openshift-ingress-canary",
	"openshift-ingress-operator",
	"openshift-insights",
	"openshift-kni-infra",
	"openshift-kube-apiserver",
	"openshift-kube-apiserver-operator",
	"openshift-kube-controller-manager",
	"openshift-kube-controller-manager-operator",
	"openshift-kube-scheduler",
	"openshift-kube-scheduler-operator",
	"openshift-kube-storage-version-migrator",
	"openshift-kube-storage-version-migrator-operator",
	"openshift-machine-api",
	"openshift-machine-config-operator",
	"openshift-marketplace",
	"openshift-monitoring",
	"openshift-multus",
	"openshift-network-diagnostics",
	"openshift-network-operator",
	"openshift-node",
	"openshift-nutanix-infra",
	"openshift-oauth-apiserver",
	"openshift-openstack-infra",
	"openshift-operator-lifecycle-manager",
	"openshift-ovirt-infra",
	"openshift-sdn",
	"openshift-service-ca",
	"openshift-service-ca-operator",
	"openshift-user-workload-monitoring",
	"openshift-vsphere-infra",
)

// IsNamespacePSALabelSyncExemptedInVendoredOCPVersion returns true if the given namespace should be exempted from
// PSA label sync'ing. NOTE: the exemption list is OCP version dependent. Ensure that your vendored
// version of 'cluster-policy-controller' is for the same OCP version as your project.
func IsNamespacePSALabelSyncExemptedInVendoredOCPVersion(namespace string) bool {
	return systemNSSyncExemptions.Has(namespace)
}
