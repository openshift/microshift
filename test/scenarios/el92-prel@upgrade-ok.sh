#!/bin/bash

# Sourced from cleanup_scenario.sh and uses functions defined there.

scenario_create_vms() {
    # Determine the starting image based on the source tree current
    # version, minus one.
    local start_image
    start_image="rhel-9.2-microshift-4.$(previous_minor_version)"

    prepare_kickstart host1 kickstart.ks.template "${start_image}"
    launch_vm host1
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
              --variable "TARGET_REF:rhel-9.2-microshift-source" \
              suites/upgrade/upgrade-successful.robot
}
