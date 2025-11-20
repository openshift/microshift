#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

dest_image=rhel96-bootc-crel

scenario_create_vms() {
    exit_if_image_not_found "${dest_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template rhel96-bootc-prel
    launch_vm --boot_blueprint rhel96-bootc
}

scenario_remove_vms() {
    exit_if_image_not_found "${dest_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_image_not_found "${dest_image}"

    run_tests host1 \
        --variable "TARGET_REF:${dest_image}" \
        --variable "BOOTC_REGISTRY:${MIRROR_REGISTRY_URL}" \
        suites/upgrade/upgrade-successful.robot
}
