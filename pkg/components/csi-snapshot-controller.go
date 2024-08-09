package components

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/klog/v2"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
)

func startCSISnapshotController(ctx context.Context, cfg *config.Config, kubeconfigPath string) error {
	csiComps := sets.New[config.OptionalCsiComponent](cfg.Storage.OptionalCSIComponents...)
	if len(cfg.Storage.OptionalCSIComponents) == 0 {
		// Upgraded clusters will not have set .storage.*, so we need to support the prior default behavior and deploy
		// CSI snapshots when .storage.optionalCsiComponents is nil.
		csiComps.Insert(config.CsiComponentSnapshot, config.CsiComponentSnapshotWebhook)
	} else if csiComps.Has(config.CsiComponentNone) {
		// User set a zero-len slice, indicating that the cluster supports optional CSI components, and that they should
		// be disabled.
		klog.Warningf("additional (non-driver) csi components have been disabled by user")
		return nil
	}

	var (
		whSA     = []string{"components/csi-snapshot-controller/webhook_serviceaccount.yaml"}
		whCfg    = []string{"components/csi-snapshot-controller/webhook_config.yaml"}
		whDeploy = []string{"components/csi-snapshot-controller/webhook_deployment.yaml"}
		whSvc    = []string{"components/csi-snapshot-controller/webhook_service.yaml"}
		cr       = []string{"components/csi-snapshot-controller/clusterrole.yaml"}
		crb      = []string{"components/csi-snapshot-controller/clusterrolebinding.yaml"}
		sa       = []string{"components/csi-snapshot-controller/serviceaccount.yaml"}
		deploy   = []string{"components/csi-snapshot-controller/csi_controller_deployment.yaml"}
	)

	// Deploy Webhook
	//nolint:nestif
	if csiComps.Has(config.CsiComponentSnapshotWebhook) {
		klog.Infof("deploying CSI snapshot webhook")
		if err := assets.ApplyServiceAccounts(ctx, whSA, kubeconfigPath); err != nil {
			return fmt.Errorf("apply service account: %w", err)
		}
		if err := assets.ApplyDeployments(ctx, whDeploy, renderTemplate, renderParamsFromConfig(cfg, nil), kubeconfigPath); err != nil {
			return fmt.Errorf("apply deployment: %w", err)
		}
		if err := assets.ApplyValidatingWebhookConfiguration(ctx, whCfg, kubeconfigPath); err != nil {
			return fmt.Errorf("apply validationWebhookConfiguration: %w", err)
		}
		if err := assets.ApplyDeployments(ctx, deploy, renderTemplate, renderParamsFromConfig(cfg, nil), kubeconfigPath); err != nil {
			return fmt.Errorf("apply deployments: %w", err)
		}
	} else {
		klog.Warningf("CSI snapshot webhook is disabled")
	}

	// Deploy CSI Controller Deployment
	//nolint:nestif
	if csiComps.Has(config.CsiComponentSnapshot) || csiComps.Len() == 0 {
		klog.Infof("deploying CSI snapshot controller")
		if err := assets.ApplyClusterRoles(ctx, cr, kubeconfigPath); err != nil {
			return fmt.Errorf("apply clusterRole: %w", err)
		}
		if err := assets.ApplyClusterRoleBindings(ctx, crb, kubeconfigPath); err != nil {
			return fmt.Errorf("apply clusterRoleBinding: %w", err)
		}
		if err := assets.ApplyServiceAccounts(ctx, sa, kubeconfigPath); err != nil {
			return fmt.Errorf("apply service account: %w", err)
		}
		if err := assets.ApplyServices(ctx, whSvc, nil, map[string]interface{}{}, kubeconfigPath); err != nil {
			return fmt.Errorf("apply service: %w", err)
		}
	} else {
		klog.Warningf("CSI snapshot controller is disabled")
	}
	return nil
}
