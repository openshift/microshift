#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

export SKIP_GREENBOOT=true

# shellcheck source=test/bin/c2cc_common.sh
source "${SCRIPTDIR}/c2cc_common.sh"

scenario_create_vms() {
    c2cc_create_vms_rpm "10.2"
}

scenario_remove_vms() {
    c2cc_remove_vms
}

scenario_run_tests() {
    # shellcheck disable=SC2119
    configure_c2cc_hosts
    c2cc_run_tests "suites/c2cc/basic/" "192.0.2.0/24"
}
