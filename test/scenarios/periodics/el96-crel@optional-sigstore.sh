#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Enable container signature verification for current release images,
# including the optional components.
# These are ec / rc / z-stream, thus must all to be signed.
# shellcheck disable=SC2034  # used elsewhere
IMAGE_SIGSTORE_ENABLED=true

start_commit=rhel-9.6-microshift-crel-optionals

scenario_create_vms() {
    if ! does_commit_exist "${start_commit}"; then
        echo "Commit '${start_commit}' not found in ostree repo - skipping test"
        return 0
    fi
	prepare_kickstart host1 kickstart.ks.template "${start_commit}"
	launch_vm
}

scenario_remove_vms() {
    if ! does_commit_exist "${start_commit}"; then
        echo "Commit '${start_commit}' not found in ostree repo - skipping test"
        return 0
    fi
	remove_vm host1
}

scenario_run_tests() {
    if ! does_commit_exist "${start_commit}"; then
        echo "Commit '${start_commit}' not found in ostree repo - skipping test"
        return 0
    fi
    # Run a minimal test for this scenario as its main functionality is
    # to verify container image signature check is enabled
	run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9.6" \
        --variable "IMAGE_SIGSTORE_ENABLED:True" \
        suites/standard1/containers-policy.robot
}
