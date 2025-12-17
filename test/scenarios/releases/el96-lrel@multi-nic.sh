#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel-9.6-microshift-brew-optionals-4.${MINOR_VERSION}-${LATEST_RELEASE_TYPE}"

scenario_create_vms() {
    exit_if_commit_not_found "${start_image}"

    prepare_kickstart host1 kickstart.ks.template "${start_image}"
    # Using multus as secondary network to have 2 nics in different networks.
    launch_vm --network default,"${VM_MULTUS_NETWORK}"
}

scenario_remove_vms() {
    exit_if_commit_not_found "${start_image}"

    remove_vm host1
}

scenario_run_tests() {
    exit_if_commit_not_found "${start_image}"

    local -r vmname=$(full_vm_name host1)
    local -r vm_ip1=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | head -1)
    local -r vm_ip2=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | tail -1)

    run_tests host1 \
        --variable "USHIFT_HOST_IP1:${vm_ip1}" \
        --variable "USHIFT_HOST_IP2:${vm_ip2}" \
        suites/network/multi-nic.robot
}
