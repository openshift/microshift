package mdns

import (
	"testing"

	"github.com/openshift/microshift/pkg/mdns/server"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

const testIP = "1.2.3.4"
const testIPv6 = "1111::cafe:cafe"
const testNodeName = "test-node.local"
const testRouteHost = "test-route-host.cluster.local"
const testRouteHost2 = "test-route-host2.cluster.local"

func newTestController() *MicroShiftmDNSController {
	return &MicroShiftmDNSController{
		NodeIP:    testIP,
		NodeName:  testNodeName,
		resolver:  server.NewResolver(),
		hostCount: make(map[string]int),
		myIPs:     []string{testIP, testIPv6},
	}
}

func Test_addedRoute(t *testing.T) {
	ctl := newTestController()
	route := &unstructured.Unstructured{Object: make(map[string]interface{})}
	unstructured.SetNestedField(route.Object, testRouteHost, "spec", "host")

	ctl.addedRoute(route)
	if !ctl.resolver.HasDomain(testRouteHost + ".") {
		t.Errorf("When a host is added, the mDNS resolver should expose it")
	}
}

func Test_deletedRoute(t *testing.T) {
	ctl := newTestController()
	route := &unstructured.Unstructured{Object: make(map[string]interface{})}

	unstructured.SetNestedField(route.Object, testRouteHost, "spec", "host")
	ctl.addedRoute(route)
	ctl.addedRoute(route)
	ctl.deletedRoute(route)
	if !ctl.resolver.HasDomain(testRouteHost + ".") {
		t.Errorf("When multiple routes share a hostname, deleting one route shouldn't stop exposing the host")
	}

	ctl.deletedRoute(route)
	if ctl.resolver.HasDomain(testRouteHost + ".") {
		t.Errorf("Deleting all routes exposing a hostname should stop exposing the host")
	}
}

func Test_updatedRoute(t *testing.T) {
	ctl := newTestController()
	routeOld := &unstructured.Unstructured{Object: make(map[string]interface{})}
	unstructured.SetNestedField(routeOld.Object, testRouteHost, "spec", "host")

	routeNew := &unstructured.Unstructured{Object: make(map[string]interface{})}
	unstructured.SetNestedField(routeNew.Object, testRouteHost2, "spec", "host")

	ctl.addedRoute(routeOld)
	ctl.updatedRoute(routeOld, routeNew)

	if ctl.resolver.HasDomain(testRouteHost + ".") {
		t.Errorf("Old domain must have updated")
	}

	if !ctl.resolver.HasDomain(testRouteHost2 + ".") {
		t.Errorf("The updated domain must resolve at this point")
	}
}

func Test_updatedRouteDupHost(t *testing.T) {
	ctl := newTestController()
	routeOld := &unstructured.Unstructured{Object: make(map[string]interface{})}
	unstructured.SetNestedField(routeOld.Object, testRouteHost, "spec", "host")

	routeNew := &unstructured.Unstructured{Object: make(map[string]interface{})}
	unstructured.SetNestedField(routeNew.Object, testRouteHost2, "spec", "host")

	ctl.addedRoute(routeOld)
	ctl.addedRoute(routeOld) // two routes with the same hostname
	ctl.updatedRoute(routeOld, routeNew)

	if !ctl.resolver.HasDomain(testRouteHost + ".") {
		t.Errorf("Old domain must have persisted, there is another route using it")
	}

	if !ctl.resolver.HasDomain(testRouteHost2 + ".") {
		t.Errorf("The updated domain must resolve at this point")
	}

	ctl.deletedRoute(routeOld) // deleted the second route we had with the host
	if ctl.resolver.HasDomain(testRouteHost + ".") {
		t.Errorf("Old domain must have be gone after deleting the 2nd route")
	}
}
