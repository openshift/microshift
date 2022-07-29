package controller

import (
	"context"
	"fmt"

	"github.com/openshift/library-go/pkg/security/uid"

	sccallocation "github.com/openshift/cluster-policy-controller/pkg/security/controller"
	"github.com/openshift/cluster-policy-controller/pkg/security/mcs"
)

func RunNamespaceSecurityAllocationController(ctx context.Context, controllerCtx *EnhancedControllerContext) (bool, error) {
	uidRange, err := uid.ParseRange(controllerCtx.OpenshiftControllerConfig.SecurityAllocator.UIDAllocatorRange)
	if err != nil {
		return true, fmt.Errorf("unable to describe UID range: %v", err)
	}
	mcsRange, err := mcs.ParseRange(controllerCtx.OpenshiftControllerConfig.SecurityAllocator.MCSAllocatorRange)
	if err != nil {
		return true, fmt.Errorf("unable to describe MCS category range: %v", err)
	}

	kubeClient, err := controllerCtx.ClientBuilder.Client(infraNamespaceSecurityAllocationControllerServiceAccountName)
	if err != nil {
		return true, err
	}
	securityClient, err := controllerCtx.ClientBuilder.OpenshiftSecurityClient(infraNamespaceSecurityAllocationControllerServiceAccountName)
	if err != nil {
		return true, err
	}

	controller := sccallocation.NewNamespaceSCCAllocationController(
		controllerCtx.KubernetesInformers.Core().V1().Namespaces(),
		kubeClient.CoreV1().Namespaces(),
		securityClient.SecurityV1(),
		uidRange,
		sccallocation.DefaultMCSAllocation(uidRange, mcsRange, controllerCtx.OpenshiftControllerConfig.SecurityAllocator.MCSLabelsPerProject),
		controllerCtx.EventRecorder,
	)
	go controller.Run(ctx, 1)

	return true, nil
}
