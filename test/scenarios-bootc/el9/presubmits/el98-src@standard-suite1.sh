#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel98-bootc-source
    launch_vm --boot_blueprint rhel98-bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    # The SYNC_FREQUENCY is set to a shorter-than-default value to speed up
    # pre-submit scenario completion time in DNS tests.
    run_tests host1 \
        --variable "EXPECTED_OS_VERSION:9.8" \
        --variable "SYNC_FREQUENCY:5s" \
        suites/standard1/
}
