package controller

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"

	"github.com/openshift/library-go/pkg/operator/csr"
)

const (
	controllerName                = "csr-approver-controller"
	monitoringNamespace           = "openshift-monitoring"
	monitoringRequesterSA         = "cluster-monitoring-operator"
	monitoringSubjectNameLabelKey = "metrics.openshift.io/csr.subject"
	prometheus                    = "prometheus"
	metricsServer                 = "metrics-server"
)

var (
	monitoringCertificateSubjects = []string{
		fmt.Sprintf("CN=system:serviceaccount:%s:%s-k8s", monitoringNamespace, prometheus),
		fmt.Sprintf("CN=system:serviceaccount:%s:%s", monitoringNamespace, metricsServer),
	}
)

func RunCSRApproverController(ctx context.Context, controllerCtx *EnhancedControllerContext) (bool, error) {
	kubeClient, err := controllerCtx.ClientBuilder.Client(infraClusterCSRApproverControllerServiceAccountName)
	if err != nil {
		return true, err
	}

	selector := labels.NewSelector()
	labelsRequirement, err := labels.NewRequirement(monitoringSubjectNameLabelKey, selection.In, []string{prometheus, metricsServer})
	if err != nil {
		return true, err
	}
	selector = selector.Add(*labelsRequirement)

	controller := csr.NewCSRApproverController(
		controllerName,
		nil,
		kubeClient.CertificatesV1().CertificateSigningRequests(),
		controllerCtx.KubernetesInformers.Certificates().V1().CertificateSigningRequests(),
		csr.NewLabelFilter(selector),
		csr.NewServiceAccountMultiSubjectsApprover(monitoringNamespace, monitoringRequesterSA, monitoringCertificateSubjects),
		controllerCtx.EventRecorder)

	go controller.Run(ctx, 1)

	return true, nil
}
