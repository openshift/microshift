#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template "rhel-9.4-microshift-brew-optionals-4.${YMINUS2_MINOR_VERSION}-zstream"
    launch_vm 
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "TARGET_REF:rhel-9.6-microshift-brew-optionals-4.${MINOR_VERSION}-${LATEST_RELEASE_VERSION}" \
        --variable "EXPECTED_OS_VERSION:9.6" \
        suites/upgrade/upgrade-successful.robot \
        suites/standard1/ suites/selinux/validate-selinux-policy.robot
}
