#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# NOTE: Unlike most suites, these tests rely on being run IN ORDER to
# ensure MicroShift is upgraded before running standard suite tests
export TEST_RANDOMIZATION=none

dest_image="rhel96-bootc-brew-${LATEST_RELEASE_TYPE}-with-optional"

scenario_create_vms() {
    if ! does_image_exist "${dest_image}"; then
        echo "Image '${dest_image}' not found - skipping test"
        return 0
    fi
    prepare_kickstart host1 kickstart-bootc.ks.template rhel96-bootc-brew-y2-with-optional
    launch_vm --boot_blueprint rhel96-bootc
}

scenario_remove_vms() {
    if ! does_image_exist "${dest_image}"; then
        echo "Image '${dest_image}' not found - skipping test"
        return 0
    fi
    remove_vm host1
}

scenario_run_tests() {
    if ! does_image_exist "${dest_image}"; then
        echo "Image '${dest_image}' not found - skipping test"
        return 0
    fi
    run_tests host1 \
        --variable "TARGET_REF:${dest_image}" \
        --variable "BOOTC_REGISTRY:${MIRROR_REGISTRY_URL}" \
        --variable "EXPECTED_OS_VERSION:9.6" \
        suites/upgrade/upgrade-successful.robot \
        suites/standard1/
}
