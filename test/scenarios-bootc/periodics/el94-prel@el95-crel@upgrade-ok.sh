#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

dest_image=rhel95-bootc-crel

scenario_create_vms() {
    if ! does_image_exist "${dest_image}"; then
        echo "Image '${dest_image}' not found - skipping test"
        return 0
    fi
    prepare_kickstart host1 kickstart-bootc.ks.template rhel94-bootc-prel
    launch_vm --boot_blueprint rhel94-bootc
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
        suites/upgrade/upgrade-successful.robot
}
