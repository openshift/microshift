package bootstrap

import (
	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/sirupsen/logrus"
)

func ApplyBootstrapClusterRoleBindings(cfg *config.MicroshiftConfig, kubeconfigPath string) error {
	var (
		clusterRoleBinding = []string{
			"assets/rbac/0000_10_bootstrap-crb-creator.yaml",
			"assets/rbac/0000_10_bootstrap-crb-approver.yaml",
			"assets/rbac/0000_10_bootstrap-crb-renewal.yaml",
		}
	)

	if err := assets.ApplyClusterRoleBindings(clusterRoleBinding, kubeconfigPath); err != nil {
		logrus.Warningf("failed to apply clusterRolebinding %v: %v", clusterRoleBinding, err)
		return err
	}

	return nil
}
