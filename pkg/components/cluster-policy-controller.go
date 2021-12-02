package components

import (
	"github.com/openshift/microshift/pkg/assets"
	"github.com/sirupsen/logrus"
)

func startClusterPolicyController(kubeconfigPath string) error {
	var (
		cr = []string{
			"assets/rbac/0000_10_cluster-policy-controller_clusterrole.yaml",
		}
		crb = []string{
			"assets/rbac/0000_10_cluster-policy-controller_clusterrolebinding.yaml",
		}
		sa = []string{
			"assets/core/0000_10_namespace-security-allocation-controller_sa.yaml",
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
	return nil

}
