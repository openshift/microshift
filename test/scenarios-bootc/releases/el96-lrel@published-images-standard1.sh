#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Enable container signature verification for published MicroShift images.
# These are ec / rc / zstream, thus guaranteed to be signed.
# shellcheck disable=SC2034  # used elsewhere
IMAGE_SIGSTORE_ENABLED=true

start_image="rhel96-bootc-konflux-lrel"

scenario_create_vms() {
    exit_if_image_not_found "${start_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template "${LATEST_RELEASE_IMAGE_URL}"
    launch_vm --boot_blueprint rhel96-bootc

    # Open the firewall ports. Other scenarios get this behavior by embedding
    # settings in the blueprint, but we cannot open firewall ports in published
    # images. We need to do this step before running the RF suite so that suite
    # can assume it can reach all of the same ports as for any other test.
    configure_vm_firewall host1
}

scenario_remove_vms() {
    exit_if_image_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_image_not_found "${start_image}"

    run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9.6" \
        --variable "IMAGE_SIGSTORE_ENABLED:True" \
        suites/standard1/ suites/selinux/validate-selinux-policy.robot
}
