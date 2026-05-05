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
	ovnNodeDontSNATSubnets     = "k8s.ovn.org/node-ingress-snat-exclude-subnets"
	c2ccSNATTrackingAnnotation = "microshift.io/c2cc-snat-subnets"
)

type annotationManager struct {
	kubeClient   kubernetes.Interface
	nodeName     string
	desiredCIDRs []string
}

func newAnnotationManager(kubeClient kubernetes.Interface, nodeName string, remoteCIDRs []string) *annotationManager {
	sorted := make([]string, len(remoteCIDRs))
	copy(sorted, remoteCIDRs)
	sort.Strings(sorted)

	return &annotationManager{
		kubeClient:   kubeClient,
		nodeName:     nodeName,
		desiredCIDRs: sorted,
	}
}

func parseCIDRAnnotation(value string) []string {
	if value == "" {
		return nil
	}
	var cidrs []string
	if err := json.Unmarshal([]byte(value), &cidrs); err != nil {
		klog.Warningf("Failed to parse SNAT annotation value %q: %v", value, err)
		return nil
	}
	return cidrs
}

func buildAnnotationPatch(annotations map[string]interface{}) ([]byte, error) {
	patch := map[string]interface{}{
		"metadata": map[string]interface{}{
			"annotations": annotations,
		},
	}
	return json.Marshal(patch)
}

func cidrSetContainsAll(superset, subset []string) bool {
	set := make(map[string]bool, len(superset))
	for _, c := range superset {
		set[c] = true
	}
	for _, c := range subset {
		if !set[c] {
			return false
		}
	}
	return true
}

func (a *annotationManager) reconcile(ctx context.Context) error {
	node, err := a.kubeClient.CoreV1().Nodes().Get(ctx, a.nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node %q: %w", a.nodeName, err)
	}

	existing := parseCIDRAnnotation(node.Annotations[ovnNodeDontSNATSubnets])
	previous := parseCIDRAnnotation(node.Annotations[c2ccSNATTrackingAnnotation])

	// Target = (existing - previous) + desired
	// This replaces only the CIDRs C2CC previously wrote, preserving anything added by other components.
	foreignCIDRs := make(map[string]bool, len(existing))
	for _, c := range existing {
		foreignCIDRs[c] = true
	}
	for _, c := range previous {
		delete(foreignCIDRs, c)
	}

	targetSet := make(map[string]bool, len(foreignCIDRs)+len(a.desiredCIDRs))
	for c := range foreignCIDRs {
		targetSet[c] = true
	}
	for _, c := range a.desiredCIDRs {
		targetSet[c] = true
	}

	target := make([]string, 0, len(targetSet))
	for c := range targetSet {
		target = append(target, c)
	}
	sort.Strings(target)

	targetJSON, _ := json.Marshal(target)
	desiredJSON, _ := json.Marshal(a.desiredCIDRs)

	if node.Annotations[ovnNodeDontSNATSubnets] == string(targetJSON) &&
		node.Annotations[c2ccSNATTrackingAnnotation] == string(desiredJSON) {
		return nil
	}

	patch, err := buildAnnotationPatch(map[string]interface{}{
		ovnNodeDontSNATSubnets:     string(targetJSON),
		c2ccSNATTrackingAnnotation: string(desiredJSON),
	})
	if err != nil {
		return fmt.Errorf("failed to marshal annotation patch: %w", err)
	}
	_, err = a.kubeClient.CoreV1().Nodes().Patch(ctx, a.nodeName,
		types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to patch node annotation: %w", err)
	}
	klog.V(2).Infof("Updated node annotation %s = %s (tracking: %s)", ovnNodeDontSNATSubnets, string(targetJSON), string(desiredJSON))
	return nil
}

func (a *annotationManager) cleanup(ctx context.Context) error {
	node, err := a.kubeClient.CoreV1().Nodes().Get(ctx, a.nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node %q for cleanup: %w", a.nodeName, err)
	}

	tracked := parseCIDRAnnotation(node.Annotations[c2ccSNATTrackingAnnotation])
	if len(tracked) == 0 {
		return nil
	}

	existing := parseCIDRAnnotation(node.Annotations[ovnNodeDontSNATSubnets])
	trackedSet := make(map[string]bool, len(tracked))
	for _, c := range tracked {
		trackedSet[c] = true
	}

	var remaining []string
	for _, c := range existing {
		if !trackedSet[c] {
			remaining = append(remaining, c)
		}
	}

	var snatValue interface{}
	if len(remaining) > 0 {
		sort.Strings(remaining)
		data, _ := json.Marshal(remaining)
		snatValue = string(data)
	}

	patch, err := buildAnnotationPatch(map[string]interface{}{
		ovnNodeDontSNATSubnets:     snatValue,
		c2ccSNATTrackingAnnotation: nil,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal cleanup patch: %w", err)
	}
	_, err = a.kubeClient.CoreV1().Nodes().Patch(ctx, a.nodeName,
		types.MergePatchType, patch, metav1.PatchOptions{})
	if err != nil {
		return fmt.Errorf("failed to cleanup node annotation: %w", err)
	}
	klog.V(2).Infof("Cleaned up node annotation %s (removed %d C2CC CIDRs, %d remaining)", ovnNodeDontSNATSubnets, len(tracked), len(remaining))
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
				current := parseCIDRAnnotation(node.Annotations[ovnNodeDontSNATSubnets])
				if cidrSetContainsAll(current, a.desiredCIDRs) {
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
