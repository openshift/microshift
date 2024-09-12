package mdns

import (
	"context"
	"fmt"
	"net"
	"slices"
	"strings"
	"sync"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/config/ovn"
	"github.com/openshift/microshift/pkg/mdns/server"
	"k8s.io/klog/v2"
)

type MicroShiftmDNSController struct {
	sync.Mutex
	NodeName   string
	NodeIP     string
	KubeConfig string
	isIpv4     bool
	isIpv6     bool
	myIPs      []string
	resolver   *server.Resolver
	hostCount  map[string]int
	stopCh     chan struct{}
}

func NewMicroShiftmDNSController(cfg *config.Config) *MicroShiftmDNSController {
	return &MicroShiftmDNSController{
		NodeIP:     cfg.Node.NodeIP,
		NodeName:   cfg.Node.HostnameOverride,
		KubeConfig: cfg.KubeConfigPath(config.KubeAdmin),
		isIpv4:     cfg.IsIPv4(),
		isIpv6:     cfg.IsIPv6(),
		hostCount:  make(map[string]int),
	}
}

func (c *MicroShiftmDNSController) Name() string { return "microshift-mdns-controller" }
func (c *MicroShiftmDNSController) Dependencies() []string {
	return []string{"openshift-default-scc-manager"}
}

func (c *MicroShiftmDNSController) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	c.stopCh = make(chan struct{})
	defer close(c.stopCh)

	c.resolver = server.NewResolver()

	ifs, _ := net.Interfaces()

	// NOTE: this will listen the physical interface of the host.
	for n := range ifs {
		name := ifs[n].Name
		if ovn.IsOVNKubernetesInternalInterface(name) || name == ovn.OVNGatewayInterface {
			continue
		}
		klog.Infof("mDNS: Starting server on interface %q, NodeIP %q, NodeName %q", name, c.NodeIP, c.NodeName)
		if _, err := server.New(&ifs[n], c.resolver, c.stopCh); err != nil {
			return fmt.Errorf("failed to start server: %w", err)
		}
	}

	ips := []string{c.NodeIP}

	// Discover additional IPs for the interface
	for n := range ifs {
		addrs, _ := ifs[n].Addrs()
		if ipInAddrs(c.NodeIP, addrs) {
			addrs = ovn.ExcludeOVNKubernetesMasqueradeIPs(addrs)
			addrs = slices.DeleteFunc(
				addrs,
				func(addr net.Addr) bool {
					ipAddr, _, _ := net.ParseCIDR(addr.String())
					return ipAddr.IsLinkLocalMulticast() || ipAddr.IsLinkLocalUnicast()
				},
			)
			addrs = slices.DeleteFunc(
				addrs,
				func(addr net.Addr) bool {
					ipAddr, _, _ := net.ParseCIDR(addr.String())
					// This function deletes on a true return.
					// If the IP family matches what MicroShift has been configured, we need
					// to keep the IP, hence the false return.
					if ipAddr.To4() == nil {
						return !c.isIpv6
					}
					return !c.isIpv4
				},
			)
			ips = addrsToStrings(addrs)
		}
	}

	c.myIPs = ips
	if strings.HasSuffix(c.NodeName, server.DefaultmDNSTLD) {
		klog.Infof("mDNS: Host FQDN %q will be announced via mDNS on IPs %q", c.NodeName, ips)
		c.resolver.AddDomain(c.NodeName+".", ips)
	}

	close(ready)

	go func() {
		if err := c.startRouteInformer(c.stopCh); err != nil {
			klog.Errorf("error running router: %v", err)
		}
	}()

	<-ctx.Done()

	return ctx.Err()
}

func ipInAddrs(ip string, addrs []net.Addr) bool {
	for _, a := range addrs {
		ipAddr, _, _ := net.ParseCIDR(a.String())
		if ipAddr.String() == ip {
			return true
		}
	}
	return false
}

func addrsToStrings(addrs []net.Addr) []string {
	var ipAddrs = make([]string, 0)

	for _, a := range addrs {
		ipAddr, _, _ := net.ParseCIDR(a.String())
		ipAddrs = append(ipAddrs, ipAddr.String())
	}
	return ipAddrs
}
