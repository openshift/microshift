#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc-container.ks.template ""
    launch_container --image rhel95-bootc-source
}

scenario_remove_vms() {
    remove_container
}

scenario_run_tests() {
    run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9.5" \
        suites/standard1/
}
