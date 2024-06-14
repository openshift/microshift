package components

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"k8s.io/klog/v2"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/config/lvmd"
)

var lvmdConfigPathInMicroShift string
var lvmdConfigPathInMicroShiftSync sync.Once

func lvmdConfigForMicroShift() string {
	lvmdConfigPathInMicroShiftSync.Do(func() {
		lvmdConfigPathInMicroShift = filepath.Join(filepath.Dir(config.ConfigFile), lvmd.LvmdConfigFileName)
	})
	return lvmdConfigPathInMicroShift
}

// loadCSIPluginConfig searches for a user-defined lvmd configuration file in /etc/microshift. If found, returns
// the lvmd config.  If not found, returns a default-value lvmd config which is saved in /etc/microshift.
func loadCSIPluginConfig() (*lvmd.Lvmd, error) {
	microshiftPath := lvmdConfigForMicroShift()
	_, err := os.Stat(microshiftPath)

	var config *lvmd.Lvmd
	if errors.Is(err, os.ErrNotExist) {
		if config, err = lvmd.DefaultLvmdConfig(); err != nil {
			return nil, err
		}
		if err := lvmd.SaveLvmdConfigToFile(config, microshiftPath); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	} else {
		config, err = lvmd.NewLvmdConfigFromFile(microshiftPath)
		if err != nil {
			return nil, err
		}
	}

	return config, nil
}

func startCSIPlugin(ctx context.Context, cfg *config.Config, kubeconfigPath string) error {
	if err := lvmd.LvmPresentOnMachine(); err != nil {
		klog.Warningf("skipping CSI deployment: %v", err)
		return nil
	}

	// the lvmd file should be located in the same directory as the microshift config to minimize coupling with the
	// csi plugin.
	lvmdCfg, err := loadCSIPluginConfig()
	if err != nil {
		return err
	}
	if !lvmdCfg.IsEnabled() {
		klog.V(2).Infof("CSI is disabled. %s", lvmdCfg.Message)
		return nil
	}

	var (
		// CRDS are handled in
		ns = []string{
			"components/lvms/topolvm-openshift-storage_namespace.yaml",
		}
		sa = []string{
			"components/lvms/lvms-operator_v1_serviceaccount.yaml",
			"components/lvms/vg-manager_v1_serviceaccount.yaml",
		}
		role = []string{
			"components/lvms/lvms-operator_rbac.authorization.k8s.io_v1_role.yaml",
			"components/lvms/vg-manager_rbac.authorization.k8s.io_v1_role.yaml",
			"components/lvms/lvms-metrics_rbac.authorization.k8s.io_v1_role.yaml",
		}
		rb = []string{
			"components/lvms/lvms-operator_rbac.authorization.k8s.io_v1_rolebinding.yaml",
			"components/lvms/vg-manager_rbac.authorization.k8s.io_v1_rolebinding.yaml",
			"components/lvms/lvms-metrics_rbac.authorization.k8s.io_v1_rolebinding.yaml",
		}
		cr = []string{
			"components/lvms/lvms-operator_rbac.authorization.k8s.io_v1_clusterrole.yaml",
			"components/lvms/vg-manager_rbac.authorization.k8s.io_v1_clusterrole.yaml",
		}
		crb = []string{
			"components/lvms/lvms-operator_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml",
			"components/lvms/vg-manager_rbac.authorization.k8s.io_v1_clusterrolebinding.yaml",
		}
		cm_ver = "components/lvms/topolvm-configmap_lvms-version.yaml"
		deploy = []string{
			"components/lvms/lvms-operator_apps_v1_deployment.yaml",
		}
		lvmClusters = []string{
			"components/lvms/lvms_default-lvmcluster.yaml",
		}
		sc = []string{
			"components/lvms/topolvm_default-storage-class.yaml",
		}
		svc = []string{
			"components/lvms/lvms-webhook-service_v1_service.yaml",
			"components/lvms/lvms-operator-metrics-service_v1_service.yaml",
			"components/lvms/vg-manager-metrics-service_v1_service.yaml",
		}
		vwc = []string{
			"components/lvms/lvms-operator_admissionregistration.k8s.io_v1_webhook.yaml",
		}
	)

	if err := assets.ApplyStorageClasses(ctx, sc, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply storage cass %v: %v", sc, err)
		return err
	}
	if err := assets.ApplyNamespaces(ctx, ns, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply ns %v: %v", ns, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(ctx, sa, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply sa %v: %v", sa, err)
		return err
	}
	if err := assets.ApplyRoles(ctx, role, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply role %v: %v", cr, err)
		return err
	}
	if err := assets.ApplyRoleBindings(ctx, rb, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply rolebinding %v: %v", cr, err)
		return err
	}
	if err := assets.ApplyClusterRoles(ctx, cr, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterrole %v: %v", cr, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(ctx, crb, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterrolebinding %v: %v", crb, err)
		return err
	}
	if err := assets.ApplyConfigMaps(ctx, []string{cm_ver}, nil, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply configMap %v: %v", crb, err)
		return err
	}
	if err := assets.ApplyServices(ctx, svc, nil, map[string]interface{}{}, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply service %v: %v", svc, err)
		return err
	}
	if err := assets.ApplyDeployments(ctx, deploy, renderTemplate, renderParamsFromConfig(cfg, nil), kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply deployment %v: %v", deploy, err)
		return err
	}
	if err := assets.ApplyValidatingWebhookConfiguration(ctx, vwc, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply validating webhook configuration %v: %v", vwc, err)
		return err
	}
	if err := assets.ApplyGeneric(ctx, lvmClusters, nil, map[string]interface{}{}, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply lvmcluster: %v", err)
		return err
	}

	return nil
}
