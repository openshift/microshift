#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# shellcheck source=test/bin/c2cc_common.sh
source "${SCRIPTDIR}/c2cc_common.sh"

# Cluster A (host1): primary IPv6, secondary IPv4
CLUSTER_A_POD_CIDR="fd01::/48"
CLUSTER_A_SVC_CIDR="fd02::/112"
CLUSTER_A_POD_CIDR_DUAL="10.42.0.0/16"
CLUSTER_A_SVC_CIDR_DUAL="10.43.0.0/16"
CLUSTER_A_DOMAIN="cluster-a.remote"

# Cluster B (host2): primary IPv6, secondary IPv4
CLUSTER_B_POD_CIDR="fd04::/48"
CLUSTER_B_SVC_CIDR="fd05::/112"
CLUSTER_B_POD_CIDR_DUAL="10.45.0.0/16"
CLUSTER_B_SVC_CIDR_DUAL="10.46.0.0/16"
CLUSTER_B_DOMAIN="cluster-b.remote"

# Cluster C (host3): primary IPv6, secondary IPv4
CLUSTER_C_POD_CIDR="fd07::/48"
CLUSTER_C_SVC_CIDR="fd08::/112"
CLUSTER_C_POD_CIDR_DUAL="10.48.0.0/16"
CLUSTER_C_SVC_CIDR_DUAL="10.49.0.0/16"
CLUSTER_C_DOMAIN="cluster-c.remote"

scenario_create_vms() {
    c2cc_create_vms rhel98-bootc-source rhel98-bootc "${VM_DUAL_STACK_NETWORK}" dual-stack
}

scenario_remove_vms() {
    c2cc_remove_vms
}

scenario_run_tests() {
    # shellcheck disable=SC2119
    configure_c2cc_hosts
    # IP_FAMILY=ipv6 so IP_CMD defaults to 'ip -6' for primary (IPv6) CIDRs.
    # Dual-stack (IPv4) tests derive ip_cmd from CIDR content via IP Command For CIDR.
    c2cc_run_tests "suites/c2cc/basic/" "2001:db8::/64" ipv6
}
