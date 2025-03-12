#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel96-bootc-source
    launch_vm --boot_blueprint rhel96-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9.5" \
        suites/standard1/
    # When SELinux is working on RHEL 9.6 bootc systems add following suite:
    # suites/selinux/validate-selinux-policy.robot
}
