package mdns

import (
	"context"
	"net"
	"path/filepath"
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

	for n := range ifs {
		name := ifs[n].Name
		if strings.HasPrefix(name, "flannel") ||
			strings.HasPrefix(name, "veth") ||
			strings.HasPrefix(name, "cni") {
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
		c.resolver.AddDomain(c.NodeName, ips)
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
