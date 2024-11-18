#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc-container.ks.template ""
    launch_container --image rhel94-bootc-source --vg_size 3
}

scenario_remove_vms() {
    remove_container
}

scenario_run_tests() {
    run_tests host1 \
        --variable "LVMD_VG_OVERRIDE:$(full_vm_name host1)" \
        suites/storage/
}
