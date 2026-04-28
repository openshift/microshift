package c2cc

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/openshift/microshift/pkg/config"
	"github.com/ovn-kubernetes/libovsdb/client"
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
	ovn        *ovnRouteManager
	annotation *annotationManager
	nftMgr     *nftablesManager
	routes     *linuxRouteManager
	svcRoutes  *serviceRouteManager
	netpol     *networkPolicyManager
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
		c.initForCleanup(ctx)
		c.cleanupAll(ctx)
		close(ready)
		return ctx.Err()
	}

	klog.Infof("C2CC is enabled with %d remote cluster(s)", len(c.cfg.C2CC.RemoteClusters))

	if err := c.initKubeClient(); err != nil {
		close(ready)
		return fmt.Errorf("create kube client: %w", err)
	}

	nbClient, err := connectOVNNB(ctx)
	if err != nil {
		close(ready)
		return fmt.Errorf("connect OVN NB: %w", err)
	}
	defer nbClient.Close()

	if err := c.initSubsystems(nbClient); err != nil {
		close(ready)
		return fmt.Errorf("init subsystems: %w", err)
	}

	reconcileCh := make(chan string, 10)

	c.ovn.subscribe(ctx, reconcileCh)

	if routeDone, err := c.routes.subscribe(reconcileCh, "linux-route-change"); err != nil {
		klog.Warningf("Could not subscribe to route events for table %d: %v", c2ccRouteTable, err)
	} else {
		defer close(routeDone)
	}

	if svcRouteDone, err := c.svcRoutes.subscribe(reconcileCh, "service-route-change"); err != nil {
		klog.Warningf("Could not subscribe to route events for table %d: %v", c2ccSvcRouteTable, err)
	} else {
		defer close(svcRouteDone)
	}

	if nftClose, err := c.nftMgr.subscribe(reconcileCh); err != nil {
		klog.Warningf("Could not subscribe to nftables events: %v", err)
	} else {
		defer nftClose()
	}

	c.annotation.subscribe(ctx, reconcileCh)

	close(ready)
	klog.Infof("Ready, starting reconciliation loop")

	c.fullReconcile(ctx)

	ticker := time.NewTicker(reconcileInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			klog.Infof("Shutting down, routes preserved")
			return ctx.Err()
		case <-ticker.C:
			klog.V(4).Infof("Periodic resync")
			c.fullReconcile(ctx)
		case reason := <-reconcileCh:
			klog.V(2).Infof("Event-triggered reconcile: %s", reason)
			c.fullReconcile(ctx)
		}
	}
}

func (c *C2CCRouteManager) initKubeClient() error {
	restCfg, err := clientcmd.BuildConfigFromFlags("", c.kubeconfig)
	if err != nil {
		return fmt.Errorf("build kubeconfig: %w", err)
	}
	kClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("create kubernetes client: %w", err)
	}
	c.kubeClient = kClient
	return nil
}

func (c *C2CCRouteManager) initSubsystems(nbClient client.Client) error {
	c.ovn = newOVNRouteManager(nbClient, c.nodeName, c.cfg.C2CC.Resolved)
	c.annotation = newAnnotationManager(c.kubeClient, c.nodeName, c.cfg.C2CC.AllRemoteCIDRs())
	c.routes = newLinuxRouteManager(c.cfg)
	c.svcRoutes = newServiceRouteManager(c.cfg)

	var remotePodCIDRs []*net.IPNet
	var allRemoteCIDRs []*net.IPNet
	for _, rc := range c.cfg.C2CC.Resolved {
		remotePodCIDRs = append(remotePodCIDRs, rc.ClusterNetwork...)
		allRemoteCIDRs = append(allRemoteCIDRs, rc.ClusterNetwork...)
		allRemoteCIDRs = append(allRemoteCIDRs, rc.ServiceNetwork...)
	}

	nftMgr, err := newNftablesManager(allRemoteCIDRs)
	if err != nil {
		return fmt.Errorf("init nftables manager: %w", err)
	}
	c.nftMgr = nftMgr

	c.netpol = newNetworkPolicyManager(c.kubeClient, remotePodCIDRs)

	return nil
}

func (c *C2CCRouteManager) initForCleanup(ctx context.Context) {
	_ = c.initKubeClient()

	cleanupCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	if nbClient, err := connectOVNNB(cleanupCtx); err == nil {
		c.ovn = newOVNRouteManager(nbClient, c.nodeName, nil)
	} else {
		klog.Warningf("Could not connect to OVN NB for cleanup, OVN routes will not be removed: %v", err)
	}

	c.routes = newLinuxRouteManager(c.cfg)
	c.svcRoutes = newServiceRouteManager(c.cfg)

	if nftMgr, err := newNftablesManager(nil); err == nil {
		c.nftMgr = nftMgr
	}

	if c.kubeClient != nil {
		c.annotation = newAnnotationManager(c.kubeClient, c.nodeName, nil)
		c.netpol = newNetworkPolicyManager(c.kubeClient, nil)
	}
}

func (c *C2CCRouteManager) fullReconcile(ctx context.Context) {
	subsystems := []struct {
		name string
		fn   func(context.Context) error
	}{
		{"ovn-routes", c.ovn.reconcile},
		{"node-annotation", c.annotation.reconcile},
		{"linux-routes", c.routes.reconcile},
		{"service-routes", c.svcRoutes.reconcile},
		{"nftables", c.nftMgr.reconcile},
		{"network-policy", c.netpol.reconcile},
	}
	for _, s := range subsystems {
		if err := s.fn(ctx); err != nil {
			klog.Errorf("Reconcile %s failed: %v", s.name, err)
		}
	}
}

func (c *C2CCRouteManager) cleanupAll(ctx context.Context) {
	klog.V(2).Infof("Cleaning up any leftover C2CC state")

	type cleanable struct {
		name string
		fn   func(context.Context) error
	}

	var cleanups []cleanable

	if c.ovn != nil {
		cleanups = append(cleanups, cleanable{"ovn-routes", c.ovn.cleanup})
	}
	if c.annotation != nil {
		cleanups = append(cleanups, cleanable{"node-annotation", c.annotation.cleanup})
	}
	if c.routes != nil {
		cleanups = append(cleanups, cleanable{"linux-routes", c.routes.cleanup})
	}
	if c.svcRoutes != nil {
		cleanups = append(cleanups, cleanable{"service-routes", c.svcRoutes.cleanup})
	}
	if c.nftMgr != nil {
		cleanups = append(cleanups, cleanable{"nftables", c.nftMgr.cleanup})
	}
	if c.netpol != nil {
		cleanups = append(cleanups, cleanable{"network-policy", c.netpol.cleanup})
	}

	for _, cl := range cleanups {
		if err := cl.fn(ctx); err != nil {
			klog.Errorf("Cleanup %s failed: %v", cl.name, err)
		}
	}
}
