package mdns

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/openshift/microshift/pkg/mdns/server"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

const defaultResyncTime = time.Hour * 1

type MicroShiftmDNSRouteController struct {
	sync.Mutex
	parent    *MicroShiftmDNSController
	hostCount map[string]int
}

func (s *MicroShiftmDNSRouteController) Name() string { return "microshift-mdns-route-controller" }
func (s *MicroShiftmDNSRouteController) Dependencies() []string {
	return []string{"openshift-api-components-manager"}
}

func (c *MicroShiftmDNSRouteController) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	stopCh := make(chan struct{})
	defer close(stopCh)
	defer close(stopped)

	go c.startRouteInformer(stopCh, ready)

	select {
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (c *MicroShiftmDNSRouteController) restConfig() (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", c.parent.KubeConfig)
}

func (c *MicroShiftmDNSRouteController) startRouteInformer(stopCh chan struct{}, ready chan<- struct{}) error {
	klog.InfoS("Starting MicroShift mDNS route watcher")
	cfg, err := c.restConfig()
	if err != nil {
		return errors.Wrap(err, "error creating rest config for route informer")
	}

	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error creating dynamic informer")
	}

	return c.runInformers(stopCh, dc, ready)
}

func (c *MicroShiftmDNSRouteController) runInformers(stopCh chan struct{}, dc dynamic.Interface, ready chan<- struct{}) error {
	informerFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dc, defaultResyncTime, v1.NamespaceAll, nil)
	routersGVR, _ := schema.ParseResourceArg("routes.v1.route.openshift.io")

	informer := informerFactory.ForResource(*routersGVR).Informer()
	c.waitForRouterAPI(dc, routersGVR)

	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc:    c.addedRoute,
		UpdateFunc: c.updatedRoute,
		DeleteFunc: c.deletedRoute,
	}

	informer.AddEventHandler(handlers)
	informer.Run(stopCh)
	close(ready)
	return nil
}

// TODO: The need for this indicates that the openshift-default-scc-manager is declaring itself ready before
//       it really is. If we don't wait for the route API the informer will start throwing ugly errors.
func (c *MicroShiftmDNSRouteController) waitForRouterAPI(dc dynamic.Interface, routersGVR *schema.GroupVersionResource) {
	backoff := wait.Backoff{
		Cap:      3 * time.Minute,
		Duration: 10 * time.Second,
		Factor:   1.5,
		Steps:    24,
	}

	klog.InfoS("mDNS: waiting for route API to be ready")
	err := wait.ExponentialBackoff(backoff, func() (bool, error) {
		_, err := dc.Resource(*routersGVR).List(context.TODO(), metav1.ListOptions{})
		if err == nil {
			return true, nil
		}
		return false, nil
	})

	if err != nil {
		// This is a soft error, the watcher down the line will keep waiting, but will emit
		// less nice errors
		klog.ErrorS(err, "waiting for the route.openshift.io/v1 API to come up")
	} else {
		klog.InfoS("mDNS: route API ready, watching routers")
	}
}

func (c *MicroShiftmDNSRouteController) addedRoute(obj interface{}) {
	u := obj.(*unstructured.Unstructured)

	host, found, err := unstructured.NestedString(u.UnstructuredContent(), "spec", "host")
	if !found || err != nil {
		klog.ErrorS(err, "mDNS spec.host not found")
		return
	}

	c.exposeHost(host)
}

func (c *MicroShiftmDNSRouteController) updatedRoute(oldObj, newObj interface{}) {
	oldU := oldObj.(*unstructured.Unstructured)

	oldHost, _, _ := unstructured.NestedString(oldU.UnstructuredContent(), "spec", "host")
	newU := newObj.(*unstructured.Unstructured)
	newHost, _, _ := unstructured.NestedString(newU.UnstructuredContent(), "spec", "host")

	if oldHost != newHost {
		klog.InfoS("Updating route host", "oldHost", oldHost, "newHost", newHost)
		c.unexposeHost(oldHost)
		c.exposeHost(newHost)
	}
}

func (c *MicroShiftmDNSRouteController) deletedRoute(obj interface{}) {
	u := obj.(*unstructured.Unstructured)
	host, found, err := unstructured.NestedString(u.UnstructuredContent(), "spec", "host")
	if !found || err != nil {
		klog.ErrorS(err, "mDNS spec.host not found")
		return
	}

	c.unexposeHost(host)
}

func (c *MicroShiftmDNSRouteController) exposeHost(host string) {
	if !strings.HasSuffix(host, server.DefaultmDNSTLD) {
		klog.V(2).InfoS("mDNS ignoring host without mDNS suffix", "host", host)
		return
	}

	klog.InfoS("mDNS: route found for", "host", host, "ips", c.parent.myIPs)

	// TODO(multi-node) look up for the exact router service Endpoints instead of assuming our own IP (ok for single-node)
	c.incHost(host)
	c.parent.resolver.AddDomain(host+".", c.parent.myIPs)
}

func (c *MicroShiftmDNSRouteController) unexposeHost(oldHost string) {
	if c.decHost(oldHost) == 0 {
		klog.InfoS("mDNS removing, no more router references", "host", oldHost)
		c.parent.resolver.DeleteDomain(oldHost + ".")
	}
}

func (c *MicroShiftmDNSRouteController) incHost(name string) {
	c.Lock()
	defer c.Unlock()
	c.hostCount[name]++
}

func (c *MicroShiftmDNSRouteController) decHost(name string) int {
	c.Lock()
	defer c.Unlock()
	if c.hostCount[name] > 0 {
		c.hostCount[name]--
	}
	return c.hostCount[name]
}
