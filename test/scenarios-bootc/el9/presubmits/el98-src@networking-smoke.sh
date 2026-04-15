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
    run_tests host1 \
        suites/standard1/networking-smoke.robot
}

