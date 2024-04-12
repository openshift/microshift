package loadbalancerservice

import (
	"slices"

	"github.com/openshift/microshift/pkg/config"
	"github.com/vishvananda/netlink"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/klog/v2"
)

const (
	defaultRouterServiceName            = "router-default"
	defaultRouterServiceNamespace       = "openshift-ingress"
	defaultRouterServiceAnnotationKey   = "ingresscontroller.operator.openshift.io/owning-ingresscontroller"
	defaultRouterServiceAnnotationValue = "default"
)

type serviceUpdateFunction func([]string) error

func defaultRouterWatch(ipAddresses, nicNames []string, updateFunc serviceUpdateFunction, stopCh <-chan struct{}) {
	for {
		ips, err := defaultRouterListenAddresses(ipAddresses, nicNames)
		if err != nil {
			klog.Errorf("unable to determine default router listening addresses: %v", err)
			continue
		}
		if err := updateFunc(ips); err != nil {
			klog.Errorf("unable to update default router service status: %v", err)
			continue
		}
		break
	}

	updateChan := make(chan netlink.AddrUpdate)
	doneChan := make(chan struct{})
	err := netlink.AddrSubscribe(updateChan, doneChan)
	if err != nil {
		klog.Errorf("unable to subscribe to IP address changes: %v", err)
		return
	}

	for {
		select {
		case <-updateChan:
			ips, err := defaultRouterListenAddresses(ipAddresses, nicNames)
			if err != nil {
				klog.Errorf("unable to determine default router listening addresses: %v", err)
				break
			}
			err = updateFunc(ips)
			if err != nil {
				klog.Errorf("unable to update default router service status: %v", err)
				break
			}
		case <-stopCh:
			klog.Info("default router watcher stopping")
			close(doneChan)
			return
		}
	}
}

func isDefaultRouterService(svc *corev1.Service) bool {
	annotationValue, annotationFound := svc.Labels[defaultRouterServiceAnnotationKey]
	return annotationFound &&
		annotationValue == defaultRouterServiceAnnotationValue &&
		svc.Name == defaultRouterServiceName &&
		svc.Namespace == defaultRouterServiceNamespace
}

func defaultRouterListenAddresses(ipAddresses, nicNames []string) ([]string, error) {
	allowedAddresses, err := config.AllowedListeningIPAddresses()
	if err != nil {
		return nil, err
	}

	allowedNicNames, err := config.AllowedNICNames()
	if err != nil {
		return nil, err
	}

	ipList := make([]string, 0, len(ipAddresses)+len(nicNames)*2)

	for _, ip := range ipAddresses {
		if !slices.Contains(allowedAddresses, ip) {
			klog.Warningf("IP address %v not found in the host. Skipping", ip)
			continue
		}
		ipList = append(ipList, ip)
	}

	for _, nicName := range nicNames {
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

	slices.Sort(ipList)
	ipList = slices.Compact(ipList)

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

	ipList := make([]string, 0, len(addrList))

	for _, addr := range addrList {
		ipList = append(ipList, addr.IP.String())
	}

	return ipList, nil
}
