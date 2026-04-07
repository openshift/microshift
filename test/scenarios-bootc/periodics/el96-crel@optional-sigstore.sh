#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Enable container signature verification for current release images,
# including the optional components.
# These are ec / rc / z-stream, thus must all to be signed.
# shellcheck disable=SC2034  # used elsewhere
IMAGE_SIGSTORE_ENABLED=true

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

    # Run a minimal test for this scenario as its main functionality is
    # to verify container image signature check is enabled
	run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9.6" \
        --variable "IMAGE_SIGSTORE_ENABLED:True" \
        suites/standard1/containers-policy.robot
}
