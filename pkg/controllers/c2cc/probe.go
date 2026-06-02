package c2cc

import (
	"context"
	"fmt"
	"net/http"
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
		client: msClient,
		probes: make(map[string]context.CancelFunc),
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
			if ok1 && ok2 && (oldRC.Spec.ProbeTarget != newRC.Spec.ProbeTarget ||
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
	client microshiftclientset.Interface
	mu     sync.Mutex
	probes map[string]context.CancelFunc
}

func (pm *probeManager) startProbe(ctx context.Context, rc *microshiftv1alpha1.RemoteCluster) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if _, exists := pm.probes[rc.Name]; exists {
		return
	}

	probeCtx, cancel := context.WithCancel(ctx)
	pm.probes[rc.Name] = cancel

	klog.Infof("Starting probe for %q (target=%s, interval=%s)",
		rc.Name, rc.Spec.ProbeTarget, rc.Spec.ProbeInterval.Duration)
	go pm.runProbeLoop(probeCtx, rc.Name, rc.Spec.ProbeTarget, rc.Spec.ProbeInterval.Duration)
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
		klog.Infof("Stopped probe for %q", name)
	}
}

func (pm *probeManager) stopAll() {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for name, cancel := range pm.probes {
		cancel()
		delete(pm.probes, name)
	}
}

func (pm *probeManager) runProbeLoop(ctx context.Context, name, target string, interval time.Duration) {
	httpClient := &http.Client{Timeout: probeHTTPTimeout}
	consecutiveFailures := 0
	url := "http://" + target + "/"

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			probeErr := doProbe(ctx, httpClient, url)
			now := metav1.Now()

			status := microshiftv1alpha1.RemoteClusterStatus{
				LastProbeTime: &now,
			}

			if probeErr != nil {
				consecutiveFailures++
				klog.V(2).Infof("Probe %q failed (%d consecutive): %v", name, consecutiveFailures, probeErr)

				if consecutiveFailures >= unhealthyThreshold {
					status.State = "Unhealthy"
				} else {
					status.State = "Healthy"
				}
				status.Errors = []string{probeErr.Error()}
			} else {
				consecutiveFailures = 0
				status.State = "Healthy"
				status.LastSuccessfulProbe = &now
			}

			if err := pm.updateStatus(ctx, name, status); err != nil {
				klog.Errorf("Failed to update status for %q: %v", name, err)
			}
		}
	}
}

func doProbe(ctx context.Context, client *http.Client, url string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	resp, err := client.Do(req) // #nosec G704 -- URL built from trusted RemoteCluster CR spec
	if err != nil {
		return fmt.Errorf("failed to execute probe request: %w", err)
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			klog.Errorf("Failed to close probe response body: %v", err)
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed with unexpected status %d", resp.StatusCode)
	}
	return nil
}

func (pm *probeManager) updateStatus(ctx context.Context, name string, status microshiftv1alpha1.RemoteClusterStatus) error {
	rcClient := pm.client.MicroshiftV1alpha1().RemoteClusters()

	return retry.RetryOnConflict(retry.DefaultBackoff, func() error {
		rc, err := rcClient.Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get RemoteCluster %q: %w", name, err)
		}

		// Preserve LastSuccessfulProbe from the existing status if this probe failed
		if rc.Status.LastSuccessfulProbe != nil && status.LastSuccessfulProbe == nil {
			status.LastSuccessfulProbe = rc.Status.LastSuccessfulProbe
		}

		rc.Status = status
		_, err = rcClient.UpdateStatus(ctx, rc, metav1.UpdateOptions{})
		return err
	})
}
