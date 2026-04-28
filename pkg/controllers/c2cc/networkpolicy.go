package c2cc

import (
	"context"
	"fmt"
	"net"

	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

const (
	c2ccNetworkPolicyName      = "c2cc-allow-remote-pods"
	c2ccNetworkPolicyNamespace = "default"
	c2ccManagedByLabel         = "app.kubernetes.io/managed-by"
	c2ccManagedByValue         = "microshift-c2cc"
)

type networkPolicyManager struct {
	kubeClient kubernetes.Interface
	desired    *networkingv1.NetworkPolicy
}

func newNetworkPolicyManager(kubeClient kubernetes.Interface, remotePodCIDRs []*net.IPNet) *networkPolicyManager {
	var ingressPeers []networkingv1.NetworkPolicyPeer
	for _, cidr := range remotePodCIDRs {
		ingressPeers = append(ingressPeers, networkingv1.NetworkPolicyPeer{
			IPBlock: &networkingv1.IPBlock{
				CIDR: cidr.String(),
			},
		})
	}

	policy := &networkingv1.NetworkPolicy{
		ObjectMeta: metav1.ObjectMeta{
			Name:      c2ccNetworkPolicyName,
			Namespace: c2ccNetworkPolicyNamespace,
			Labels: map[string]string{
				c2ccManagedByLabel: c2ccManagedByValue,
			},
		},
		Spec: networkingv1.NetworkPolicySpec{
			PodSelector: metav1.LabelSelector{},
			Ingress: []networkingv1.NetworkPolicyIngressRule{
				{From: ingressPeers},
			},
			PolicyTypes: []networkingv1.PolicyType{
				networkingv1.PolicyTypeIngress,
			},
		},
	}

	return &networkPolicyManager{
		kubeClient: kubeClient,
		desired:    policy,
	}
}

func (m *networkPolicyManager) reconcile(ctx context.Context) error {
	client := m.kubeClient.NetworkingV1().NetworkPolicies(c2ccNetworkPolicyNamespace)

	existing, err := client.Get(ctx, c2ccNetworkPolicyName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		_, err = client.Create(ctx, m.desired, metav1.CreateOptions{})
		if err != nil {
			return fmt.Errorf("creating NetworkPolicy: %w", err)
		}
		klog.V(2).Infof("Created NetworkPolicy %s/%s", c2ccNetworkPolicyNamespace, c2ccNetworkPolicyName)
		return nil
	}
	if err != nil {
		return fmt.Errorf("getting NetworkPolicy: %w", err)
	}

	toUpdate := m.desired.DeepCopy()
	toUpdate.ResourceVersion = existing.ResourceVersion
	_, err = client.Update(ctx, toUpdate, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("updating NetworkPolicy: %w", err)
	}
	return nil
}

func (m *networkPolicyManager) cleanup(ctx context.Context) error {
	err := m.kubeClient.NetworkingV1().NetworkPolicies(c2ccNetworkPolicyNamespace).Delete(
		ctx, c2ccNetworkPolicyName, metav1.DeleteOptions{})
	if errors.IsNotFound(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("deleting NetworkPolicy: %w", err)
	}
	klog.V(2).Infof("Deleted NetworkPolicy %s/%s", c2ccNetworkPolicyNamespace, c2ccNetworkPolicyName)
	return nil
}
