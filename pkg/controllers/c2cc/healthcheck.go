package c2cc

import (
	"context"
	"errors"
	"fmt"
	"net"
	"slices"
	"sort"
	"strings"

	microshiftv1alpha1 "github.com/openshift/microshift/pkg/apis/microshift/v1alpha1"
	"github.com/openshift/microshift/pkg/config"
	microshiftclient "github.com/openshift/microshift/pkg/generated/clientset/versioned/typed/microshift/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

const (
	probePort      = 8080
	managedByLabel = "app.kubernetes.io/managed-by"
	managerName    = "c2cc-route-manager"
)

type healthcheckCRManager struct {
	client microshiftclient.MicroshiftV1alpha1Interface
	cfg    *config.Config
}

func newHealthcheckCRManager(client microshiftclient.MicroshiftV1alpha1Interface, cfg *config.Config) *healthcheckCRManager {
	return &healthcheckCRManager{
		client: client,
		cfg:    cfg,
	}
}

func (h *healthcheckCRManager) reconcile(ctx context.Context) error {
	desired := h.buildDesiredCRs()

	existing, err := h.client.RemoteClusters().List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=%s", managedByLabel, managerName),
	})
	if err != nil {
		return fmt.Errorf("listing RemoteCluster CRs: %w", err)
	}

	existingByName := make(map[string]*microshiftv1alpha1.RemoteCluster, len(existing.Items))
	for i := range existing.Items {
		existingByName[existing.Items[i].Name] = &existing.Items[i]
	}

	var errs []error

	for name, want := range desired {
		got, ok := existingByName[name]
		if !ok {
			if _, err := h.client.RemoteClusters().Create(ctx, want, metav1.CreateOptions{}); err != nil {
				errs = append(errs, fmt.Errorf("creating RemoteCluster %q: %w", name, err))
			} else {
				klog.Infof("Created RemoteCluster CR %q", name)
			}
			continue
		}

		delete(existingByName, name)

		if slices.Equal(got.Spec.ProbeTargets, want.Spec.ProbeTargets) && got.Spec.ProbeInterval == want.Spec.ProbeInterval {
			continue
		}

		got.Spec = want.Spec
		if _, err := h.client.RemoteClusters().Update(ctx, got, metav1.UpdateOptions{}); err != nil {
			errs = append(errs, fmt.Errorf("updating RemoteCluster %q: %w", name, err))
		} else {
			klog.V(2).Infof("Updated RemoteCluster CR %q", name)
		}
	}

	for name := range existingByName {
		if err := h.client.RemoteClusters().Delete(ctx, name, metav1.DeleteOptions{}); err != nil {
			errs = append(errs, fmt.Errorf("deleting stale RemoteCluster %q: %w", name, err))
		} else {
			klog.Infof("Deleted stale RemoteCluster CR %q", name)
		}
	}

	return errors.Join(errs...)
}

func (h *healthcheckCRManager) buildDesiredCRs() map[string]*microshiftv1alpha1.RemoteCluster {
	desired := make(map[string]*microshiftv1alpha1.RemoteCluster, len(h.cfg.C2CC.Resolved))
	for _, rc := range h.cfg.C2CC.Resolved {
		name := crNameForRemote(rc.PrimaryNextHop())
		var targets []string
		for _, probeIP := range rc.ProbeIPs {
			targets = append(targets, net.JoinHostPort(probeIP, fmt.Sprintf("%d", probePort)))
		}
		sort.Strings(targets) // deterministic order for comparison
		desired[name] = &microshiftv1alpha1.RemoteCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name: name,
				Labels: map[string]string{
					managedByLabel: managerName,
				},
			},
			Spec: microshiftv1alpha1.RemoteClusterSpec{
				ProbeTargets:  targets,
				ProbeInterval: metav1.Duration{Duration: h.cfg.C2CC.ResolvedProbeInterval},
			},
		}
	}
	return desired
}

func crNameForRemote(nextHop net.IP) string {
	s := nextHop.String()
	s = strings.ReplaceAll(s, ".", "-")
	s = strings.ReplaceAll(s, ":", "-")
	return "c2cc-" + s
}
