package c2cc

import (
	"context"
	"fmt"
	"time"

	"github.com/openshift/microshift/pkg/assets"
	"github.com/openshift/microshift/pkg/config"
	microshiftclient "github.com/openshift/microshift/pkg/generated/clientset/versioned/typed/microshift/v1alpha1"
	"github.com/ovn-kubernetes/libovsdb/client"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const (
	reconcileInterval = 2 * time.Second
)

var healthcheckCRD = []string{"crd/microshift.io_remoteclusters.yaml"}

type C2CCRouteManager struct {
	cfg        *config.Config
	nodeName   string
	kubeconfig string

	kubeClient       kubernetes.Interface
	microshiftClient microshiftclient.MicroshiftV1alpha1Interface
	ovn              *ovnRouteManager
	annotation       *annotationManager
	nftMgr           *nftablesManager
	routes           *linuxRouteManager
	svcRoutes        *serviceRouteManager
	healthcheck      *healthcheckCRManager
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
		klog.Infof("C2CC is disabled - attempting best effort cleanup")
		close(ready)
		closeCleanup := c.initForCleanup(ctx)
		defer closeCleanup()
		c.cleanupAll(ctx)
		return ctx.Err()
	}

	klog.Infof("C2CC is enabled with %d remote cluster(s)", len(c.cfg.C2CC.RemoteClusters))

	// Declaring ready even before init because many of the components it tries to communicate with are not up yet
	// and excessive waiting before readiness can cause them to never become ready resulting in MicroShift restart.
	close(ready)

	if err := c.initKubeClient(); err != nil {
		return fmt.Errorf("failed to create kube client: %w", err)
	}

	nbClient, err := connectOVNNB(ctx)
	if err != nil {
		return fmt.Errorf("failed to connect OVN NB: %w", err)
	}
	defer nbClient.Close()

	if err := c.initSubsystems(nbClient); err != nil {
		return fmt.Errorf("failed to init subsystems: %w", err)
	}

	reconcileCh := make(chan string, 50)

	c.ovn.subscribe(ctx, reconcileCh)

	if routeDone, err := c.routes.subscribe(reconcileCh, "linux-route-change"); err != nil {
		klog.Warningf("Could not subscribe to route events for table %d: %v", c.routes.table, err)
	} else {
		defer close(routeDone)
	}

	if svcRouteDone, err := c.svcRoutes.subscribe(reconcileCh, "service-route-change"); err != nil {
		klog.Warningf("Could not subscribe to route events for table %d: %v", c.svcRoutes.table, err)
	} else {
		defer close(svcRouteDone)
	}

	if nftClose, err := c.nftMgr.subscribe(ctx, reconcileCh); err != nil {
		klog.Warningf("Could not subscribe to nftables events: %v", err)
	} else {
		defer nftClose()
	}

	c.annotation.subscribe(ctx, reconcileCh)

	if err := assets.ApplyCRDAndWaitForEstablish(ctx, healthcheckCRD, c.kubeconfig); err != nil {
		return fmt.Errorf("failed to apply C2CC healthcheck CRD: %w", err)
	}

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
			// Drain the channel to debounce reconcile events and avoid queueing - each reconcile performs setup of all subsystems,
			// so it's not necessary to reconcile several times in a row.
			coalesced := 0
			for {
				select {
				case <-reconcileCh:
					coalesced++
				default:
					goto drained
				}
			}
		drained:
			if coalesced > 0 {
				klog.V(2).Infof("Event-triggered reconcile: %s (+%d coalesced)", reason, coalesced)
			} else {
				klog.V(2).Infof("Event-triggered reconcile: %s", reason)
			}
			c.fullReconcile(ctx)
		}
	}
}

func (c *C2CCRouteManager) initKubeClient() error {
	restCfg, err := clientcmd.BuildConfigFromFlags("", c.kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to build kubeconfig: %w", err)
	}
	kClient, err := kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}
	c.kubeClient = kClient

	msClient, err := microshiftclient.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("failed to create microshift client: %w", err)
	}
	c.microshiftClient = msClient

	return nil
}

func (c *C2CCRouteManager) initSubsystems(nbClient client.Client) error {
	c.ovn = newOVNRouteManager(nbClient, c.nodeName, c.cfg.C2CC.Resolved)
	c.annotation = newAnnotationManager(c.kubeClient, c.nodeName, c.cfg.C2CC.AllRemoteCIDRStrings())
	c.routes = newLinuxRouteManager(c.cfg)
	c.svcRoutes = newServiceRouteManager(c.cfg)

	nftMgr, err := newNftablesManager(c.cfg.C2CC.ResolvedAllCIDRs)
	if err != nil {
		return fmt.Errorf("failed to init nftables manager: %w", err)
	}
	c.nftMgr = nftMgr
	c.healthcheck = newHealthcheckCRManager(c.microshiftClient, c.cfg)

	return nil
}

func (c *C2CCRouteManager) initForCleanup(ctx context.Context) func() {
	if err := c.initKubeClient(); err != nil {
		klog.Warningf("Could not init kube client for cleanup, Kubernetes C2CC state will not be removed: %v", err)
	}

	var closers []func()

	cleanupCtx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()
	if nbClient, err := connectOVNNB(cleanupCtx); err == nil {
		c.ovn = newOVNRouteManager(nbClient, c.nodeName, nil)
		closers = append(closers, func() { nbClient.Close() })
	} else {
		klog.Warningf("Could not connect to OVN NB for cleanup, OVN routes will not be removed: %v", err)
	}

	c.routes = newLinuxRouteManager(c.cfg)
	c.svcRoutes = newServiceRouteManager(c.cfg)

	if nftMgr, err := newNftablesManager(nil); err == nil {
		c.nftMgr = nftMgr
	} else {
		klog.Warningf("Could not init nftables manager for cleanup: %v", err)
	}

	if c.kubeClient != nil {
		c.annotation = newAnnotationManager(c.kubeClient, c.nodeName, nil)
	}

	return func() {
		for _, fn := range closers {
			fn()
		}
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
		{"healthcheck-crs", c.healthcheck.reconcile},
		{"probe-deployment", c.deployProbe},
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
	cleanups = append(cleanups, cleanable{"probe-namespace", func(ctx context.Context) error {
		return assets.DeleteNamespaces(ctx, c2ccNamespace, c.kubeconfig)
	}})
	cleanups = append(cleanups, cleanable{"probe-clusterrolebinding", func(ctx context.Context) error {
		return assets.DeleteClusterRoleBindings(ctx, c2ccClusterRoleBinding, c.kubeconfig)
	}})
	cleanups = append(cleanups, cleanable{"probe-clusterrole", func(ctx context.Context) error {
		return assets.DeleteClusterRoles(ctx, c2ccClusterRole, c.kubeconfig)
	}})
	cleanups = append(cleanups, cleanable{"healthcheck-crd", func(ctx context.Context) error {
		return assets.DeleteCRDs(ctx, healthcheckCRD, c.kubeconfig)
	}})

	for _, cl := range cleanups {
		if err := cl.fn(ctx); err != nil {
			klog.Errorf("Cleanup %s failed: %v", cl.name, err)
		}
	}
}
