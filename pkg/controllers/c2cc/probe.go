package c2cc

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"sync"
	"time"

	microshiftv1alpha1 "github.com/openshift/microshift/pkg/apis/microshift/v1alpha1"
	microshiftclientset "github.com/openshift/microshift/pkg/generated/clientset/versioned"
	microshiftinformers "github.com/openshift/microshift/pkg/generated/informers/externalversions"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/retry"
	"k8s.io/klog/v2"
)

const (
	unhealthyThreshold = 3
	probeHTTPTimeout   = 5 * time.Second
	informerResync     = 30 * time.Second

	stateHealthy   = "Healthy"
	stateUnhealthy = "Unhealthy"
)

// RunProbe is the entrypoint for the healthcheck-probe subcommand.
// It runs inside a pod on the cluster network, serving as both a probe
// target (HTTP :8080) and an active prober of remote clusters.
func RunProbe(ctx context.Context) error {
	restCfg, err := rest.InClusterConfig()
	if err != nil {
		return fmt.Errorf("failed to build in-cluster config: %w", err)
	}

	msClient, err := microshiftclientset.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("failed to create microshift client: %w", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		if _, err := fmt.Fprint(w, "ok"); err != nil {
			klog.Errorf("Failed to write probe response: %v", err)
		}
	})
	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		klog.Infof("Starting probe target HTTP server on :8080")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			klog.Errorf("Probe HTTP server error: %v", err)
		}
	}()

	pm := &probeManager{
		client:    msClient,
		probes:    make(map[string]context.CancelFunc),
		latencies: make(map[string]map[string]*latencyWindow),
	}

	factory := microshiftinformers.NewSharedInformerFactory(msClient, informerResync)
	informer := factory.Microshift().V1alpha1().RemoteClusters().Informer()

	if _, err := informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			if rc, ok := obj.(*microshiftv1alpha1.RemoteCluster); ok {
				pm.startProbe(ctx, rc)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			oldRC, ok1 := oldObj.(*microshiftv1alpha1.RemoteCluster)
			newRC, ok2 := newObj.(*microshiftv1alpha1.RemoteCluster)
			if ok1 && ok2 && (!slices.Equal(oldRC.Spec.ProbeTargets, newRC.Spec.ProbeTargets) ||
				oldRC.Spec.ProbeInterval != newRC.Spec.ProbeInterval) {
				pm.restartProbe(ctx, newRC)
			}
		},
		DeleteFunc: func(obj interface{}) {
			rc, ok := obj.(*microshiftv1alpha1.RemoteCluster)
			if !ok {
				if tombstone, ok := obj.(cache.DeletedFinalStateUnknown); ok {
					rc, _ = tombstone.Obj.(*microshiftv1alpha1.RemoteCluster)
				}
			}
			if rc != nil {
				pm.stopProbe(rc.Name)
			}
		},
	}); err != nil {
		return fmt.Errorf("failed to add RemoteCluster informer handlers: %w", err)
	}

	factory.Start(ctx.Done())
	factory.WaitForCacheSync(ctx.Done())
	klog.Infof("Probe manager running, watching RemoteCluster CRs")

	<-ctx.Done()
	pm.stopAll()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil { //nolint:contextcheck // parent ctx is already cancelled
		klog.Errorf("Probe HTTP server shutdown error: %v", err)
	}
	klog.Infof("Probe manager shut down")
	return nil
}

type probeManager struct {
	client    microshiftclientset.Interface
	mu        sync.Mutex
	probes    map[string]context.CancelFunc        // keyed by CR name
	latencies map[string]map[string]*latencyWindow // CR name → target → latency window
}

func (pm *probeManager) startProbe(ctx context.Context, rc *microshiftv1alpha1.RemoteCluster) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.probes[rc.Name]; exists {
		return
	}

	probeCtx, cancel := context.WithCancel(ctx)
	pm.probes[rc.Name] = cancel
	pm.latencies[rc.Name] = make(map[string]*latencyWindow, len(rc.Spec.ProbeTargets))
	for _, target := range rc.Spec.ProbeTargets {
		pm.latencies[rc.Name][target] = &latencyWindow{}
	}

	klog.Infof("Starting probe for %q (targets=%v, interval=%s)",
		rc.Name, rc.Spec.ProbeTargets, rc.Spec.ProbeInterval.Duration)
	go pm.runProbeLoop(probeCtx, rc.Name, rc.Spec.ProbeTargets, rc.Spec.ProbeInterval.Duration, pm.latencies[rc.Name])
}

func (pm *probeManager) restartProbe(ctx context.Context, rc *microshiftv1alpha1.RemoteCluster) {
	pm.stopProbe(rc.Name)
	pm.startProbe(ctx, rc)
}

func (pm *probeManager) stopProbe(name string) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if cancel, exists := pm.probes[name]; exists {
		cancel()
		delete(pm.probes, name)
		delete(pm.latencies, name)
		klog.Infof("Stopped probe for %q", name)
	}
}

