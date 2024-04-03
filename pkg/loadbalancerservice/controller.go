package loadbalancerservice

import (
	"context"
	"fmt"
	"net"
	"slices"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/cloud-provider/service/helpers"
	"k8s.io/klog/v2"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/servicemanager"
	"github.com/vishvananda/netlink"
)

const defaultInformerResyncPeriod = 10 * time.Minute

type LoadbalancerServiceController struct {
	IPAddresses []string
	NICNames    []string
	NodeIP      string
	KubeConfig  string
	client      *kubernetes.Clientset
	indexer     cache.Indexer
	queue       workqueue.RateLimitingInterface
	informer    cache.SharedIndexInformer
}

var _ servicemanager.Service = &LoadbalancerServiceController{}

func NewLoadbalancerServiceController(cfg *config.Config) *LoadbalancerServiceController {
	ipAddresses := make([]string, 0)
	nicNames := make([]string, 0)
	for _, entry := range cfg.Ingress.ListenAddress {
		if net.ParseIP(entry) != nil {
			ipAddresses = append(ipAddresses, entry)
		} else {
			nicNames = append(nicNames, entry)
		}
	}
	return &LoadbalancerServiceController{
		IPAddresses: ipAddresses,
		NICNames:    nicNames,
		NodeIP:      cfg.Node.NodeIP,
		KubeConfig:  cfg.KubeConfigPath(config.KubeAdmin),
	}
}

func (c *LoadbalancerServiceController) Name() string {
	return "microshift-loadbalancer-service-controller"
}

func (c *LoadbalancerServiceController) Dependencies() []string {
	return []string{
		"network-configuration",
		"kube-apiserver",                  // needed for informers
		"infrastructure-services-manager", // starts CNI
	}
}

func (c *LoadbalancerServiceController) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)
	stopCh := make(chan struct{})
	defer close(stopCh)

	restCfg, err := c.restConfig()
	if err != nil {
		return fmt.Errorf("failed to create rest config for service controller: %w", err)
	}
	c.client, err = kubernetes.NewForConfig(restCfg)
	if err != nil {
		return fmt.Errorf("failed to create clientset for service controller: %w", err)
	}

	klog.Infof("Starting service controller")

	factory := informers.NewSharedInformerFactory(c.client, defaultInformerResyncPeriod)
	serviceInformer := factory.Core().V1().Services()
	c.informer = serviceInformer.Informer()
	c.queue = workqueue.NewRateLimitingQueue(workqueue.DefaultControllerRateLimiter())
	c.indexer = c.informer.GetIndexer()
	_, err = c.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(key)
			}
		},
		UpdateFunc: func(oldObj interface{}, newObj interface{}) {
			key, err := cache.MetaNamespaceKeyFunc(newObj)
			if err == nil {
				c.queue.Add(key)
			}
		},
		DeleteFunc: func(obj interface{}) {
			key, err := cache.DeletionHandlingMetaNamespaceKeyFunc(obj)
			if err == nil {
				c.queue.Add(key)
			}
		},
	})
	if err != nil {
		return fmt.Errorf("failed to initialize informer event handlers: %w", err)
	}

	factory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, c.informer.HasSynced) {
		return fmt.Errorf("timed out waiting for caches to sync")
	}

	go wait.Until(c.runWorker, time.Second, stopCh)

	close(ready)

	<-ctx.Done()

	return ctx.Err()
}

func (c *LoadbalancerServiceController) runWorker() {
	for c.processNextItem() {
	}
}

func (c *LoadbalancerServiceController) processNextItem() bool {
	key, quit := c.queue.Get()
	if quit {
		return false
	}
	defer c.queue.Done(key)

	err := c.updateServiceStatus(key.(string))
	c.handleErr(err, key)
	return true
}

// handleErr checks if an error happened and makes sure we will retry later.
func (c *LoadbalancerServiceController) handleErr(err error, key interface{}) {
	if err == nil {
		c.queue.Forget(key)
		return
	}

	klog.Infof("Error syncing service %v: %v", key, err)

	// Re-enqueue the key rate limited. Based on the rate limiter on the
	// queue and the re-enqueue history, the key will be processed later again.
	c.queue.AddRateLimited(key)
}

