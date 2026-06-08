#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# shellcheck source=test/bin/c2cc_common.sh
source "${SCRIPTDIR}/c2cc_common.sh"

export TEST_RANDOMIZATION=suites

# Cluster A (host1): default MicroShift CIDRs
CLUSTER_A_POD_CIDR="10.42.0.0/16"
CLUSTER_A_SVC_CIDR="10.43.0.0/16"
CLUSTER_A_DOMAIN="cluster-a.remote"

# Cluster B (host2): non-overlapping CIDRs
CLUSTER_B_POD_CIDR="10.45.0.0/16"
CLUSTER_B_SVC_CIDR="10.46.0.0/16"
CLUSTER_B_DOMAIN="cluster-b.remote"

# Cluster C (host3): non-overlapping CIDRs
CLUSTER_C_POD_CIDR="10.48.0.0/16"
CLUSTER_C_SVC_CIDR="10.49.0.0/16"
CLUSTER_C_DOMAIN="cluster-c.remote"

scenario_create_vms() {
    c2cc_create_vms rhel98-bootc-source rhel98-bootc
}

scenario_remove_vms() {
    c2cc_remove_vms
}

scenario_run_tests() {
    c2cc_run_tests "192.0.2.0/24"
}
