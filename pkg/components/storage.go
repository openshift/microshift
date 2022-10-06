package components

import (
	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"k8s.io/klog/v2"
)

func startCSIPlugin(cfg *config.MicroshiftConfig, kubeconfigPath string) error {
	var (
		ns = []string{
			"assets/components/odf-lvm/topolvm-openshift-storage_namespace.yaml",
		}
		sa = []string{
			"assets/components/odf-lvm/topolvm-node_v1_serviceaccount.yaml",
			"assets/components/odf-lvm/topolvm-controller_v1_serviceaccount.yaml",
		}
		role = []string{
			"assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_role.yaml",
			"assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_role.yaml",
			"assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_role.yaml",
		}
		rb = []string{
			"assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_rolebinding.yaml",
			"assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_rolebinding.yaml",
			"assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_rolebinding.yaml",
		}
		cr = []string{
			"assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_clusterrole.yaml",
			"assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_clusterrole.yaml",
			"assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_clusterrole.yaml",
			"assets/components/odf-lvm/topolvm-node-scc_rbac.authorization.k8s.io_v1_clusterrole.yaml",
			"assets/components/odf-lvm/topolvm-node_rbac.authorization.k8s.io_v1_clusterrole.yaml",
		}
		crb = []string{
			"assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml",
			"assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml",
			"assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml",
			"assets/components/odf-lvm/topolvm-node-scc_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml",
			"assets/components/odf-lvm/topolvm-node_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml",
		}
		cd = []string{
			"assets/components/odf-lvm/csi-driver.yaml",
		}
		cm = []string{
			"assets/components/odf-lvm/topolvm-lvmd-config_configmap_v1.yaml",
		}
		ds = []string{
			"assets/components/odf-lvm/topolvm-node_daemonset.yaml",
		}
		deploy = []string{
			"assets/components/odf-lvm/topolvm-controller_deployment.yaml",
		}
		sc = []string{
			"assets/components/odf-lvm/topolvm_default-storage-class.yaml",
		}
		scc = []string{
			"assets/components/odf-lvm/topolvm-node-securitycontextconstraint.yaml",
		}
	)
	if err := assets.ApplyStorageClasses(sc, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply storage cass %v: %v", sc, err)
		return err
	}
	if err := assets.ApplyCSIDrivers(cd, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply csiDriver %v: %v", sc, err)
		return err
	}
	if err := assets.ApplyNamespaces(ns, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply ns %v: %v", ns, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(sa, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply sa %v: %v", sa, err)
		return err
	}
	if err := assets.ApplyRoles(role, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply role %v: %v", cr, err)
		return err
	}
	if err := assets.ApplyRoleBindings(rb, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply rolebinding %v: %v", cr, err)
		return err
	}
	if err := assets.ApplyClusterRoles(cr, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterrole %v: %v", cr, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(crb, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterrolebinding %v: %v", crb, err)
		return err
	}
	if err := assets.ApplyConfigMaps(cm, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply configMap %v: %v", crb, err)
		return err
	}
	if err := assets.ApplyDeployments(deploy, renderTemplate, renderParamsFromConfig(cfg, nil), kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply deployment %v: %v", deploy, err)
		return err
	}
	if err := assets.ApplyDaemonSets(ds, renderTemplate, renderParamsFromConfig(cfg, nil), kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply daemonsets %v: %v", ds, err)
		return err
	}
	if err := assets.ApplySCCs(scc, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply sccs %v: %v", scc, err)
		return err
	}
	return nil
}

func startCSISnapshot(cfg *config.MicroshiftConfig, kubeconfigPath string) error {
	var (
		sa = []string{
			"assets/components/csi-snapshot/service-account.yaml",
		}
		role = []string{
			"assets/components/csi-snapshot/role.yaml",
		}
		rb = []string{
			"assets/components/csi-snapshot/rolebinding.yaml",
		}
		cr = []string{
			"assets/components/csi-snapshot/clusterrole.yaml",
		}
		crb = []string{
			"assets/components/csi-snapshot/clusterrolebinding.yaml",
		}
		deploy = []string{
			"assets/components/csi-snapshot/controller-deployment.yaml",
			"assets/components/csi-snapshot/webhook-deployment.yaml",
		}
		svc = []string{
			"assets/components/csi-snapshot/webhook-service.yaml",
		}
		vwc = []string {
			"assets/components/csi-snapshot/validating-webhook-config.yaml",
		}
		// scc = []string{
		// 	"assets/components/csi-snapshot/topolvm-node-securitycontextconstraint.yaml",
		// }
	)

	if err := assets.ApplyServiceAccounts(sa, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply sa %v: %v", sa, err)
		return err
	}
	if err := assets.ApplyRoles(role, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply role %v: %v", cr, err)
		return err
	}
	if err := assets.ApplyRoleBindings(rb, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply rolebinding %v: %v", cr, err)
		return err
	}
	if err := assets.ApplyClusterRoles(cr, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterrole %v: %v", cr, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(crb, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterrolebinding %v: %v", crb, err)
		return err
	}
	if err := assets.ApplyDeployments(deploy, renderTemplate, renderParamsFromConfig(cfg, nil), kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply deployment %v: %v", deploy, err)
		return err
	}

	if err := assets.ApplyServices(svc, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply services %v: %v", svc, err)
		return err
	}

	if err := assets.ApplyValidatingWebhookConfiguration(vwc, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply ValidatingWebhookConfiguration %v: %v", vwc, err)
		return err
	}
	// if err := assets.ApplySCCs(scc, nil, nil, kubeconfigPath); err != nil {
	// 	klog.Warningf("Failed to apply sccs %v: %v", scc, err)
	// 	return err
	// }
	return nil
}
