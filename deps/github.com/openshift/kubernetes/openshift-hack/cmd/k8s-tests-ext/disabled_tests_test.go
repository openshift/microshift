package main

import (
	"testing"

	et "github.com/openshift-eng/openshift-tests-extension/pkg/extension/extensiontests"
	g "github.com/openshift-eng/openshift-tests-extension/pkg/ginkgo"
)

func TestFilterOutDisabledSpecs(t *testing.T) {
	specs, err := g.BuildExtensionTestSpecsFromOpenShiftGinkgoSuite(et.AllTestsIncludingVendored())
	if err != nil {
		t.Fatalf("build registered Kubernetes test specs: %v", err)
	}

	allSpecs := specsByName(specs)
	enabledSpecs := specsByName(filterOutDisabledSpecs(specs))

	disabledTests := []struct {
		reason string
		name   string
	}{
		{
			reason: "SpecialConfig",
			name:   "[sig-node] Container Runtime blackbox test when running a container with a new image should be able to pull from private registry with secret [NodeConformance] [Serial]",
		},
		{
			reason: "Broken",
			name:   "[sig-node] [DRA] [FeatureGate:DRAExtendedResource] [Feature:DynamicResourceAllocation] must run pods with extended resource on dra nodes and device plugin nodes [KubeletMinVersion:1.35] [FeatureGate:DynamicResourceAllocation] [Serial]",
		},
		{
			reason: "RebaseInProgress",
			name:   "[sig-scheduling] SchedulerPreemption validates various priority Pods preempt expectedly with the async preemption [Feature:SchedulerAsyncPreemption] [FeatureGate:SchedulerAsyncPreemption] [Beta] [Serial]",
		},
		{
			reason: "RebaseInProgress",
			name:   "[sig-node] Pod InPlace Resize Container (deferred-resizes) [FeatureGate:InPlacePodVerticalScaling] pod-resize-retry-deferred-test-2 [Serial]",
		},
	}

	for _, test := range disabledTests {
		t.Run(test.reason+"/"+test.name, func(t *testing.T) {
			if _, ok := allSpecs[test.name]; !ok {
				t.Fatalf("expected registered Kubernetes test %q", test.name)
			}
			if _, ok := enabledSpecs[test.name]; ok {
				t.Errorf("did not expect disabled Kubernetes test %q in the enabled test list", test.name)
			}
		})
	}

	const availableTest = "[sig-node] Pod InPlace Resize Container (deferred-resizes) [FeatureGate:InPlacePodVerticalScaling] pod-resize-retry-deferred-test-1 [Serial]"
	if _, ok := allSpecs[availableTest]; !ok {
		t.Fatalf("expected registered Kubernetes test %q", availableTest)
	}
	if _, ok := enabledSpecs[availableTest]; !ok {
		t.Errorf("expected unrelated Kubernetes test %q in the enabled test list", availableTest)
	}
}
