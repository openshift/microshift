#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc-offline.ks.template rhel94-bootc-source-isolated
    launch_vm --boot_blueprint rhel94-bootc-source-isolated --vm_nics 0 --bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local -r full_guest_name=$(full_vm_name host1)
    run_tests host1 \
        --variable "GUEST_NAME:${full_guest_name}" \
        suites/network/offline.robot
}
