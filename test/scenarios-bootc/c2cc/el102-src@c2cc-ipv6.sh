#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# shellcheck source=test/bin/c2cc_common.sh
source "${SCRIPTDIR}/c2cc_common.sh"

c2cc_setup_ipv6 "${VM_IPV6_NETWORK}"

scenario_create_vms() {
    c2cc_create_vms rhel102-bootc-source rhel102-bootc "${VM_IPV6_NETWORK}" ipv6
}

scenario_remove_vms() {
    c2cc_remove_vms
}

scenario_run_tests() {
    # shellcheck disable=SC2119
    configure_c2cc_hosts
    c2cc_run_tests "suites/c2cc/basic/" "2001:db8::/64" ipv6
}
