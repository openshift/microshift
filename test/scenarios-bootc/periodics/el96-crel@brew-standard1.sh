#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel96-bootc-brew
    launch_vm --boot_blueprint rhel96-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

# Note:
# This is a workaround for the problem described in https://issues.redhat.com/browse/USHIFT-5584.
# Revert to the rhel-9.6 base image when it is in GA.
#
#       --variable "EXPECTED_OS_VERSION:9.6" \
scenario_run_tests() {
    run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9" \
        suites/standard1/ suites/selinux/validate-selinux-policy.robot
}
