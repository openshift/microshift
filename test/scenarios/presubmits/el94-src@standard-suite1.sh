#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template rhel-9.4-microshift-source
    launch_vm
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9.4" \
        suites/standard1/ suites/selinux/validate-selinux-policy.robot
}
