#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel94-bootc-source
    launch_vm --boot_blueprint rhel94-bootc --network "${VM_DUAL_STACK_NETWORK}" --bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local -r vmname=$(full_vm_name host1)
    # Valid IP addresses are the first two entries returned by manage-vm script.
    local -r vm_ip1=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | head -1)
    local -r vm_ip2=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | head -2 | tail -1)

    run_tests host1 \
        --variable "USHIFT_HOST_IP1:${vm_ip1}" \
        --variable "USHIFT_HOST_IP2:${vm_ip2}" \
        suites/ipv6/dualstack.robot
}
