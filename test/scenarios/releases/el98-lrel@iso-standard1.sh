#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel98-brew-lrel-optional"
SCENARIO_VCPUS=4

# Opt-in to dynamic VM scheduling by declaring requirements
dynamic_schedule_requirements() {
    echo "boot_image=${start_image}"
}

scenario_create_vms() {
    exit_if_commit_not_found "${start_image}"

    prepare_kickstart host1 kickstart.ks.template "${start_image}"
    launch_vm "${start_image}"
}

scenario_remove_vms() {
    exit_if_commit_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_commit_not_found "${start_image}"

    run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9.8" \
        suites/standard1/
}
