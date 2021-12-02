package controller

var ControllerInitializers = map[string]InitFunc{
	"openshift.io/namespace-security-allocation": RunNamespaceSecurityAllocationController,
	"openshift.io/resourcequota":                 RunResourceQuotaManager,
	"openshift.io/cluster-quota-reconciliation":  RunClusterQuotaReconciliationController,
	"openshift.io/cluster-csr-approver":          RunCSRApproverController,
}

const (
	infraClusterQuotaReconciliationControllerServiceAccountName  = "cluster-quota-reconciliation-controller"
	infraClusterCSRApproverControllerServiceAccountName          = "cluster-csr-approver-controller"
	infraNamespaceSecurityAllocationControllerServiceAccountName = "namespace-security-allocation-controller"
	defaultOpenShiftInfraNamespace                               = "openshift-infra"
)
