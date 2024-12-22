#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc-container.ks.template ""
    launch_container --image rhel95-bootc-source
}

scenario_remove_vms() {
    remove_container
}

scenario_run_tests() {
    local -r tests=$(find suites/standard/ -name \*.robot | grep -v -E 'validate-certificate-rotation.robot|hostname.robot')
    # shellcheck disable=SC2086
    run_tests host1 ${tests}

    # suites/standard/group2/validate-certificate-rotation.robot - gets stuck on `Sleep 5s` in Teardown
    # suites/standard/group3/hostname.robot - cannot change hostname inside a container
}
