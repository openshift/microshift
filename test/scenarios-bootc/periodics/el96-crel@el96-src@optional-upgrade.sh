#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image=rhel96-bootc-crel-optionals

scenario_create_vms() {
    exit_if_image_not_found "${start_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm --boot_blueprint rhel96-bootc
}

scenario_remove_vms() {
    exit_if_image_not_found "${start_image}"

	remove_vm host1
}

scenario_run_tests() {
    exit_if_image_not_found "${start_image}"

	run_tests host1 \
		--variable "TARGET_REF:rhel96-bootc-source-optionals" \
        --variable "BOOTC_REGISTRY:${MIRROR_REGISTRY_URL}" \
		suites/upgrade/upgrade-multus.robot
}
