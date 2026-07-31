#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.
#
# Sets up 3 MicroShift clusters with C2CC and Libreswan IPsec (tunnel mode
# protecting pod/service CIDRs).  Each host maintains IPsec tunnels forming a
# full mesh.  Tests validate ESP encapsulation, connectivity, policy
# enforcement, plaintext rejection, and MTU behaviour.

# shellcheck source=test/bin/c2cc_common.sh
source "${SCRIPTDIR}/c2cc_common.sh"

# IPsec tests have ordering dependencies (setup verification must pass before
# enforcement tests), so disable randomization.

scenario_create_vms() {
    c2cc_create_vms "rhel102-bootc-source-ipsec" "rhel102-bootc"
}

scenario_remove_vms() {
    c2cc_remove_vms
}

scenario_run_tests() {
    configure_c2cc_hosts "c2cc_ipsec_pre_greenboot" "c2cc_ipsec_greenboot"
    configure_ipsec

    c2cc_run_tests "suites/c2cc/extra/ipsec.robot"
}
