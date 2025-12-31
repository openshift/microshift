#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Enable container signature verification for published MicroShift images.
# These are ec / rc / z-stream, thus guaranteed to be signed.
# shellcheck disable=SC2034  # used elsewhere
IMAGE_SIGSTORE_ENABLED=true

scenario_create_vms() {
    if [ -z "${LATEST_RELEASE_IMAGE_URL}" ]; then
        echo "ERROR: Scenario requires a valid LATEST_RELEASE_IMAGE_URL, but got '${LATEST_RELEASE_IMAGE_URL}'"
        record_junit "scenario_create_vms" "build_vm_image_not_found" "FAILED"
        exit 1
    fi
    prepare_kickstart host1 kickstart-bootc.ks.template "${LATEST_RELEASE_IMAGE_URL}"
    launch_vm --boot_blueprint rhel96-bootc

    # Open the firewall ports. Other scenarios get this behavior by embedding
    # settings in the blueprint, but we cannot open firewall ports in published
    # images. We need to do this step before running the RF suite so that suite
    # can assume it can reach all of the same ports as for any other test.
    configure_vm_firewall host1
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "IMAGE_SIGSTORE_ENABLED:True" \
        suites/standard2/
}
