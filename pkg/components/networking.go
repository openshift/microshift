package components

import (
	"github.com/openshift/microshift/pkg/assets"
	"github.com/sirupsen/logrus"
)

func startFlannel(kubeconfigPath string) error {
	var (
		// psp = []string{
		// 	"assets/rbac/0000_00_podsecuritypolicy-flannel.yaml",
		// }
		cr = []string{
			"assets/rbac/0000_00_flannel-clusterrole.yaml",
		}
		crb = []string{
			"assets/rbac/0000_00_flannel-clusterrolebinding.yaml",
		}
		sa = []string{
			"assets/core/0000_00_flannel-service-account.yaml",
		}
		cm = []string{
			"assets/core/0000_00_flannel-configmap.yaml",
		}
		ds = []string{
			"assets/apps/0000_00_flannel-daemonset.yaml",
		}
	)

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
	if err := assets.ApplyConfigMaps(cm, kubeconfigPath); err != nil {
		logrus.Warningf("failed to apply cm %v: %v", cm, err)
		return err
	}
	if err := assets.ApplyDaemonSets(ds, renderReleaseImage, nil, kubeconfigPath); err != nil {
		logrus.Warningf("failed to apply ds %v: %v", ds, err)
		return err
	}
	return nil

}
