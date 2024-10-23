#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

scenario_create_vms() {
    prepare_kickstart host1 kickstart-bootc.ks.template rhel94-bootc-source
    launch_vm --boot_blueprint rhel94-bootc --vm_nics 2 --bootc
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local -r vmname=$(full_vm_name host1)
    local -r vm_ip1=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | head -1)
    local -r vm_ip2=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | tail -1)
    local -r nic1_name=$(run_command_on_vm host1 nmcli -f name,type connection | awk '$2 == "ethernet" {print $1}' | sort | head -1)
    local -r nic2_name=$(run_command_on_vm host1 nmcli -f name,type connection | awk '$2 == "ethernet" {print $1}' | sort | tail -1)

    run_tests host1 \
        --variable "USHIFT_HOST_IP1:${vm_ip1}" \
        --variable "USHIFT_HOST_IP2:${vm_ip2}" \
        --variable "NIC1_NAME:${nic1_name}" \
        --variable "NIC2_NAME:${nic2_name}" \
        suites/network/multi-nic.robot
}
