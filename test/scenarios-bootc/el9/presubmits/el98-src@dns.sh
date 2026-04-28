#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel98-bootc-source"
SCENARIO_VCPUS=4

# Opt-in to dynamic VM scheduling by declaring requirements
dynamic_schedule_requirements() {
    echo "boot_image=${start_image}"
}

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template "${start_image}"
    launch_vm rhel98-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    # The SYNC_FREQUENCY is set to a shorter-than-default value to speed up
    # pre-submit scenario completion time in DNS tests.
    run_tests host1 \
        --variable "SYNC_FREQUENCY:5s" \
        suites/standard1/dns.robot
}
