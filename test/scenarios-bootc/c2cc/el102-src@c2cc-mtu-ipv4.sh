#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.
#
# Sets up 3 MicroShift clusters with C2CC on a jumbo-frame network
# (MTU 9000). Tests validate MTU boundary behavior for plain C2CC
# traffic without IPsec.

# shellcheck source=test/bin/c2cc_common.sh
source "${SCRIPTDIR}/c2cc_common.sh"

scenario_create_vms() {
    c2cc_create_vms "rhel102-bootc-source" "rhel102-bootc" "jumbo" "ipv4" "9000"
}

scenario_remove_vms() {
    c2cc_remove_vms
}

scenario_run_tests() {
    configure_c2cc_hosts "c2cc_mtu_pre_greenboot" "c2cc_mtu_greenboot"

    c2cc_run_tests "suites/c2cc/extra/mtu.robot"
}
