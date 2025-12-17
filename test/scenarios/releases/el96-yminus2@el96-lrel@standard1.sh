#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# NOTE: Unlike most suites, these tests rely on being run IN ORDER to
# ensure MicroShift is upgraded before running standard suite tests
export TEST_RANDOMIZATION=none

dest_image="rhel-9.6-microshift-brew-optionals-4.${MINOR_VERSION}-${LATEST_RELEASE_TYPE}"

scenario_create_vms() {
    exit_if_commit_not_found "${dest_image}"

    prepare_kickstart host1 kickstart.ks.template "rhel-9.6-microshift-brew-optionals-4.${YMINUS2_MINOR_VERSION}-zstream"
    launch_vm
}

scenario_remove_vms() {
    exit_if_commit_not_found "${dest_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_commit_not_found "${dest_image}"

    run_tests host1 \
        --variable "TARGET_REF:${dest_image}" \
        --variable "EXPECTED_OS_VERSION:9.6" \
        suites/upgrade/upgrade-successful.robot \
        suites/standard1/
}
