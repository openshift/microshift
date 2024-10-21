#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel94-bootc-source
    launch_vm --boot_blueprint rhel94-bootc --bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    run_tests host1 \
        --variable "FAILING_REF:rhel94-bootc-source-aux" \
        --variable "REASON:prevent_backup" \
        --variable "BOOTC_REGISTRY:${BOOTC_REGISTRY_URL}" \
        suites/upgrade/upgrade-fails-and-rolls-back.robot
}
