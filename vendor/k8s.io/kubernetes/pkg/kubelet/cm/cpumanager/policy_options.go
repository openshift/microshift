/*
Copyright 2021 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cpumanager

import (
	"fmt"
	"strconv"

	"k8s.io/apimachinery/pkg/util/sets"
	utilfeature "k8s.io/apiserver/pkg/util/feature"
	kubefeatures "k8s.io/kubernetes/pkg/features"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/topology"
	"k8s.io/kubernetes/pkg/kubelet/cm/topologymanager"
)

const (
	FullPCPUsOnlyOption            string = "full-pcpus-only"
	DistributeCPUsAcrossNUMAOption string = "distribute-cpus-across-numa"
	AlignBySocketOption            string = "align-by-socket"
)

var (
	alphaOptions = sets.NewString(
		DistributeCPUsAcrossNUMAOption,
		AlignBySocketOption,
	)
	betaOptions = sets.NewString(
		FullPCPUsOnlyOption,
	)
	stableOptions = sets.NewString()
)

func CheckPolicyOptionAvailable(option string) error {
	if !alphaOptions.Has(option) && !betaOptions.Has(option) && !stableOptions.Has(option) {
		return fmt.Errorf("unknown CPU Manager Policy option: %q", option)
	}

	if alphaOptions.Has(option) && !utilfeature.DefaultFeatureGate.Enabled(kubefeatures.CPUManagerPolicyAlphaOptions) {
		return fmt.Errorf("CPU Manager Policy Alpha-level Options not enabled, but option %q provided", option)
	}

	if betaOptions.Has(option) && !utilfeature.DefaultFeatureGate.Enabled(kubefeatures.CPUManagerPolicyBetaOptions) {
		return fmt.Errorf("CPU Manager Policy Beta-level Options not enabled, but option %q provided", option)
	}

	return nil
}

type StaticPolicyOptions struct {
	// flag to enable extra allocation restrictions to avoid
	// different containers to possibly end up on the same core.
	// we consider "core" and "physical CPU" synonim here, leaning
	// towards the terminoloy k8s hints. We acknowledge this is confusing.
	//
	// looking at https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/,
	// any possible naming scheme will lead to ambiguity to some extent.
	// We picked "pcpu" because it the established docs hints at vCPU already.
	FullPhysicalCPUsOnly bool
	// Flag to evenly distribute CPUs across NUMA nodes in cases where more
	// than one NUMA node is required to satisfy the allocation.
	DistributeCPUsAcrossNUMA bool
	// Flag to ensure CPUs are considered aligned at socket boundary rather than
	// NUMA boundary
	AlignBySocket bool
}

func NewStaticPolicyOptions(policyOptions map[string]string) (StaticPolicyOptions, error) {
	opts := StaticPolicyOptions{}
	for name, value := range policyOptions {
		if err := CheckPolicyOptionAvailable(name); err != nil {
			return opts, err
		}

		switch name {
		case FullPCPUsOnlyOption:
			optValue, err := strconv.ParseBool(value)
			if err != nil {
				return opts, fmt.Errorf("bad value for option %q: %w", name, err)
			}
			opts.FullPhysicalCPUsOnly = optValue
		case DistributeCPUsAcrossNUMAOption:
			optValue, err := strconv.ParseBool(value)
			if err != nil {
				return opts, fmt.Errorf("bad value for option %q: %w", name, err)
			}
			opts.DistributeCPUsAcrossNUMA = optValue
		case AlignBySocketOption:
			optValue, err := strconv.ParseBool(value)
			if err != nil {
				return opts, fmt.Errorf("bad value for option %q: %w", name, err)
			}
			opts.AlignBySocket = optValue
		default:
			// this should never be reached, we already detect unknown options,
			// but we keep it as further safety.
			return opts, fmt.Errorf("unsupported cpumanager option: %q (%s)", name, value)
		}
	}
	return opts, nil
}

func ValidateStaticPolicyOptions(opts StaticPolicyOptions, topology *topology.CPUTopology, topologyManager topologymanager.Store) error {
	if opts.AlignBySocket {
		// Not compatible with topology manager single-numa-node policy option.
		if topologyManager.GetPolicy().Name() == topologymanager.PolicySingleNumaNode {
			return fmt.Errorf("Topolgy manager %s policy is incompatible with CPUManager %s policy option", topologymanager.PolicySingleNumaNode, AlignBySocketOption)
		}
		// Not compatible with topology when number of sockets are more than number of NUMA nodes.
		if topology.NumSockets > topology.NumNUMANodes {
			return fmt.Errorf("Align by socket is not compatible with hardware where number of sockets are more than number of NUMA")
		}
	}
	return nil
}
