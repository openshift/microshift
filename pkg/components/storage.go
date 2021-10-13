package components

import (
	"github.com/openshift/microshift/pkg/assets"
	"github.com/sirupsen/logrus"
)

func startHostpathProvisioner(kubeconfigPath string) error {
	var (
		ns = []string{
			"assets/core/0000_80_hostpath-provisioner-namespace.yaml",
		}
		sa = []string{
			"assets/core/0000_80_hostpath-provisioner-serviceaccount.yaml",
		}
		cr = []string{
			"assets/rbac/0000_80_hostpath-provisioner-clusterrole.yaml",
		}
		crb = []string{
			"assets/rbac/0000_80_hostpath-provisioner-clusterrolebinding.yaml",
		}
		scc = []string{
			"assets/scc/0000_80_hostpath-provisioner-securitycontextconstraints.yaml",
		}
		ds = []string{
			"assets/apps/000_80_hostpath-provisioner-daemonset.yaml",
		}
		sc = []string{
			"assets/storage/0000_80_hostpath-provisioner-storageclass.yaml",
		}
	)
	if err := assets.ApplyNamespaces(ns, kubeconfigPath); err != nil {
		logrus.Warningf("failed to apply ns %v: %v", ns, err)
		return err
	}
	if err := assets.ApplyClusterRoles(cr, kubeconfigPath); err != nil {
		logrus.Warningf("failed to apply clusterrole %v: %v", cr, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(crb, kubeconfigPath); err != nil {
		logrus.Warningf("failed to apply clusterrolebinding %v: %v", crb, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(sa, kubeconfigPath); err != nil {
		logrus.Warningf("failed to apply sa %v: %v", sa, err)
		return err
	}
	if err := assets.ApplySCCs(scc, nil, nil, kubeconfigPath); err != nil {
		logrus.Warningf("failed to apply scc %v: %v", scc, err)
		return err
	}
	if err := assets.ApplyDaemonSets(ds, renderReleaseImage, nil, kubeconfigPath); err != nil {
		logrus.Warningf("failed to apply ds %v: %v", ds, err)
		return err
	}
	if err := assets.ApplyStorageClasses(sc, nil, nil, kubeconfigPath); err != nil {
		logrus.Warningf("failed to apply sc %v: %v", sc, err)
		return err
	}
	return nil
}
