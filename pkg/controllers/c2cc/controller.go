package c2cc

import (
	"context"
	"fmt"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	reconcileInterval = 10 * time.Second
)

type C2CCRouteManager struct {
	cfg        *config.Config
	nodeName   string
	kubeconfig string
	kubeClient kubernetes.Interface
}

func NewC2CCRouteManager(cfg *config.Config) *C2CCRouteManager {
	return &C2CCRouteManager{
		cfg:        cfg,
		nodeName:   cfg.Node.HostnameOverride,
		kubeconfig: cfg.KubeConfigPath(config.KubeAdmin),
	}
}

func (c *C2CCRouteManager) Name() string           { return "c2cc-route-manager" }
func (c *C2CCRouteManager) Dependencies() []string { return []string{"kubelet"} }

func (c *C2CCRouteManager) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	if !c.cfg.C2CC.IsEnabled() {
		klog.Infof("C2CC is disabled")
		c.cleanupAll(ctx)
		close(ready)
		return ctx.Err()
	}

	klog.Infof("C2CC is enabled with %d remote cluster(s)", len(c.cfg.C2CC.RemoteClusters))

	if err := c.initKubeClient(); err != nil {
		close(ready)
		return fmt.Errorf("create kube client: %w", err)
	}

	close(ready)
	klog.Infof("Ready, starting reconciliation loop")

	ticker := time.NewTicker(reconcileInterval)
	defer ticker.Stop()

	c.fullReconcile(ctx)

	for {
		select {
		case <-ctx.Done():
			klog.Infof("Shutting down, routes preserved")
			return ctx.Err()
		case <-ticker.C:
			klog.V(4).Infof("Periodic resync")
			c.fullReconcile(ctx)
		}
	}
}

func (c *C2CCRouteManager) initKubeClient() error {
	restCfg, err := clientcmd.BuildConfigFromFlags("", c.kubeconfig)
	if err != nil {
		return fmt.Errorf("build kubeconfig: %w", err)
	}
	client, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}
	c.kubeClient = client
	return nil
}

func (c *C2CCRouteManager) fullReconcile(ctx context.Context) {
	var errs []error
	for i := range c.cfg.C2CC.RemoteClusters {
		rc := &c.cfg.C2CC.RemoteClusters[i]
		if err := c.reconcileRemote(ctx, rc); err != nil {
			errs = append(errs, fmt.Errorf("remote %d (%s): %w", i, rc.NextHop, err))
		}
	}
	for _, err := range errs {
		klog.Errorf("Reconciliation failed: %v", err)
	}
}

func (c *C2CCRouteManager) reconcileRemote(ctx context.Context, rc *config.RemoteCluster) error {
	return nil
}

func (c *C2CCRouteManager) cleanupAll(ctx context.Context) {
	klog.V(2).Infof("Cleaning up any leftover C2CC state")
}
