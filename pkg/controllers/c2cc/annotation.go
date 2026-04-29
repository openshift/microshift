package c2cc

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/watch"
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

func (a *annotationManager) subscribe(ctx context.Context, reconcileCh chan<- string) {
	go func() {
		for {
			watcher, err := a.kubeClient.CoreV1().Nodes().Watch(ctx, metav1.ListOptions{
				FieldSelector: "metadata.name=" + a.nodeName,
			})
			if err != nil {
				klog.Warningf("Could not watch node for annotation changes: %v", err)
				select {
				case <-ctx.Done():
					return
				case <-time.After(10 * time.Second):
					continue
				}
			}
			for event := range watcher.ResultChan() {
				if event.Type != watch.Modified {
					continue
				}
				node, ok := event.Object.(*corev1.Node)
				if !ok {
					continue
				}
				if node.Annotations[ovnNodeDontSNATSubnets] == a.desiredAnnotation {
					continue
				}
				select {
				case reconcileCh <- "node-annotation-changed":
				default:
				}
			}
			watcher.Stop()
			if ctx.Err() != nil {
				return
			}
			klog.V(4).Infof("Node watch closed unexpectedly, reconnecting")
			select {
			case <-ctx.Done():
				return
			case <-time.After(1 * time.Second):
			}
		}
	}()
	klog.V(2).Infof("Subscribed to node annotation changes for %s", a.nodeName)
}
