package components

import (
	"github.com/openshift/microshift/pkg/assets"
	"k8s.io/klog/v2"
)

func startLVMProvisioner(kubeconfigPath string) error {
	var (
		ns = []string{
			"assets/components/odf-lvm/topolvm-openshift-storage_namespace.yaml",
		}
		sa = []string{
			"assets/components/odf-lvm/topolvm-node_v1_serviceaccount.yaml",
			"assets/components/odf-lvm/topolvm-controller_v1_serviceaccount.yaml",
		}
		roles = []string{
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
			"assets/components/odf-lvm/topolvm-node-scc_rbac.authorization.k8s.io_v1_clusterrole.yaml",
		}
		crb = []string{
			"assets/components/odf-lvm/topolvm-controller_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml",
			"assets/components/odf-lvm/topolvm-csi-provisioner_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml",
			"assets/components/odf-lvm/topolvm-csi-resizer_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml",
			"assets/components/odf-lvm/topolvm-node-scc_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml",
			"assets/components/odf-lvm/topolvm-node_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml",
			"assets/components/odf-lvm/topolvm-node-scc_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml",
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
	)
	if err := assets.ApplyStorageClasses(sc, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply storage cass %v: %v", sc, err)
		panic(err)
		//return err
	}
	if err := assets.Apply(sc, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply storage cass %v: %v", sc, err)
		panic(err)
		//return err
	}

	if err := assets.ApplyNamespaces(ns, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply ns %v: %v", ns, err)
		panic(err)
		//return err
	}
	if err := assets.ApplyServiceAccounts(sa, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply sa %v: %v", sa, err)
		panic(err)
		//return err
	}
	if err := assets.ApplyRoles(roles, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply role %v: %v", cr, err)
		panic(err)
		//return err
	}
	if err := assets.ApplyRoleBindings(rb, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply rolebinding %v: %v", cr, err)
		panic(err)
		//return err
	}
	if err := assets.ApplyClusterRoles(cr, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterrole %v: %v", cr, err)
		panic(err)
		//return err
	}
	if err := assets.ApplyClusterRoleBindings(crb, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterrolebinding %v: %v", crb, err)
		panic(err)
		//return err
	}
	if err := assets.ApplyDeployments(deploy, renderReleaseImage, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply deployment %v: %v", deploy, err)
		panic(err)
		//return err
	}
	if err := assets.ApplyDaemonSets(ds, renderReleaseImage, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply daemonsets %v: %v", ds, err)
		panic(err)
		//return err
	}
	return nil
}
