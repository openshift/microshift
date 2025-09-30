#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# NOTE: Unlike most suites, these tests rely on being run IN ORDER to
# ensure MicroShift is upgraded before running standard suite tests
export TEST_RANDOMIZATION=none

start_image="rhel-9.6-microshift-brew-optionals-4.${PREVIOUS_MINOR_VERSION}-zstream"

scenario_create_vms() {
    if ! does_commit_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi

    prepare_kickstart host1 kickstart.ks.template "${start_image}"
    launch_vm
}

scenario_remove_vms() {
    if ! does_commit_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi

    remove_vm host1
}

scenario_run_tests() {
    if ! does_commit_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi

    run_tests host1 \
        --variable "TARGET_REF:rhel-9.6-microshift-brew-optionals-4.${MINOR_VERSION}-${LATEST_RELEASE_TYPE}" \
        --variable "EXPECTED_OS_VERSION:9.6" \
        suites/upgrade/upgrade-successful.robot \
        suites/standard1/ suites/selinux/validate-selinux-policy.robot
}
