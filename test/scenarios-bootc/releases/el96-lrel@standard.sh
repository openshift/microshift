#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

export TEST_EXECUTION_TIMEOUT="60m"

start_image="rhel96-bootc-brew-${LATEST_RELEASE_TYPE}-with-optional"

scenario_create_vms() {
    exit_if_image_not_found "${start_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm --boot_blueprint "${start_image}"
}

scenario_remove_vms() {
    exit_if_image_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_image_not_found "${start_image}"

    run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9.6" \
        suites/standard1/ \
        suites/standard2/ \
        suites/selinux/validate-selinux-policy.robot
}
