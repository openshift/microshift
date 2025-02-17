package components

import (
	"context"
	"fmt"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
)

func startCSISnapshotController(ctx context.Context, cfg *config.Config, kubeconfigPath string) error {
	var (
		cr = []string{
			"components/csi-snapshot-controller/clusterrole.yaml",
			"components/csi-snapshot-controller/webhook_clusterrole.yaml",
		}
		crb = []string{
			"components/csi-snapshot-controller/clusterrolebinding.yaml",
			"components/csi-snapshot-controller/webhook_clusterrolebinding.yaml",
		}
		sa = []string{
			"components/csi-snapshot-controller/serviceaccount.yaml",
			"components/csi-snapshot-controller/webhook_serviceaccount.yaml",
		}
		svc        = []string{"components/csi-snapshot-controller/webhook_service.yaml"}
		webhookCfg = []string{"components/csi-snapshot-controller/webhook_config.yaml"}
		deploy     = []string{
			"components/csi-snapshot-controller/csi_controller_deployment.yaml",
			"components/csi-snapshot-controller/webhook_deployment.yaml",
		}
	)

	if err := assets.ApplyClusterRoles(ctx, cr, kubeconfigPath); err != nil {
		return fmt.Errorf("apply clusterRole: %w", err)
	}
	if err := assets.ApplyClusterRoleBindings(ctx, crb, kubeconfigPath); err != nil {
		return fmt.Errorf("apply clusterRoleBinding: %w", err)
	}
	if err := assets.ApplyServiceAccounts(ctx, sa, kubeconfigPath); err != nil {
		return fmt.Errorf("apply service account: %w", err)
	}
	if err := assets.ApplyServices(ctx, svc, nil, map[string]interface{}{}, kubeconfigPath); err != nil {
		return fmt.Errorf("apply service: %w", err)
	}
	if err := assets.ApplyValidatingWebhookConfiguration(ctx, webhookCfg, kubeconfigPath); err != nil {
		return fmt.Errorf("apply validationWebhookConfiguration: %w", err)
	}
	if err := assets.ApplyDeployments(ctx, deploy, renderTemplate, renderParamsFromConfig(cfg, nil), kubeconfigPath); err != nil {
		return fmt.Errorf("apply deployments: %w", err)
	}

	return nil
}
