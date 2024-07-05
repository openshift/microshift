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

func defaultRouterWatch(ipAddresses, nicNames []string, ipv4, ipv6 bool, updateFunc serviceUpdateFunction, stopCh <-chan struct{}) {
	updateChan := make(chan netlink.AddrUpdate)
	doneChan := make(chan struct{})
	for {
		ips, err := defaultRouterListenAddresses(ipAddresses, nicNames, ipv4, ipv6)
		if err != nil {
			klog.ErrorS(err, "unable to determine default router listening addresses")
			continue
		}
		if err := updateFunc(ips); err != nil {
			klog.ErrorS(err, "unable to update default router service status")
			continue
		}
		if err := netlink.AddrSubscribe(updateChan, doneChan); err != nil {
			klog.ErrorS(err, "unable to subscribe to IP address changes")
			continue
		}
		break
	}
	klog.Info("Default router watcher configured, waiting on IP address changes ")
	for {
		select {
		case <-updateChan:
			ips, err := defaultRouterListenAddresses(ipAddresses, nicNames, ipv4, ipv6)
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

func defaultRouterListenAddresses(ipAddresses, nicNames []string, ipv4, ipv6 bool) ([]string, error) {
	allowedAddresses, err := config.AllowedListeningIPAddresses(ipv4, ipv6)
	if err != nil {
		return nil, err
	}

	allowedNicNames, err := config.AllowedNICNames()
	if err != nil {
		return nil, err
	}

	if len(ipAddresses) == 0 && len(nicNames) == 0 {
		ipAddresses = allowedAddresses
		nicNames = allowedNicNames
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
		nicAddresses, err := ipAddressesFromNIC(nicName, ipv4, ipv6)
		if err != nil {
			return nil, err
		}
		for _, nicAddress := range nicAddresses {
			if !slices.Contains(allowedAddresses, nicAddress) {
				// NICs may get their addresses changed while MicroShift is running.
				// None of the forbidden IPs (from config pkg) can be configured because
				// MicroShift will refuse to start. If any of those IPs is configured in
				// a NIC after starting MicroShift it should be ignored. As for non
				// forbidden addresses, they are already included in the allowed list, so
				// any new IP will always be there.
				continue
			}
			ipList = append(ipList, nicAddress)
		}
	}

	slices.Sort(ipList)
	ipList = slices.Compact(ipList)

	return ipList, nil
}

func ipAddressesFromNIC(name string, ipv4, ipv6 bool) ([]string, error) {
	link, err := netlink.LinkByName(name)
	if err != nil {
		return nil, err
	}

	family := netlink.FAMILY_V4
	if ipv6 {
		family = netlink.FAMILY_ALL
		if !ipv4 {
			family = netlink.FAMILY_V6
		}
	}
	addrList, err := netlink.AddrList(link, family)
	if err != nil {
		return nil, err
	}

	ipList := make([]string, 0, len(addrList))

	for _, addr := range addrList {
		ipList = append(ipList, addr.IP.String())
	}

	return ipList, nil
}
