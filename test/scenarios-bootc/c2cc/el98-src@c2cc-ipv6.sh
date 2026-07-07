#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# shellcheck source=test/bin/c2cc_common.sh
source "${SCRIPTDIR}/c2cc_common.sh"

# Redefine network-related settings to use the dedicated IPv6 network bridge
# shellcheck disable=SC2034  # used elsewhere
VM_BRIDGE_IP="$(get_vm_bridge_ip "${VM_IPV6_NETWORK}")"
# shellcheck disable=SC2034  # used elsewhere
WEB_SERVER_URL="http://[${VM_BRIDGE_IP}]:${WEB_SERVER_PORT}"
# Using `hostname` here instead of a raw ip because skopeo only allows either
# ipv4 or fqdn's, but not ipv6. Since the registry is hosted on the ipv6
# network gateway in the host, we need to use a combination of the hostname
# plus /etc/hosts resolution (which is taken care of by kickstart).
# shellcheck disable=SC2034  # used elsewhere
MIRROR_REGISTRY_URL="$(hostname):${MIRROR_REGISTRY_PORT}/microshift"

# Cluster A (host1): non-overlapping CIDRs
CLUSTER_A_POD_CIDR="fd01::/48"
CLUSTER_A_SVC_CIDR="fd02::/112"
CLUSTER_A_DOMAIN="cluster-a.remote"

# Cluster B (host2): non-overlapping CIDRs
CLUSTER_B_POD_CIDR="fd04::/48"
CLUSTER_B_SVC_CIDR="fd05::/112"
CLUSTER_B_DOMAIN="cluster-b.remote"

# Cluster C (host3): non-overlapping CIDRs
CLUSTER_C_POD_CIDR="fd07::/48"
CLUSTER_C_SVC_CIDR="fd08::/112"
CLUSTER_C_DOMAIN="cluster-c.remote"

scenario_create_vms() {
    c2cc_create_vms rhel98-bootc-source rhel98-bootc "${VM_IPV6_NETWORK}" ipv6
}

scenario_remove_vms() {
    c2cc_remove_vms
}

scenario_run_tests() {
    # shellcheck disable=SC2119
    configure_c2cc_hosts
    c2cc_run_tests "suites/c2cc/basic/" "2001:db8::/64" ipv6
}