func (pm *probeManager) stopAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for name, cancel := range pm.probes {
		cancel()
		delete(pm.probes, name)
		delete(pm.latencies, name)
	}
}

func (pm *probeManager) runProbeLoop(ctx context.Context, name string, targets []string, interval time.Duration, windows map[string]*latencyWindow) {
	httpClient := &http.Client{Timeout: probeHTTPTimeout}
	consecutiveFailures := make(map[string]int, len(targets))

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			now := metav1.Now()
			targetResults := pm.probeAllTargets(ctx, name, targets, httpClient, consecutiveFailures, windows)
			state := pm.deriveAggregateState(targetResults)
			status := pm.buildStatus(state, &now, targetResults)

			if err := pm.updateStatus(ctx, name, status); err != nil {
				klog.Errorf("Failed to update status for %q: %v", name, err)
			}
		}
	}
}

func (pm *probeManager) probeAllTargets(ctx context.Context, name string, targets []string, httpClient *http.Client,
	consecutiveFailures map[string]int, windows map[string]*latencyWindow) []microshiftv1alpha1.TargetResult {
	targetResults := make([]microshiftv1alpha1.TargetResult, 0, len(targets))
	for _, target := range targets {
		url := "http://" + target + "/"
		rtt, probeErr := doProbe(ctx, httpClient, url)

		result := microshiftv1alpha1.TargetResult{Target: target}

		if probeErr != nil {
			consecutiveFailures[target]++
			klog.V(2).Infof("Probe %q target %s failed (%d consecutive): %v",
				name, target, consecutiveFailures[target], probeErr)
			result.State = pm.deriveTargetState(consecutiveFailures[target])
			result.Error = probeErr.Error()
		} else {
			consecutiveFailures[target] = 0
			result.State = stateHealthy
			windows[target].add(rtt)
		}

		result.Latency = windows[target].stats()
		targetResults = append(targetResults, result)
	}
	return targetResults
}

func (pm *probeManager) deriveTargetState(failures int) string {
	if failures >= unhealthyThreshold {
		return stateUnhealthy
	}
	return stateHealthy
}

func (pm *probeManager) deriveAggregateState(targetResults []microshiftv1alpha1.TargetResult) string {
	for _, tr := range targetResults {
		if tr.State == stateUnhealthy {
			return stateUnhealthy
		}
	}
	return stateHealthy
}

func (pm *probeManager) buildStatus(state string, now *metav1.Time, targetResults []microshiftv1alpha1.TargetResult) microshiftv1alpha1.RemoteClusterStatus {
	status := microshiftv1alpha1.RemoteClusterStatus{
		State:         state,
		LastProbeTime: now,
		TargetResults: targetResults,
	}

	for _, tr := range targetResults {
		if tr.Error != "" {
			status.Errors = append(status.Errors, fmt.Sprintf("%s: %s", tr.Target, tr.Error))
		} else {
			status.LastSuccessfulProbe = now
		}
	}

	return status
}

func doProbe(ctx context.Context, client *http.Client, url string) (time.Duration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to create request: %w", err)
	}
	start := time.Now()
	resp, err := client.Do(req) // #nosec G704 -- URL built from trusted RemoteCluster CR spec
	rtt := time.Since(start)
	if err != nil {
		return 0, fmt.Errorf("failed to execute probe request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			klog.Errorf("Failed to close probe response body: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed with unexpected status %d", resp.StatusCode)
	}
	return rtt, nil
}

func (pm *probeManager) updateStatus(ctx context.Context, name string, status microshiftv1alpha1.RemoteClusterStatus) error {
	rcClient := pm.client.MicroshiftV1alpha1().RemoteClusters()

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		rc, err := rcClient.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get RemoteCluster %q: %w", name, err)
		}

		// Preserve LastSuccessfulProbe from the existing status if no target succeeded this tick
		if rc.Status.LastSuccessfulProbe != nil && status.LastSuccessfulProbe == nil {
			status.LastSuccessfulProbe = rc.Status.LastSuccessfulProbe
		}

		// Preserve per-target latency from the existing CR when the in-memory
		// window is empty (e.g., after pod restart before samples accumulate).
		if len(rc.Status.TargetResults) > 0 {
			existing := make(map[string]*microshiftv1alpha1.LatencyStats, len(rc.Status.TargetResults))
			for i := range rc.Status.TargetResults {
				if rc.Status.TargetResults[i].Latency != nil {
					existing[rc.Status.TargetResults[i].Target] = rc.Status.TargetResults[i].Latency
				}
			}
			for i := range status.TargetResults {
				if status.TargetResults[i].Latency == nil {
					status.TargetResults[i].Latency = existing[status.TargetResults[i].Target]
				}
			}
		}

		rc.Status = status
		_, err = rcClient.UpdateStatus(ctx, rc, metav1.UpdateOptions{})
		return err
	})
}
