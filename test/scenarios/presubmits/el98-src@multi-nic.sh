#!/bin/bash

# Sourced from scenario.sh and uses functions defined there.

start_image="rhel-9.8-microshift-source"
SCENARIO_NETWORKS="default,${VM_MULTUS_NETWORK}"

# Opt-in to dynamic VM scheduling by declaring requirements
dynamic_schedule_requirements() {
    cat <<EOF
boot_image=${start_image}
slow=true
EOF
}

scenario_create_vms() {
    prepare_kickstart host1 kickstart.ks.template "${start_image}"
    # Using multus as secondary network to have 2 nics in different networks.
    launch_vm rhel-9.8
}

scenario_remove_vms() {
    remove_vm host1
}

scenario_run_tests() {
    local -r vmname=$(full_vm_name host1)
    local -r vm_ip1=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | head -1)
    local -r vm_ip2=$("${ROOTDIR}/scripts/devenv-builder/manage-vm.sh" ip -n "${vmname}" | tail -1)

    run_tests host1 \
        --variable "USHIFT_HOST_IP1:${vm_ip1}" \
        --variable "USHIFT_HOST_IP2:${vm_ip2}" \
        suites/network/multi-nic.robot
}
