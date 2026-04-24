package c2cc

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	ovnNodeDontSNATSubnets = "k8s.ovn.org/node-ingress-snat-exclude-subnets"
)

type annotationManager struct {
	kubeClient        kubernetes.Interface
	nodeName          string
	desiredAnnotation string
}

func newAnnotationManager(kubeClient kubernetes.Interface, nodeName string, remoteCIDRs []string) *annotationManager {
	sorted := make([]string, len(remoteCIDRs))
	copy(sorted, remoteCIDRs)
	sort.Strings(sorted)
	data, _ := json.Marshal(sorted)

	return &annotationManager{
		kubeClient:        kubeClient,
		nodeName:          nodeName,
		desiredAnnotation: string(data),
	}
}

func (a *annotationManager) reconcile(ctx context.Context) error {
	node, err := a.kubeClient.CoreV1().Nodes().Get(ctx, a.nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get node %q: %w", a.nodeName, err)
	}

	if node.Annotations[ovnNodeDontSNATSubnets] == a.desiredAnnotation {
		return nil
	}

	patch := fmt.Sprintf(`{"metadata":{"annotations":{%q:%q}}}`, ovnNodeDontSNATSubnets, a.desiredAnnotation)
	_, err = a.kubeClient.CoreV1().Nodes().Patch(ctx, a.nodeName,
		types.MergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("patch node annotation: %w", err)
	}
	klog.V(2).Infof("Updated node annotation %s = %s", ovnNodeDontSNATSubnets, a.desiredAnnotation)
	return nil
}

func (a *annotationManager) cleanup(ctx context.Context) error {
	patch := fmt.Sprintf(`{"metadata":{"annotations":{%q:null}}}`, ovnNodeDontSNATSubnets)
	_, err := a.kubeClient.CoreV1().Nodes().Patch(ctx, a.nodeName,
		types.MergePatchType, []byte(patch), metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("remove node annotation: %w", err)
	}
	klog.V(2).Infof("Removed node annotation %s", ovnNodeDontSNATSubnets)
	return nil
}
