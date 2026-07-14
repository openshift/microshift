#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Each optional suite restarts MicroShift with its own kustomizePaths config,
# adding ~10 minutes of restart overhead to the total execution time.
# shellcheck disable=SC2034  # used elsewhere
TEST_EXECUTION_TIMEOUT=60m

# shellcheck disable=SC2034  # used elsewhere
# Increase greenboot timeout for optional packages (more services to start)
GREENBOOT_TIMEOUT=1200

# Enable container signature verification for current release images,
# including the optional components.
# These are ec / rc / z-stream, thus must all to be signed.
# shellcheck disable=SC2034  # used elsewhere
IMAGE_SIGSTORE_ENABLED=true

start_image="rhel102-bootc-brew-lrel-optional"

scenario_create_vms() {
    exit_if_image_not_found "${start_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm rhel102-bootc --vm_disksize 25 --vm_vcpus 4
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
        --variable "EXPECTED_OS_VERSION:10.2" \
        --variable "IMAGE_SIGSTORE_ENABLED:True" \
        suites/standard1/containers-policy.robot
}
