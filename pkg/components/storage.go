package components

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	sccv1 "github.com/openshift/api/security/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/config/lvmd"
)

func startCSIPlugin(ctx context.Context, cfg *config.Config, kubeconfigPath string) error {
	if !cfg.Storage.IsEnabled() {
		klog.Warningf("CSI driver deployment disabled, persistent storage will not be available")
		return nil
	}
	if err := lvmd.LvmPresentOnMachine(); err != nil {
		klog.Warningf("skipping CSI deployment: %v", err)
		return nil
	}

	usrCfg := filepath.Join(filepath.Dir(config.ConfigFile), lvmd.LvmdConfigFileName)
	runtimeCfg := lvmd.RuntimeLvmdConfigFile

	lvmdCfg, err := loadCSIPluginConfig(
		ctx,
		usrCfg,
		runtimeCfg,
	)
	if err != nil {
		return err
	}
	if !lvmdCfg.IsEnabled() {
		klog.V(2).Infof("CSI is disabled. %s", lvmdCfg.Message)
		return nil
	}

	if err := deleteLegacyResources(ctx, kubeconfigPath); err != nil {
		klog.Warningf("Failed to delete legacy resources of CSI installation, possibly corrupt state: %v", err)
		return err
	}

	// This sets up the resources in the cluster
	// Note that currently the lvmdCfg is only checked for being enabled at startup
	// That means that the loading of the configuration file only respects enabled / disabled on restarts.
	// That means that
	// 1. If the configuration file is changed to disable the CSI plugin, the plugin will still be running until the next restart and until all resources are removed
	// 2. If the configuration file is changed to enable the CSI plugin, the plugin will not be running until the next restart
	// TODO: Implement a mechanism to reload the configuration file even when enabling or disabling, this requires a mechanism in k8s to delete resources that are no longer wanted
	// 3. If the plugin is already enabled, the configuration file is changed and the plugin is restarted with hot-reloading.
	return setupPluginResources(ctx, cfg, kubeconfigPath)
}

// deleteLegacyResources deletes the legacy resources of TopoLVM.
// TODO: Remove this function in the next release.
func deleteLegacyResources(ctx context.Context, kubeconfigPath string) error {
	klog.Infof("Deleting legacy resources of TopoLVM (this is a migration operation and will disappear in the next release)")
	return assets.DeleteGeneric(ctx, []runtime.Object{
		&appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "topolvm-controller", Namespace: "openshift-storage"}},
		&appsv1.DaemonSet{ObjectMeta: metav1.ObjectMeta{Name: "topolvm-node", Namespace: "openshift-storage"}},
		&rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "topolvm-node-scc"}},
		&rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "topolvm-node"}},
		&rbacv1.ClusterRoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "topolvm-controller"}},
		&rbacv1.RoleBinding{ObjectMeta: metav1.ObjectMeta{Name: "topolvm-controller", Namespace: "openshift-storage"}},
		&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "topolvm-controller"}},
		&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "topolvm-node-scc"}},
		&rbacv1.ClusterRole{ObjectMeta: metav1.ObjectMeta{Name: "topolvm-node"}},
		&rbacv1.Role{ObjectMeta: metav1.ObjectMeta{Name: "topolvm-controller", Namespace: "openshift-storage"}},
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "topolvm-controller", Namespace: "openshift-storage"}},
		&corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: "topolvm-node", Namespace: "openshift-storage"}},
		&corev1.ConfigMap{ObjectMeta: metav1.ObjectMeta{Name: "lvmd", Namespace: "openshift-storage"}},
		&sccv1.SecurityContextConstraints{ObjectMeta: metav1.ObjectMeta{Name: "topolvm-node"}},
	}, kubeconfigPath)
}