func (c *LoadbalancerServiceController) updateServiceStatus(key string) error {
	obj, exists, err := c.indexer.GetByKey(key)
	if err != nil {
		klog.Errorf("Fetching service object with key %s from store failed with %v", key, err)
		return err
	}

	if !exists {
		klog.Infof("Service %s does not exist anymore", key)
	} else {
		svc := obj.(*corev1.Service)
		if svc.Spec.Type != corev1.ServiceTypeLoadBalancer || svc.Spec.LoadBalancerClass != nil {
			return nil
		}
		klog.Infof("Process service %s/%s", svc.Namespace, svc.Name)

		newStatus, err := c.getNewStatus(svc)
		if err != nil {
			return err
		}
		err = c.patchStatus(svc, newStatus)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *LoadbalancerServiceController) restConfig() (*rest.Config, error) {
	return clientcmd.BuildConfigFromFlags("", c.KubeConfig)
}

func (c *LoadbalancerServiceController) getNewStatus(svc *corev1.Service) (*corev1.LoadBalancerStatus, error) {
	newStatus := &corev1.LoadBalancerStatus{}
	objs := c.indexer.List()
	for _, obj := range objs {
		s := obj.(*corev1.Service)
		if (s.Name == svc.Name && s.Namespace == svc.Namespace) || len(s.Status.LoadBalancer.Ingress) == 0 {
			continue
		}
		for _, ep := range s.Spec.Ports {
			for _, np := range svc.Spec.Ports {
				if ep.Port == np.Port {
					klog.Infof("Node port %d occupied", ep.Port)
					return newStatus, fmt.Errorf("node port %d occupied", ep.Port)
				}
			}
		}
	}

	//TODO use annotations instead.
	if svc.Name == "router-default" {
		ips, err := c.getRouterIPAddressList()
		if err != nil {
			return newStatus, fmt.Errorf("unable to update router IP list: %v", err)
		}
		for _, ip := range ips {
			newStatus.Ingress = append(newStatus.Ingress, corev1.LoadBalancerIngress{
				IP: ip,
			})
		}
	} else {
		newStatus.Ingress = append(newStatus.Ingress, corev1.LoadBalancerIngress{
			IP: c.NodeIP,
		})
	}
	return newStatus, nil
}

func (c *LoadbalancerServiceController) patchStatus(svc *corev1.Service, newStatus *corev1.LoadBalancerStatus) error {
	if helpers.LoadBalancerStatusEqual(&svc.Status.LoadBalancer, newStatus) {
		return nil
	}
	updated := svc.DeepCopy()
	updated.Status.LoadBalancer = *newStatus
	_, err := helpers.PatchService(c.client.CoreV1(), svc, updated)

	return err
}

func (c *LoadbalancerServiceController) getRouterIPAddressList() ([]string, error) {
	allowedAddresses, err := config.AllowedListeningIPAddresses()
	if err != nil {
		return nil, err
	}

	allowedNicNames, err := config.AllowedNICNames()
	if err != nil {
		return nil, err
	}

	ipList := make([]string, 0)

	for _, ip := range c.IPAddresses {
		if !slices.Contains(allowedAddresses, ip) {
			klog.Warningf("IP address %v not found in the host. Skipping", ip)
			continue
		}
		ipList = append(ipList, ip)
	}

	for _, nicName := range c.NICNames {
		if !slices.Contains(allowedNicNames, nicName) {
			klog.Warningf("NIC %v not found in the host. Skipping", nicName)
			continue
		}
		nicAddresses, err := ipAddressesFromNIC(nicName)
		if err != nil {
			return nil, err
		}
		for _, nicAddress := range nicAddresses {
			if !slices.Contains(allowedAddresses, nicAddress) {
				klog.Warningf("IP address %v from NIC %v is not allowed. Skipping", nicAddress, nicName)
				continue
			}
			ipList = append(ipList, nicAddress)
		}
	}

	return ipList, nil
}

func ipAddressesFromNIC(name string) ([]string, error) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, err
	}

	addrList, err := netlink.AddrList(link, netlink.FAMILY_ALL)
	if err != nil {
		return nil, err
	}

	ipList := make([]string, 0)

	for _, addr := range addrList {
		ipList = append(ipList, addr.IP.String())
	}

	return ipList, nil
}
