#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.
#
# Sets up 3 MicroShift clusters with C2CC and Libreswan IPsec (tunnel mode
# protecting pod/service CIDRs) on a single-stack IPv6 network. Same test
# suite as c2cc-ipsec, run over IPv6 instead of IPv4.

# shellcheck source=test/bin/c2cc_common.sh
source "${SCRIPTDIR}/c2cc_common.sh"

# IPsec tests have ordering dependencies (setup verification must pass before
# enforcement tests), so disable randomization.

c2cc_setup_ipv6 "${VM_IPV6_NETWORK}"

scenario_create_vms() {
    c2cc_create_vms "rhel98-bootc-source-ipsec" "rhel98-bootc" "${VM_IPV6_NETWORK}" ipv6
}

scenario_remove_vms() {
    c2cc_remove_vms
}

scenario_run_tests() {
    configure_c2cc_hosts "c2cc_ipsec_pre_greenboot" "c2cc_ipsec_greenboot"
    configure_ipsec

    c2cc_run_tests "suites/c2cc/extra/ipsec.robot" "" ipv6
}
