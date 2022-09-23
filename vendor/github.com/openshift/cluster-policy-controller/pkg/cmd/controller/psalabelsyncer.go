package controller

import (
	"context"

	"github.com/openshift/cluster-policy-controller/pkg/psalabelsyncer"
)

func runPodSecurityAdmissionLabelSynchronizationController(ctx context.Context, controllerCtx *EnhancedControllerContext) (bool, error) {

	kubeClient, err := controllerCtx.ClientBuilder.Client(podSecurityAdmissionLabelSyncerControllerServiceAccountName)
	if err != nil {
		return true, err
	}

	controller, err := psalabelsyncer.NewPodSecurityAdmissionLabelSynchronizationController(
		kubeClient.CoreV1().Namespaces(),
		controllerCtx.KubernetesInformers.Core().V1().Namespaces(),
		controllerCtx.KubernetesInformers.Rbac().V1(),
		controllerCtx.KubernetesInformers.Core().V1().ServiceAccounts(),
		controllerCtx.SecurityInformers.Security().V1().SecurityContextConstraints(),
		controllerCtx.EventRecorder.ForComponent("podsecurity-admission-label-sync-controller"),
	)

	if err != nil {
		return true, err
	}

	go controller.Run(ctx, 1)
	return true, nil
}
