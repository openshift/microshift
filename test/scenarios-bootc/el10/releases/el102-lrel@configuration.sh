#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

# Increase greenboot timeout for release scenarios (optional operators need more time)
# shellcheck disable=SC2034  # used elsewhere
export GREENBOOT_TIMEOUT=1200

start_image="rhel102-bootc-brew-lrel-optional"

scenario_create_vms() {
    exit_if_image_not_found "${start_image}"

    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm rhel102-bootc --vm_vcpus 4
}

scenario_remove_vms() {
    exit_if_image_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_image_not_found "${start_image}"

    run_tests host1 suites/configuration/
}
