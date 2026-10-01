#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# shellcheck source=test/bin/c2cc_common.sh
source "${SCRIPTDIR}/c2cc_common.sh"

# Chaos tests assume a healthy baseline between tests,
# so disable randomization to ensure predictable ordering.
export TEST_RANDOMIZATION=none

scenario_create_vms() {
    c2cc_create_vms rhel98-bootc-source rhel98-bootc
}

scenario_remove_vms() {
    c2cc_remove_vms
}

scenario_run_tests() {
    # shellcheck disable=SC2119
    configure_c2cc_hosts
    c2cc_run_tests "suites/c2cc/disruptive/"
}
