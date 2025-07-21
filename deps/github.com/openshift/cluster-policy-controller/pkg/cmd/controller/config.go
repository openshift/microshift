package controller

var ControllerInitializers = map[string]InitFunc{
	"openshift.io/namespace-security-allocation":          RunNamespaceSecurityAllocationController,
	"openshift.io/resourcequota":                          RunResourceQuotaManager,
	"openshift.io/cluster-quota-reconciliation":           RunClusterQuotaReconciliationController,
	"openshift.io/cluster-csr-approver":                   RunCSRApproverController,
	"openshift.io/podsecurity-admission-label-syncer":     runPodSecurityAdmissionLabelSynchronizationController,
	"openshift.io/privileged-namespaces-psa-label-syncer": runPrivilegedNamespacesPSALabelSyncer,
}

const (
	infraClusterQuotaReconciliationControllerServiceAccountName           = "cluster-quota-reconciliation-controller"
	infraClusterCSRApproverControllerServiceAccountName                   = "cluster-csr-approver-controller"
	infraNamespaceSecurityAllocationControllerServiceAccountName          = "namespace-security-allocation-controller"
	podSecurityAdmissionLabelSyncerControllerServiceAccountName           = "podsecurity-admission-label-syncer-controller"
	privilegedNamespacesPodSecurityAdmissionLabelSyncerServiceAccountName = "privileged-namespaces-psa-label-syncer"

	defaultOpenShiftInfraNamespace = "openshift-infra"
)
