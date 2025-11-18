#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# NOTE: Unlike most suites, these tests rely on being run IN ORDER to
# ensure MicroShift is upgraded before running standard suite tests
export TEST_RANDOMIZATION=none

start_image="rhel96-bootc-brew-y1-with-optional"
dest_image="rhel96-bootc-brew-${LATEST_RELEASE_TYPE}-with-optional"

scenario_create_vms() {
    exit_if_image_not_found "${start_image}"
    exit_if_image_not_found "${dest_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm --boot_blueprint rhel96-bootc
}

scenario_remove_vms() {
    exit_if_image_not_found "${start_image}"
    exit_if_image_not_found "${dest_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_image_not_found "${start_image}"
    exit_if_image_not_found "${dest_image}"

    run_tests host1 \
        --variable "TARGET_REF:${dest_image}" \
        --variable "BOOTC_REGISTRY:${MIRROR_REGISTRY_URL}" \
        suites/upgrade/upgrade-successful.robot \
        suites/standard2
}
