#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Increase greenboot timeout for release scenarios (optional operators need more time)
# shellcheck disable=SC2034  # used elsewhere
export GREENBOOT_TIMEOUT=1200

# NOTE: Unlike most suites, these tests rely on being run IN ORDER to
# ensure MicroShift is upgraded before running standard suite tests
export TEST_RANDOMIZATION=none

start_image="rhel-9.6-microshift-brew-optionals-${YMINUS2_MAJOR_VERSION}.${YMINUS2_MINOR_VERSION}-zstream"
dest_image="rhel98-brew-lrel-optional"

scenario_create_vms() {
    exit_if_commit_not_found "${start_image}"
    exit_if_commit_not_found "${dest_image}"

    prepare_kickstart host1 kickstart.ks.template "${start_image}"
    launch_vm rhel-9.6 --vm_vcpus 4
}

scenario_remove_vms() {
    exit_if_commit_not_found "${start_image}"
    exit_if_commit_not_found "${dest_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_commit_not_found "${start_image}"
    exit_if_commit_not_found "${dest_image}"

    run_tests host1 \
        --variable "TARGET_REF:${dest_image}" \
        --variable "EXPECTED_OS_VERSION:9.8" \
        suites/upgrade/upgrade-successful.robot \
        suites/standard1/
}
