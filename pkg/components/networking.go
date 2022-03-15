package components

import (
	"github.com/openshift/microshift/pkg/assets"
	"k8s.io/klog/v2"
)

func startFlannel(kubeconfigPath string) error {
	var (
		// psp = []string{
		// 	"assets/components/flannel/podsecuritypolicy.yaml",
		// }
		cr = []string{
			"assets/components/flannel/clusterrole.yaml",
		}
		crb = []string{
			"assets/components/flannel/clusterrolebinding.yaml",
		}
		sa = []string{
			"assets/components/flannel/service-account.yaml",
		}
		cm = []string{
			"assets/components/flannel/configmap.yaml",
		}
		ds = []string{
			"assets/components/flannel/daemonset.yaml",
		}
	)

	if err := assets.ApplyClusterRoles(cr, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRole %v %v", cr, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(crb, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRoleBinding %v %v", crb, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(sa, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply serviceAccount %v %v", sa, err)
		return err
	}
	if err := assets.ApplyConfigMaps(cm, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply configMap %v %v", cm, err)
		return err
	}
	if err := assets.ApplyDaemonSets(ds, renderReleaseImage, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply daemonSet %v %v", ds, err)
		return err
	}
	return nil

}
