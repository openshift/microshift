#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.
#
# Sets up 3 MicroShift clusters with C2CC and Libreswan IPsec on a
# single-stack IPv6 jumbo-frame network (MTU 9000). Same test suite as
# c2cc-ipsec-mtu, run over IPv6 instead of IPv4.

# shellcheck source=test/bin/c2cc_common.sh
source "${SCRIPTDIR}/c2cc_common.sh"


# shellcheck disable=SC2119
c2cc_setup_ipv6

scenario_create_vms() {
    # Read jumbo-ipv6 bridge IP for /etc/hosts resolution in VMs.
    # Keep WEB_SERVER_URL at the default IPv4 — bootc kickstarts don't
    # use REPLACE_OSTREE_SERVER_URL, and the hypervisor-side curl can't
    # reach the IPv6 bridge due to libvirt's nftables filtering.
    # shellcheck disable=SC2034  # used elsewhere
    VM_BRIDGE_IP="$(get_vm_bridge_ip "jumbo-ipv6")"
    c2cc_create_vms "rhel98-bootc-source-ipsec" "rhel98-bootc" "jumbo-ipv6" "ipv6" "9000"
}

scenario_remove_vms() {
    c2cc_remove_vms
}

scenario_run_tests() {
    configure_c2cc_hosts "c2cc_ipsec_mtu_pre_greenboot" "c2cc_ipsec_mtu_greenboot"
    configure_ipsec

    c2cc_run_tests "suites/c2cc/extra/mtu.robot" "" ipv6 "--variable IPSEC:true"
}
