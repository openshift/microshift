package components

import (
	"github.com/openshift/microshift/pkg/assets"
	"k8s.io/klog/v2"
)

func startHostpathProvisioner(kubeconfigPath string) error {
	var (
		ns = []string{
			"assets/components/hostpath-provisioner/namespace.yaml",
		}
		sa = []string{
			"assets/components/hostpath-provisioner/service-account.yaml",
		}
		cr = []string{
			"assets/components/hostpath-provisioner/clusterrole.yaml",
		}
		crb = []string{
			"assets/components/hostpath-provisioner/clusterrolebinding.yaml",
		}
		scc = []string{
			"assets/components/hostpath-provisioner/scc.yaml",
		}
		ds = []string{
			"assets/components/hostpath-provisioner/daemonset.yaml",
		}
		sc = []string{
			"assets/components/hostpath-provisioner/storageclass.yaml",
		}
	)
	if err := assets.ApplyNamespaces(ns, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply ns %v: %v", ns, err)
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
	if err := assets.ApplyServiceAccounts(sa, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply sa %v: %v", sa, err)
		return err
	}
	if err := assets.ApplySCCs(scc, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply scc %v: %v", scc, err)
		return err
	}
	if err := assets.ApplyDaemonSets(ds, renderReleaseImage, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply daemonsets %v: %v", ds, err)
		return err
	}
	if err := assets.ApplyStorageClasses(sc, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply storage cass %v: %v", sc, err)
		return err
	}
	return nil
}
