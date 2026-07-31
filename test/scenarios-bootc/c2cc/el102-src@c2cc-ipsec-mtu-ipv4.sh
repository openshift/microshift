#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.
#
# Sets up 3 MicroShift clusters with C2CC and Libreswan IPsec on a
# jumbo-frame network (MTU 9000). Tests validate MTU boundary behavior
# through IPsec tunnel-mode ESP encapsulation and the recommended
# pod MTU reduction to 8900.

# shellcheck source=test/bin/c2cc_common.sh
source "${SCRIPTDIR}/c2cc_common.sh"


scenario_create_vms() {
    c2cc_create_vms "rhel102-bootc-source-ipsec" "rhel102-bootc" "jumbo" "ipv4" "9000"
}

scenario_remove_vms() {
    c2cc_remove_vms
}

scenario_run_tests() {
    configure_c2cc_hosts "c2cc_ipsec_mtu_pre_greenboot" "c2cc_ipsec_mtu_greenboot"
    configure_ipsec

    c2cc_run_tests "suites/c2cc/extra/mtu.robot" "" "" "--variable IPSEC:true"
}