func setupPluginResources(ctx context.Context, cfg *config.Config, kubeconfigPath string) error {
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

// loadCSIPluginConfig sets up a file watcher on the configuration directory. It first creates a cancellable context
// and determines the directory and file path of the configuration file. It then checks if the configuration directory
// and file exist. If the configuration file exists, it copies the user's configuration, otherwise, it copies the default configuration.
//
// The function then creates a new file watcher and adds the configuration directory to the watcher's watch list.
// It listens for file events in a separate goroutine. If the configuration file is modified, it copies the user's configuration again.
// If the configuration file is removed, it copies the default configuration. If the file permissions are changed, it logs a warning.
//
// The function is designed to ensure that lvms always uses the most recent configuration through its hot-reload mechanism,
// whether it's user-defined or the default, and reacts to changes in the configuration file in real-time.
func loadCSIPluginConfig(ctx context.Context,
	// usrCfg is the user's configuration file.
	usrCfg string,
	runtimeCfg string,
) (*lvmd.Lvmd, error) {
	usrCfgDir := filepath.Dir(usrCfg)
	ctx, cancel := context.WithCancelCause(ctx)

	// check if dir exists, otherwise the watcher errors
	fi, err := os.Stat(usrCfgDir)
	if err != nil {
		return nil, fmt.Errorf("config directory %q cannot be watched: %v", usrCfgDir, err)
	} else if !fi.IsDir() {
		return nil, fmt.Errorf("config directory %q is not a directory", usrCfgDir)
	}

	if _, err := os.Stat(usrCfg); err == nil {
		copyUserCfg(usrCfg, runtimeCfg, cancel)
	} else {
		copyDefaultCfg(runtimeCfg, cancel)
	}

	// Create new watcher.
	go func(ctx context.Context) {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			klog.Errorf("unable to set up file watcher: %v", err)
			return
		}
		defer func() {
			if err := watcher.Close(); err != nil {
				klog.Errorf("unable to close file watcher: %v", err)
			}
		}()
		if err = watcher.Add(usrCfgDir); err != nil {
			klog.Errorf("unable to add file path to watcher: %v", err)
			return
		}

		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Name != usrCfg {
					continue
				}
				if event.Has(fsnotify.Write) {
					klog.Infof("lvmd config file %q was modified, this may be due to changed user configuration or by mistake; "+
						"now, the new config will be applied from %q", event.Name, usrCfg)
					copyUserCfg(usrCfg, runtimeCfg, cancel)
				}
				if event.Has(fsnotify.Remove) {
					klog.Warningf("lvmd config file %q was removed, this may be due to a reset user configuration or by mistake; "+
						"now, the new config will be applied from inbuilt defaults", event.Name)
					copyDefaultCfg(runtimeCfg, cancel)
				}
				if event.Has(fsnotify.Chmod) {
					klog.Warningf("permissions were modified for %q, this may cause side-effects", event.Name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				if err != nil {
					klog.Errorf("file watcher error: %v", err)
					cancel(err)
				}
			case <-ctx.Done():
				klog.Errorf("stop watching for lvmd config changes: %v", ctx.Err())
				return
			}
		}
	}(ctx)

	return lvmd.NewLvmdConfigFromFile(runtimeCfg)
}

func copyUserCfg(userCfg string, runtimeCfg string, onFail context.CancelCauseFunc) {
	userCfgFile, err := os.OpenFile(userCfg, os.O_RDONLY, 0)
	if err != nil {
		onFail(fmt.Errorf("unable to open lvmd config file %q: %w", userCfg, err))
		return
	}
	defer func() {
		if err := userCfgFile.Close(); err != nil {
			onFail(fmt.Errorf("unable to close lvmd config file %q: %w", userCfg, err))
		}
	}()

	if err := os.MkdirAll(filepath.Dir(runtimeCfg), 0755); err != nil {
		onFail(fmt.Errorf("unable to create lvmd runtime config directory: %w", err))
		return
	}

	runtimeCfgFile, err := os.OpenFile(runtimeCfg, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		onFail(fmt.Errorf("unable to open lvmd runtime config file %q: %w", runtimeCfg, err))
		return
	}
	defer func() {
		if err := runtimeCfgFile.Close(); err != nil {
			onFail(fmt.Errorf("unable to close lvmd runtime config file %q: %w", runtimeCfg, err))
		}
	}()

	if _, err := io.Copy(runtimeCfgFile, userCfgFile); err != nil {
		onFail(fmt.Errorf("unable to copy lvmd config file %q to runtime config file %q: %w", userCfg, runtimeCfg, err))
		return
	}
}

// defaultCfgLoader is a function that loads the default lvmd configuration.
// Used for testing overrides as the default loader uses host lvm for volume group detection.
var defaultCfgLoader = lvmd.DefaultLvmdConfig

func copyDefaultCfg(runtimeCfg string, onFail context.CancelCauseFunc) {
	if err := os.MkdirAll(filepath.Dir(runtimeCfg), 0755); err != nil {
		onFail(fmt.Errorf("failed to ensure runtime lvmd config directory: %w", err))
		return
	}

	if cfg, err := defaultCfgLoader(); err != nil {
		onFail(fmt.Errorf("failed to load default lvmd config: %v", err))
	} else if err := lvmd.SaveLvmdConfigToFile(cfg, runtimeCfg); err != nil {
		onFail(fmt.Errorf("failed to save default lvmd config: %w", err))
	}
}
