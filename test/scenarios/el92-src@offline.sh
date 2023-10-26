#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.
# shellcheck disable=SC2034  # used elsewhere
scenario_create_vms() {
    prepare_kickstart host1 kickstart-offline.ks.template rhel-9.2-microshift-source-isolated
    #         vmname boot_blueprint        network_name vm_vcpus vm_memory vm_disksize vm_nicss
    launch_vm host1  rhel-9.2-source-isolated-installer        ""           ""       ""        ""          0
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local full_vmname
    full_vmname=$(full_vm_name host1)
    run_tests host1 --variable="VM_NAME:${full_vmname}" suites/network/offline.robot
}
