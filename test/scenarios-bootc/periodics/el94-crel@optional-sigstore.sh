#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Enable container signature verification for current release images,
# including the optional components.
# These are ec / rc / z-stream, thus must all to be signed.
# shellcheck disable=SC2034  # used elsewhere
IMAGE_SIGSTORE_ENABLED=true

start_image=rhel94-bootc-crel-optionals

scenario_create_vms() {
    if ! does_image_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi
    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm --boot_blueprint rhel94-bootc
}

scenario_remove_vms() {
    if ! does_image_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi
	remove_vm host1
}

scenario_run_tests() {
    if ! does_image_exist "${start_image}"; then
        echo "Image '${start_image}' not found - skipping test"
        return 0
    fi
    # Run a minimal test for this scenario as its main functionality is
    # to verify container image signature check is enabled
	run_tests host1 \
        --variable "IMAGE_SIGSTORE_ENABLED:True" \
        suites/standard1/containers-policy.robot
}
