#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# shellcheck source=test/bin/c2cc_common.sh
source "${SCRIPTDIR}/c2cc_common.sh"

# Cluster A (host1): primary IPv4, secondary IPv6
CLUSTER_A_POD_CIDR="10.42.0.0/16"
CLUSTER_A_SVC_CIDR="10.43.0.0/16"
CLUSTER_A_POD_CIDR_DUAL="fd01::/48"
CLUSTER_A_SVC_CIDR_DUAL="fd02::/112"
CLUSTER_A_DOMAIN="cluster-a.remote"

# Cluster B (host2): primary IPv4, secondary IPv6
CLUSTER_B_POD_CIDR="10.45.0.0/16"
CLUSTER_B_SVC_CIDR="10.46.0.0/16"
CLUSTER_B_POD_CIDR_DUAL="fd04::/48"
CLUSTER_B_SVC_CIDR_DUAL="fd05::/112"
CLUSTER_B_DOMAIN="cluster-b.remote"

# Cluster C (host3): primary IPv4, secondary IPv6
CLUSTER_C_POD_CIDR="10.48.0.0/16"
CLUSTER_C_SVC_CIDR="10.49.0.0/16"
CLUSTER_C_POD_CIDR_DUAL="fd07::/48"
CLUSTER_C_SVC_CIDR_DUAL="fd08::/112"
CLUSTER_C_DOMAIN="cluster-c.remote"

scenario_create_vms() {
    c2cc_create_vms rhel102-bootc-source rhel102-bootc "${VM_DUAL_STACK_NETWORK}" dual-stack
}

scenario_remove_vms() {
    c2cc_remove_vms
}

scenario_run_tests() {
    # shellcheck disable=SC2119
    configure_c2cc_hosts
    # IP_FAMILY=dual-stack (not 'ipv6'), so IP_CMD defaults to 'ip -4' for primary (IPv4) CIDRs.
    # Dual-stack (IPv6) tests derive ip_cmd from CIDR content via IP Command For CIDR.
    c2cc_run_tests "suites/c2cc/basic/" "192.0.2.0/24" dual-stack
}
