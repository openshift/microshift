package components

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	embedded "github.com/openshift/microshift/assets"
	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/config/ovn"
	"github.com/vishvananda/netlink"
	"k8s.io/klog/v2"
)

func startCNIPlugin(ctx context.Context, cfg *config.Config, kubeconfigPath string) error {
	if !cfg.Network.IsEnabled() {
		klog.Warningf("CNI deployment disabled, OVN-K will not be available")
		return nil
	}
	var (
		ns = []string{
			"components/ovn/common/namespace.yaml",
		}
		sa = []string{
			"components/ovn/common/master-serviceaccount.yaml",
			"components/ovn/common/node-serviceaccount.yaml",
		}
		r = []string{
			"components/ovn/common/role.yaml",
		}
		rb = []string{
			"components/ovn/common/rolebinding.yaml",
		}
		cr = []string{
			"components/ovn/common/clusterrole.yaml",
		}
		crb = []string{
			"components/ovn/common/clusterrolebinding.yaml",
		}
		cm = []string{
			"components/ovn/common/configmap.yaml",
		}
		apps = []string{
			"components/ovn/single-node/master/daemonset.yaml",
			"components/ovn/single-node/node/daemonset.yaml",
		}
	)

	if cfg.MultiNode.Enabled {
		apps = []string{
			"components/ovn/multi-node/master/daemonset.yaml",
			"components/ovn/multi-node/node/daemonset.yaml",
		}
	}

	ipFamily := netlink.FAMILY_ALL
	if cfg.IsIPv4() && !cfg.IsIPv6() {
		ipFamily = netlink.FAMILY_V4
	}

	if cfg.IsIPv6() && !cfg.IsIPv4() {
		ipFamily = netlink.FAMILY_V6
	}

	ovnConfig, err := ovn.NewOVNKubernetesConfigFromFileOrDefault(filepath.Dir(config.ConfigFile), cfg.MultiNode.Enabled, ipFamily)
	if err != nil {
		return fmt.Errorf("failed to create OVN-K configuration from %q: %w", config.ConfigFile, err)
	}

	if err := ovnConfig.Validate(); err != nil {
		return fmt.Errorf("failed to validate ovn-kubernetes configuration: %w", err)
	}

	if err := assets.ApplyNamespaces(ctx, ns, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply ns %v: %v", ns, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(ctx, sa, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply serviceAccount %v %v", sa, err)
		return err
	}
	if err := assets.ApplyRoles(ctx, r, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply role %v: %v", r, err)
		return err
	}
	if err := assets.ApplyRoleBindings(ctx, rb, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply rolebinding %v: %v", rb, err)
		return err
	}
	if err := assets.ApplyClusterRoles(ctx, cr, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRole %v %v", cr, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(ctx, crb, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply clusterRoleBinding %v %v", crb, err)
		return err
	}

	// Multinode only params: OVN_NB_DB_LIST, OVN_SB_DB_LIST, OVN_NB_PORT, OVN_SB_PORT
	extraParams := assets.RenderParams{
		"OVNConfig":      ovnConfig,
		"KubeconfigPath": kubeconfigPath,
		"KubeconfigDir":  filepath.Join(config.DataDir, "/resources/kubeadmin"),
		"OVN_NB_DB_LIST": fmt.Sprintf("tcp:%s:%s", cfg.MultiNode.Controlplane, ovn.OVN_NB_PORT),
		"OVN_SB_DB_LIST": fmt.Sprintf("tcp:%s:%s", cfg.MultiNode.Controlplane, ovn.OVN_SB_PORT),
		"OVN_NB_PORT":    ovn.OVN_NB_PORT,
		"OVN_SB_PORT":    ovn.OVN_SB_PORT,
	}
	if err := assets.ApplyConfigMaps(ctx, cm, renderTemplate, renderParamsFromConfig(cfg, extraParams), kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply configMap %v %v", cm, err)
		return err
	}
	if err := assets.ApplyDaemonSets(ctx, apps, renderTemplate, renderParamsFromConfig(cfg, extraParams), kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply apps %v %v", apps, err)
		return err
	}
	return nil
}

func deployMultus(ctx context.Context, cfg *config.Config, kubeconfigPath string) error {
	if !cfg.Network.IsMultusEnabled() {
		klog.Warningf("Multus CNI is disabled. Uninstall is not supported if it was installed previously.")
		return nil
	}

	var (
		ns = []string{
			"components/multus/00-namespace.yaml",
		}
		crd = []string{
			"components/multus/01-crd-networkattachmentdefinition.yaml",
		}
		sa = []string{
			"components/multus/02-service-account.yaml",
		}
		cr = []string{
			"components/multus/03-cluster-role.yaml",
		}
		crb = []string{
			"components/multus/04-cluster-role-binding.yaml",
		}
		cm = []string{
			"components/multus/05-configmap.yaml",
		}
		ds = []string{
			"components/multus/06-daemonset.yaml",
			"components/multus/07-daemonset-dhcp.yaml",
		}
	)

	if err := assets.ApplyNamespaces(ctx, ns, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply ns %v: %v", ns, err)
		return err
	}
	if err := assets.ApplyCRDAndWaitForEstablish(ctx, crd, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply ns %v: %v", ns, err)
		return err
	}
	if err := assets.ApplyServiceAccounts(ctx, sa, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply serviceAccount %v %v", sa, err)
		return err
	}
	if err := assets.ApplyClusterRoles(ctx, cr, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply cluster role %v: %v", cr, err)
		return err
	}
	if err := assets.ApplyClusterRoleBindings(ctx, crb, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply rolebinding %v: %v", crb, err)
		return err
	}
	if err := assets.ApplyConfigMaps(ctx, cm, renderTemplate, nil, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply configMap %v %v", cm, err)
		return err
	}

	params, err := getMultusRenderParams()
	if err != nil {
		return fmt.Errorf("error creating Multus render params: %v", err)
	}

	if err := assets.ApplyDaemonSets(ctx, ds, renderTemplate, params, kubeconfigPath); err != nil {
		klog.Warningf("Failed to apply daemonset %v %v", ds, err)
		return err
	}

	return nil
}

func getMultusRenderParams() (assets.RenderParams, error) {
	arch := strings.NewReplacer("amd64", "x86_64", "arm64", "aarch64").Replace(runtime.GOARCH)
	releaseInfoPath := fmt.Sprintf("components/multus/release-multus-%s.json", arch)
	releaseInfo, err := embedded.Asset(releaseInfoPath)
	if err != nil {
		return nil, fmt.Errorf("error getting asset %s: %v", releaseInfoPath, err)
	}

	var release map[string]any
	if err := json.Unmarshal(releaseInfo, &release); err != nil {
		return nil, fmt.Errorf("unmarshaling %s: %v", releaseInfoPath, err)
	}

	images, ok := release["images"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("multus release info does not contain 'images' field")
	}
	imageMultus, ok := images["multus-cni-microshift"].(string)
	if !ok {
		return nil, fmt.Errorf("multus release info does not contain 'multus-cni-microshift' image")
	}
	imagePlugins, ok := images["containernetworking-plugins-microshift"].(string)
	if !ok {
		return nil, fmt.Errorf("multus release info does not contain 'containernetworking-plugins-microshift' image")
	}

	params := assets.RenderParams{
		"MultusCNIImage":                  imageMultus,
		"ContainerNetworkingPluginsImage": imagePlugins,
	}

	return params, nil
}
