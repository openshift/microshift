#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    # Increased disk size because of the additional embedded images (especially OVMS which is ~3.5GiB)
    LVM_SYSROOT_SIZE=20480 prepare_kickstart host1 kickstart-bootc-offline.ks.template rhel95-bootc-source-ai-model-serving
    launch_vm --boot_blueprint rhel95-bootc-source-ai-model-serving --no_network --vm_disksize 30
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local -r full_guest_name=$(full_vm_name host1)
    run_tests host1 \
        --variable "GUEST_NAME:${full_guest_name}" \
        suites/ai-model-serving/ai-model-serving.robot
}
