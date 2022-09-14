package mdns

import (
	"context"
	"net"
	"regexp"
	"strings"
	"sync"

	"github.com/openshift/microshift/pkg/config"
	"github.com/openshift/microshift/pkg/mdns/server"
	"k8s.io/klog/v2"
)

type MicroShiftmDNSController struct {
	sync.Mutex
	NodeName   string
	NodeIP     string
	KubeConfig string
	myIPs      []string
	resolver   *server.Resolver
	hostCount  map[string]int
	stopCh     chan struct{}
}

func NewMicroShiftmDNSController(cfg *config.MicroshiftConfig) *MicroShiftmDNSController {
	return &MicroShiftmDNSController{
		NodeIP:     cfg.NodeIP,
		NodeName:   cfg.NodeName,
		KubeConfig: cfg.KubeConfigPath(config.KubeAdmin),
		hostCount:  make(map[string]int),
	}
}

func (s *MicroShiftmDNSController) Name() string { return "microshift-mdns-controller" }
func (s *MicroShiftmDNSController) Dependencies() []string {
	return []string{"openshift-default-scc-manager"}
}

func (c *MicroShiftmDNSController) Run(ctx context.Context, ready chan<- struct{}, stopped chan<- struct{}) error {
	defer close(stopped)

	c.stopCh = make(chan struct{})
	defer close(c.stopCh)

	c.resolver = server.NewResolver()

	ifs, _ := net.Interfaces()

	excludedInterfacesRegexp := regexp.MustCompile(
		"^[A-Fa-f0-9]{15}|" + // OVN pod interfaces
			"ovn.*|" + // OVN ovn-k8s-mp0 and similar interfaces
			"br-int|" + // OVN integration bridge
			"veth.*|cni.*|" + // Interfaces used in bridge-cni or flannel
			"ovs-system$") // Internal OVS interface

	// NOTE: this will listen on both br-ex and the physical interface attached to it
	//       i.e. eth0 . We don't believe it's worth going into the complexities (and coupling)
	//       of talking to OpenvSwitch to discover the physical interface(s) on br-ex. And
	//       we have also verified that no duplicate mDNS answers will happen because of this,
	//       if those were to happend it would be harmless.
	for n := range ifs {
		name := ifs[n].Name
		if excludedInterfacesRegexp.MatchString(name) {
			continue
		}
		klog.Infof("mDNS: Starting server on interface %q, NodeIP %q, NodeName %q", name, c.NodeIP, c.NodeName)
		server.New(&ifs[n], c.resolver, c.stopCh)
	}

	ips := []string{c.NodeIP}

	// Discover additional IPs for the interface (IPv6 LLA ...)
	for n := range ifs {
		addrs, _ := ifs[n].Addrs()
		if ipInAddrs(c.NodeIP, addrs) {
			ips = addrsToStrings(addrs)
		}
	}

	c.myIPs = ips

	if strings.HasSuffix(c.NodeName, server.DefaultmDNSTLD) {

		klog.Infof("mDNS: Host FQDN %q will be announced via mDNS on IPs %q", c.NodeName, ips)
		c.resolver.AddDomain(c.NodeName+".", ips)
	}

	close(ready)

	go c.startRouteInformer(c.stopCh)

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
	var ipAddrs []string

	for _, a := range addrs {
		ipAddr, _, _ := net.ParseCIDR(a.String())
		ipAddrs = append(ipAddrs, ipAddr.String())
	}
	return ipAddrs
}
