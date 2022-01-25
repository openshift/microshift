package mdns

import (
	"context"
	"strings"
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

func (c *MicroShiftmDNSController) restConfig() (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", c.KubeConfig)
}

func (c *MicroShiftmDNSController) startRouteInformer(stopCh chan struct{}) error {
	klog.Infof("Starting MicroShift mDNS route watcher")
	cfg, err := c.restConfig()
	if err != nil {
		return errors.Wrap(err, "error creating rest config for route informer")
	}

	dc, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return errors.Wrap(err, "error creating dynamic informer")
	}

	return c.run(stopCh, dc)
}

func (c *MicroShiftmDNSController) run(stopCh chan struct{}, dc dynamic.Interface) error {
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

	return nil
}

// TODO: The need for this indicates that the openshift-default-scc-manager is declaring itself ready before
//       it really is. If we don't wait for the route API the informer will start throwing ugly errors.
func (c *MicroShiftmDNSController) waitForRouterAPI(dc dynamic.Interface, routersGVR *schema.GroupVersionResource) {
	backoff := wait.Backoff{
		Cap:      3 * time.Minute,
		Duration: 10 * time.Second,
		Factor:   1.5,
		Steps:    24,
	}

	klog.Infof("mDNS: waiting for route API to be ready")
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
		klog.Errorf("waiting for the route.openshift.io/v1 API to come up", err)
	} else {
		klog.Infof("mDNS: route API ready, watching routers")
	}
}

func (c *MicroShiftmDNSController) addedRoute(obj interface{}) {
	u := obj.(*unstructured.Unstructured)

	host, found, err := unstructured.NestedString(u.UnstructuredContent(), "spec", "host")
	if !found || err != nil {
		klog.Errorf("mDNS spec.host not found", err)
		return
	}

	c.exposeHost(host)
}

func (c *MicroShiftmDNSController) updatedRoute(oldObj, newObj interface{}) {
	oldU := oldObj.(*unstructured.Unstructured)

	oldHost, _, _ := unstructured.NestedString(oldU.UnstructuredContent(), "spec", "host")
	newU := newObj.(*unstructured.Unstructured)
	newHost, _, _ := unstructured.NestedString(newU.UnstructuredContent(), "spec", "host")

	if oldHost != newHost {
		klog.Infof("Updating route host", "oldHost", oldHost, "newHost", newHost)
		c.unexposeHost(oldHost)
		c.exposeHost(newHost)
	}
}

func (c *MicroShiftmDNSController) deletedRoute(obj interface{}) {
	u := obj.(*unstructured.Unstructured)
	host, found, err := unstructured.NestedString(u.UnstructuredContent(), "spec", "host")
	if !found || err != nil {
		klog.Errorf("mDNS spec.host not found", err)
		return
	}

	c.unexposeHost(host)
}

func (c *MicroShiftmDNSController) exposeHost(host string) {
	if !strings.HasSuffix(host, server.DefaultmDNSTLD) {
		klog.V(2).InfoS("mDNS ignoring host without mDNS suffix", "host", host)
		return
	}

	klog.Infof("mDNS: route found for", "host", host, "ips", c.myIPs)

	// TODO(multi-node) look up for the exact router service Endpoints instead of assuming our own IP (ok for single-node)
	c.incHost(host)
	c.resolver.AddDomain(host+".", c.myIPs)
}

func (c *MicroShiftmDNSController) unexposeHost(oldHost string) {
	if c.decHost(oldHost) == 0 {
		klog.Infof("mDNS removing, no more router references", "host", oldHost)
		c.resolver.DeleteDomain(oldHost + ".")
	}
}

func (c *MicroShiftmDNSController) incHost(name string) {
	c.Lock()
	defer c.Unlock()
	c.hostCount[name]++
}

func (c *MicroShiftmDNSController) decHost(name string) int {
	c.Lock()
	defer c.Unlock()
	if c.hostCount[name] > 0 {
		c.hostCount[name]--
	}
	return c.hostCount[name]
}
