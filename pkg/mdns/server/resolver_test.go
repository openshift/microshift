package server

import (
	"strings"
	"testing"

	"github.com/miekg/dns"
)

const testDomain = "testdomain.local."
const testIPv4 = "1.2.3.4"
const testIPv6 = "2001:db8::dead:beef"

const testDomain2 = "testdomain2.local."
const testIPv4_2 = "4.5.6.7"
const testIPv6_2 = "2001:db8::cafe:cafe"

func populateResolverForTests(r *Resolver) {
	r.AddDomain(testDomain, []string{testIPv4, testIPv6})
	r.AddDomain(testDomain2, []string{testIPv4_2, testIPv6_2})
}

func TestResolverNoDomains(t *testing.T) {
	r := NewResolver()
	res := r.Answer(dns.Question{Qtype: dns.TypeA, Name: testDomain})
	if len(res) != 0 {
		t.Errorf("With no domains resolver should not respond, but it did: %+v", res)
	}
}

func TestResolver_AddDomain(t *testing.T) {
	r := NewResolver()
	populateResolverForTests(r)

	testResolverDomainTypeAddress(t, r, testDomain, dns.TypeA, testIPv4)
	testResolverDomainTypeAddress(t, r, testDomain, dns.TypeAAAA, testIPv6)
	testResolverDomainTypeAddress(t, r, testDomain2, dns.TypeA, testIPv4_2)
	testResolverDomainTypeAddress(t, r, testDomain2, dns.TypeAAAA, testIPv6_2)
}

func TestResolver_HasDomain(t *testing.T) {
	r := NewResolver()
	populateResolverForTests(r)

	if !r.HasDomain(testDomain) || !r.HasDomain(testDomain2) {
		t.Errorf("HasDomain was false for test domains")
	}
}

func TestResolver_DeleteDomain(t *testing.T) {
	r := NewResolver()
	populateResolverForTests(r)
	r.DeleteDomain(testDomain)

	res := r.Answer(dns.Question{Qtype: dns.TypeA, Name: testDomain})
	if len(res) != 0 {
		t.Errorf("With no domains resolver should not respond, but it did: %+v", res)
	}

}
func testResolverDomainTypeAddress(t *testing.T, r *Resolver, name string, qtype uint16, addr string) {
	res := r.Answer(dns.Question{Qtype: qtype, Name: name})

	if len(res) != 1 {
		t.Errorf("With one domain should respond with len 1, but instead: %+v", res)
		return
	}

	if res[0].Header().Ttl != defaultTTL {
		t.Errorf("Incorrect TTL")
	}

	if res[0].Header().Name != name {
		t.Errorf("Incorrect RR domain name %s", res[0].Header().Name)
	}

	if res[0].Header().Rrtype != qtype {
		t.Errorf("Incorrect RR type, %d", res[0].Header().Rrtype)
	}

	if res[0].Header().Class != dns.ClassINET {
		t.Errorf("Incorrect RR class, %d", res[0].Header().Class)
	}

	if !strings.HasSuffix(res[0].String(), addr) {
		t.Errorf("%s didn't respond with address %s", res[0], addr)
	}
}
