package main

import (
	"reflect"
	"sort"
	"strings"
	"testing"

	et "github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
	"github.com/openshift-eng/openshift-tests-extension/pkg/flags"
)

const (
	loadBalancerDifferentNodesTestName = "[sig-network] LoadBalancers [Feature:LoadBalancer] should be able to preserve UDP traffic when server pod cycles for a LoadBalancer service on different nodes"
	loadBalancerSameNodesTestName      = "[sig-network] LoadBalancers [Feature:LoadBalancer] should be able to preserve UDP traffic when server pod cycles for a LoadBalancer service on the same nodes"
)

var singleReplicaExcludedTestNames = []string{
	"[sig-apps] Daemon set should rollback without unnecessary restarts [Conformance] [Serial]",
	"[sig-node] NoExecuteTaintManager Single Pod doesn't evict pod with tolerations from tainted nodes [Serial]",
	"[sig-node] NoExecuteTaintManager Single Pod eventually evict pod with finite tolerations from tainted nodes [Serial]",
	"[sig-node] NoExecuteTaintManager Single Pod evicts pods from tainted nodes [Serial]",
	"[sig-node] NoExecuteTaintManager Single Pod removing taint cancels eviction [Conformance] [Disruptive] [Serial]",
	"[sig-node] NoExecuteTaintManager Single Pod pods evicted from tainted nodes have pod disruption condition [Serial]",
	"[sig-node] NoExecuteTaintManager Multiple Pods evicts pods with minTolerationSeconds [Conformance] [Disruptive] [Serial]",
	"[sig-node] NoExecuteTaintManager Multiple Pods only evicts pods without tolerations from tainted nodes [Serial]",
	"[sig-cli] Kubectl client Kubectl taint should remove all the taints with the same key off a node [Serial]",
	loadBalancerDifferentNodesTestName,
	loadBalancerSameNodesTestName,
	"[sig-architecture] Conformance Tests should have at least two untainted nodes [Conformance]",
}

func TestSingleReplicaTopologyExclusions(t *testing.T) {
	specs, err := buildKubeTestSpecs()
	if err != nil {
		t.Fatalf("build registered Kubernetes test specs: %v", err)
	}

	allSpecs := specsByName(specs)
	for _, name := range singleReplicaExcludedTestNames {
		spec, ok := allSpecs[name]
		if !ok {
			t.Errorf("expected registered Kubernetes test %q", name)
			continue
		}
		if !spec.Labels.Has("[Skipped:SingleReplica]") {
			t.Errorf("expected %q to have the SingleReplica skip label", name)
		}
		if !strings.Contains(spec.EnvironmentSelector.Exclude, et.TopologyEquals("SingleReplica")) {
			t.Errorf("expected %q to have the SingleReplica exclusion, got %q", name, spec.EnvironmentSelector.Exclude)
		}
	}

	highlyAvailable, err := specs.FilterByEnvironment(flags.EnvironmentalFlags{Topology: "HighlyAvailable"})
	if err != nil {
		t.Fatalf("filter HighlyAvailable tests: %v", err)
	}
	singleReplica, err := specs.FilterByEnvironment(flags.EnvironmentalFlags{Topology: "SingleReplica"})
	if err != nil {
		t.Fatalf("filter SingleReplica tests: %v", err)
	}

	highlyAvailableSpecs := specsByName(highlyAvailable)
	singleReplicaSpecs := specsByName(singleReplica)
	for _, name := range singleReplicaExcludedTestNames {
		if _, ok := highlyAvailableSpecs[name]; !ok {
			t.Errorf("expected %q in the HighlyAvailable test list", name)
		}
		if _, ok := singleReplicaSpecs[name]; ok {
			t.Errorf("did not expect %q in the SingleReplica test list", name)
		}
	}

	actualDifference := difference(highlyAvailableSpecs, singleReplicaSpecs)
	expectedDifference := append([]string(nil), singleReplicaExcludedTestNames...)
	sort.Strings(expectedDifference)
	if !reflect.DeepEqual(actualDifference, expectedDifference) {
		t.Errorf("unexpected HighlyAvailable-minus-SingleReplica tests (-want +got):\nwant: %q\n got: %q", expectedDifference, actualDifference)
	}

	highlyAvailableAWS, err := specs.FilterByEnvironment(flags.EnvironmentalFlags{
		Platform: "aws",
		Topology: "HighlyAvailable",
	})
	if err != nil {
		t.Fatalf("filter HighlyAvailable AWS tests: %v", err)
	}
	highlyAvailableAWSSpecs := specsByName(highlyAvailableAWS)
	for _, name := range []string{loadBalancerDifferentNodesTestName, loadBalancerSameNodesTestName} {
		if _, ok := highlyAvailableAWSSpecs[name]; ok {
			t.Errorf("did not expect platform-excluded LoadBalancer test %q in the HighlyAvailable AWS test list", name)
		}
	}
}

func specsByName(specs et.ExtensionTestSpecs) map[string]*et.ExtensionTestSpec {
	byName := make(map[string]*et.ExtensionTestSpec, len(specs))
	for _, spec := range specs {
		byName[spec.Name] = spec
	}
	return byName
}

func difference(left, right map[string]*et.ExtensionTestSpec) []string {
	var names []string
	for name := range left {
		if _, ok := right[name]; !ok {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}
