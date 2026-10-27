#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

export SKIP_GREENBOOT=true

# Certificate renewal adds service restarts to the existing restart-heavy suite.
# Keep the full lifecycle coverage within a scenario-specific timeout.
# shellcheck disable=SC2034  # used by scenario.sh
TEST_EXECUTION_TIMEOUT=45m

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.8-microshift-source
    launch_vm rhel-9.8
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 suites/standard2/
}
