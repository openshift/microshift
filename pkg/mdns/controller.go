package mdns

import (
	"context"
	"net"
	"path/filepath"
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
		KubeConfig: filepath.Join(cfg.DataDir, "resources", "kubeadmin", "kubeconfig"),
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
		klog.Infof("starting mDNS server", "interface", name, "NodeIP", c.NodeIP, "Node", c.NodeName)
		server.New(&ifs[n], c.resolver, c.stopCh)
	}

	if strings.HasSuffix(c.NodeName, server.DefaultmDNSTLD) {
		ips := []string{c.NodeIP}

		// Discover additional IPs for the interface (IPv6 LLA ...)
		for n := range ifs {
			addrs, _ := ifs[n].Addrs()
			if ipInAddrs(c.NodeIP, addrs) {
				ips = addrsToStrings(addrs)
			}
		}

		klog.Infof("Host FQDN will be announced via mDNS", "fqdn", c.NodeName, "ips", ips)
		c.resolver.AddDomain(c.NodeName+".", ips)
		c.myIPs = ips
	}

	close(ready)

	go c.startRouteInformer(c.stopCh)

	select {
	case <-ctx.Done():

		return ctx.Err()
	}

	return nil
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
