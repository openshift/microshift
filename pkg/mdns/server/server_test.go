package server

import (
	"net"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFunctionalServer(t *testing.T) {
	var stopCh = make(chan struct{})
	defer close(stopCh)
	resolver := NewResolver()
	populateResolverForTests(resolver)

	loopbackInterface, _ := net.InterfaceByName("lo")
	_, err := New(loopbackInterface, resolver, stopCh)
	if err != nil {
		t.Errorf("Error starting mDNS server on loopback: %q", err)
		return
	}

	for fullHost, ips := range resolver.domain {
		t.Run(fullHost, func(t *testing.T) {

			host := strings.TrimRight(fullHost, ".")
			addrs, err := net.LookupHost(host)
			if err != nil {
				t.Errorf("Error resolving mDNS host: %q, %s", host, err)
			}
			for _, ip := range ips {

				// Avahi is configured to only return IPv4 addresses by
				// default. Enabling IPv6 may introduce significant delays when
				// looking for a name that does not exist. Therefore this test
				// only checks IPv4 addresses, to keep the test fast and to ensure
				// that the developer host configuration does not need to be set
				// in a degraded state by default.
				// https://unix.stackexchange.com/questions/586334/how-to-enable-mdns-for-ipv6-on-ubuntu-and-debian
				// https://bugzilla.redhat.com/show_bug.cgi?id=821127
				// https://fedoraproject.org/wiki/Tools/Avahi
				if len(ip) == net.IPv6len {
					continue
				}

				t.Run(ip.String(), func(t *testing.T) {
					assert.True(t, ipInList(ip, addrs), "did not find %s in %v", ip, addrs)
				})
			}
		})
	}
}

func ipInList(ip net.IP, addrs []string) bool {
	for _, addr := range addrs {
		// resolved IPv6 come in the IP%interface format, like 2001:db8::dead:beef%lo
		addrParts := strings.Split(addr, "%")
		ip2 := net.ParseIP(addrParts[0])
		if ip.Equal(ip2) {
			return true
		}
	}
	return false
}
