#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Disable signature verification because the test performs an upgrade to
# a target reference unsigned image that was generated by local builds
# shellcheck disable=SC2034  # used elsewhere
IMAGE_SIGSTORE_ENABLED=false

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel94-bootc-yminus2
    launch_vm --boot_blueprint rhel94-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "TARGET_REF:rhel95-bootc-source" \
        --variable "BOOTC_REGISTRY:${MIRROR_REGISTRY_URL}" \
        suites/upgrade/upgrade-successful.robot
}
