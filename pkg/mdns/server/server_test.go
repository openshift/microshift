package server

import (
	"net"
	"strings"
	"testing"
)

func TestFunctionalServer(t *testing.T) {
	var stopCh = make(chan struct{})
	defer close(stopCh)
	r := NewResolver()
	populateResolverForTests(r)

	loopbackInterface, _ := net.InterfaceByName("lo")
	_, err := New(loopbackInterface, r, stopCh)
	if err != nil {
		t.Errorf("Error starting mDNS server on loopback: %q", err)
		return
	}

	for fullHost, ips := range r.domain {
		host := strings.TrimRight(fullHost, ".")
		addrs, err := net.LookupHost(host)
		if err != nil {
			t.Errorf("Error resolving mDNS host: %q, %s", host, err)
		}

		if countIPMatches(ips, addrs) != len(ips) {
			t.Errorf("Not all ips %+v for %q found in resolution: %+v", ips, fullHost, addrs)
		}
	}
}

func countIPMatches(ips []net.IP, addrs []string) int {
	found := 0
	for _, ip := range ips {
		for _, addr := range addrs {
			// resolved IPv6 come in the IP%interface format, like 2001:db8::dead:beef%lo
			addrParts := strings.Split(addr, "%")
			ip2 := net.ParseIP(addrParts[0])
			if ip.Equal(ip2) {
				found += 1
				break
			}
		}
	}
	return found
}
